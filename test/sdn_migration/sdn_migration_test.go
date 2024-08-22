package sdn_migration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	configv1 "github.com/openshift/api/config/v1"
	openshiftclient "github.com/openshift/osde2e-common/pkg/clients/openshift"
	prometheusclient "github.com/openshift/osde2e-common/pkg/clients/prometheus"
	"github.com/openshift/osde2e-common/pkg/clouds/aws"
	osdprovider "github.com/openshift/osde2e-common/pkg/openshift/osd"
	rosaprovider "github.com/openshift/osde2e-common/pkg/openshift/rosa"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
- before all create clients
- test cluster creation
- test cluster upgrade
- test cluster migration<
	- test apply manifests
*/

const (
	osdClusterReadyJobName    = "osd-cluster-ready"
	osdClusterReadyJobTimeout = 45 * time.Minute
	version4_15               = "4.15.8"
	version4_16               = "4.16.0-rc.0"
	image4_15                 = "quay.io/openshift-release-dev/ocp-release@sha256:8032c4248d7ae032d5e79debf975d08683cc34d5f08ab2e937ce2d1e940c007b"
	image4_16                 = "quay.io/openshift-release-dev/ocp-release@sha256:c56b01191de4cbb4b97c6eeaf61c5c122fcd465d1d0d671db640d877638ed790"
	channel4_15               = "14.15.8"
	channel4_16               = "4.16.0-rc.0"
	upgradeMaxAttempts        = 1080
	upgradeDelay              = 10
)

type rosaCluster struct {
	id             string
	name           string
	channelGroup   string
	version        string //*semver.Version
	reportDir      string
	upgradeVersion *semver.Version
	kubeconfigFile string

	client *openshiftclient.Client
}

var (
	logger *log.Logger
)

var _ = Describe("SDN migration", ginkgo.Ordered, func() {
	const clusterName = "rosa-sdn-ovn-1"
	var (
		testRosaCluster    *rosaCluster
		reportDir          = getEnvVar("REPORT_DIR", envconf.RandomName(fmt.Sprintf("%s/sdn_migration", os.TempDir()), 25))
		ocmToken           = os.Getenv("OCM_TOKEN")
		clientID           = os.Getenv("CLIENT_ID")
		clientSecret       = os.Getenv("CLIENT_SECRET")
		ocmEnv             = ocm.Stage
		logger             = GinkgoLogr
		rosaProvider       *rosaprovider.Provider
		createRosaCluster  = Label("CreateRosaCluster")
		removeRosaCluster  = Label("RemoveRosaCluster")
		postMigrationCheck = Label("PostMigrationCheck")
		rosaUpgrade        = Label("RosaUpgrade")
		postUpgradeCheck   = Label("PostUpgradeCheck")
		sdnToOvn           = Label("SdnToOvn")
		secretAccessKey    = os.Getenv("AWS_SECRET_ACCESS_KEY")
		accessKeyID        = os.Getenv("AWS_ACCESS_KEY_ID")
	)

	_ = BeforeAll(func(ctx context.Context) {
		var err error
		testRosaCluster = &rosaCluster{}

		Expect(ocmToken).ShouldNot(BeEmpty(), "ocm token is undefined")

		rosaProvider, err = rosaprovider.New(ctx, ocmToken, clientID, clientSecret, ocmEnv, logger, &aws.AWSCredentials{
			Region:          "us-east-1",
			SecretAccessKey: secretAccessKey,
			AccessKeyID:     accessKeyID,
		})
		Expect(err).ShouldNot(HaveOccurred(), "failed to construct rosa provider")
		osdProvider, err := osdprovider.New(ctx, ocmToken, clientID, clientSecret, ocmEnv, logger)
		Expect(err).ShouldNot(HaveOccurred(), "failed to construct osd provider")
		DeferCleanup(osdProvider.Client.Close)

		if createRosaCluster.MatchesLabelFilter(GinkgoLabelFilter()) && os.Getenv("CLUSTER_ID") == "" {
			testRosaCluster.id, err = rosaProvider.CreateCluster(ctx, &rosaprovider.CreateClusterOptions{
				ClusterName:                  clusterName,
				Version:                      "4.14.14",
				UseDefaultAccountRolesPrefix: true,
				STS:                          true,
				Mode:                         "auto",
				ChannelGroup:                 "stable",
				ComputeMachineType:           "m5.xlarge",
				MinReplicas:                  3,
				MaxReplicas:                  24,
				MultiAZ:                      true,
				EnableAutoscaling:            true,
				ETCDEncryption:               true,
				NetworkType:                  "OpenShiftSDN",
			})
			Expect(err).ShouldNot(HaveOccurred(), "failed to create rosa cluster")
		} else {
			testRosaCluster.id = os.Getenv("CLUSTER_ID")
		}

		rosaCluster, err := osdProvider.ClustersMgmt().V1().Clusters().Cluster(testRosaCluster.id).Get().SendContext(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		testRosaCluster.name = rosaCluster.Body().Name()
		testRosaCluster.version = rosaCluster.Body().Version().RawID()
		testRosaCluster.channelGroup = rosaCluster.Body().Version().ChannelGroup()

		testRosaCluster.kubeconfigFile, err = rosaProvider.KubeconfigFile(ctx, testRosaCluster.id, os.TempDir())
		Expect(err).ShouldNot(HaveOccurred())

		testRosaCluster.client, err = openshiftclient.NewFromKubeconfig(testRosaCluster.kubeconfigFile, logger)
		Expect(err).ShouldNot(HaveOccurred(), "failed to construct service cluster client")

		testRosaCluster.reportDir = fmt.Sprintf("%s/%s", reportDir, testRosaCluster.name)
		Expect(os.MkdirAll(reportDir, os.ModePerm)).ShouldNot(HaveOccurred(), "failed to create report directory")
		Expect(os.MkdirAll(testRosaCluster.reportDir, os.ModePerm)).ShouldNot(HaveOccurred(), "failed to create rosa cluster report directory")

	})

	AfterAll(func(ctx context.Context) {
		if removeRosaCluster.MatchesLabelFilter(GinkgoLabelFilter()) {
			rosaProvider, err := rosaprovider.New(ctx, ocmToken, clientID, clientSecret, ocmEnv, logger, &aws.AWSCredentials{
				Profile:         "",
				Region:          "us-east-1",
				SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
				AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			})
			err = rosaProvider.DeleteCluster(ctx, &rosaprovider.DeleteClusterOptions{
				ClusterName:        testRosaCluster.name,
				WorkingDir:         testRosaCluster.reportDir,
				STS:                true,
				DeleteOidcConfigID: true,
			})
			Expect(err).Should(BeNil(), "failed to delete rosa cluster")
		}
	})

	It("rosa cluster is upgraded to 4.15.8 successfully", rosaUpgrade, func(ctx context.Context) {
		err := patchVersionConfig(ctx, testRosaCluster.client, channel4_15, image4_15, version4_15)
		Expect(err).ShouldNot(HaveOccurred(), "rosa cluster upgrade failed")
		err = checkUpgradeStatus(ctx, testRosaCluster.client, version4_15)
		Expect(err).ShouldNot(HaveOccurred(), err)

	})

	It("rosa cluster is healthy post 4.15.8 upgrade", postUpgradeCheck, func(ctx context.Context) {
		criticalAlerts, _, err := queryPrometheusAlerts(ctx, testRosaCluster.client, fmt.Sprintf("%s/4.15.8-prometheus-alerts-pre-upgrade.log", testRosaCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing pre upgrade")

		err = osdClusterReadyHealthCheck(ctx, testRosaCluster.client, "post-upgrade", testRosaCluster.reportDir)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")

	})

	It("rosa cluster is upgraded to 4.16.0-rc.0 successfully", rosaUpgrade, func(ctx context.Context) {
		err := patchVersionConfig(ctx, testRosaCluster.client, channel4_16, image4_16, version4_16)
		Expect(err).ShouldNot(HaveOccurred(), "rosa cluster upgrade failed")
		err = checkUpgradeStatus(ctx, testRosaCluster.client, version4_16)
		Expect(err).ShouldNot(HaveOccurred(), err)

	})

	It("rosa cluster is healthy post 4.16.0-rc.0 upgrade", postUpgradeCheck, func(ctx context.Context) {
		criticalAlerts, _, err := queryPrometheusAlerts(ctx, testRosaCluster.client, fmt.Sprintf("%s/4.16.0-rc.0-prometheus-alerts-pre-upgrade.log", testRosaCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing pre upgrade")
		err = osdClusterReadyHealthCheck(ctx, testRosaCluster.client, "post-upgrade", testRosaCluster.reportDir)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")

	})

	It("rosa cluster migrated from sdn to ovn successfully", sdnToOvn, func(ctx context.Context) {
		err := patchNetworkConfigv1(ctx, testRosaCluster.client)
		Expect(err).ShouldNot(HaveOccurred(), "Rosa Cluster failed to patch network")
		err = patchNetworkConfig(ctx, testRosaCluster.client)
		Expect(err).ShouldNot(HaveOccurred(), "Rosa Cluster failed to patch network")
		err = checkMigrationStatus(ctx, testRosaCluster.client)
		Expect(err).ShouldNot(HaveOccurred(), "Rosa Cluster failed to patch network")

	})
	It("rosa cluster has no critical alerts firing post sdn to ovn migration", postMigrationCheck, func(ctx context.Context) {
		criticalAlerts, _, err := queryPrometheusAlerts(ctx, testRosaCluster.client, fmt.Sprintf("%s/prometheus-alerts-pre-upgrade.log", testRosaCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing pre upgrade")
		err = osdClusterReadyHealthCheck(ctx, testRosaCluster.client, "post-upgrade", testRosaCluster.reportDir)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")

	})
})

// queryPrometheusAlerts queries prometheus for alerts and provides a count for critical and warning alerts
func queryPrometheusAlerts(ctx context.Context, client *openshiftclient.Client, logFilename string) (int, int, error) {
	criticalAlertCount, warningAlertCount := 0, 0
	alerts := ""

	type metric struct {
		AlertName  string `json:"alertname"`
		AlertState string `json:"alertstate"`
		Condition  string `json:"condition"`
		Endpoint   string `json:"endpoint"`
		Name       string `json:"name"`
		Namespace  string `json:"namespace"`
		Severity   string `json:"severity"`
	}

	prometheusClient, _ := prometheusclient.New(ctx, client)
	vector, err := prometheusClient.InstantQuery(ctx, "ALERTS{alertstate!=\"pending\",alertname!=\"Watchdog\"}")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query prometheus: %v", err)
	}

	for _, model := range vector {
		metric := metric{}

		metricEncoded, err := json.Marshal(model.Metric)
		if err != nil {
			continue
		}

		err = json.Unmarshal(metricEncoded, &metric)
		if err != nil {
			continue
		}

		alerts += fmt.Sprintf("Since: %s : %+v\n", time.Unix(model.Timestamp.Unix(), 0), metric)

		switch model.Metric["severity"] {
		case "critical":
			criticalAlertCount += 1
		case "warning":
			warningAlertCount += 1
		}
	}

	if alerts != "" {
		if err = os.WriteFile(logFilename, []byte(alerts), os.FileMode(0o644)); err != nil {
			return criticalAlertCount, warningAlertCount, fmt.Errorf("failed to write prometheus alerts to file: %v", err)
		}
	}

	return criticalAlertCount, warningAlertCount, nil
}

// getEnvVar returns the env variable value and if unset returns default provided
func getEnvVar(key, value string) string {
	result, exist := os.LookupEnv(key)
	if exist {
		return result
	}
	return value
}

// osdClusterReadyHealthCheck verifies the osd-cluster-ready health check job is passing
func osdClusterReadyHealthCheck(ctx context.Context, clusterClient *openshiftclient.Client, action, reportDir string) error {
	var (
		err error
		job batchv1.Job
	)

	if err = clusterClient.Get(ctx, osdClusterReadyJobName, "openshift-monitoring", &job); err != nil {
		return fmt.Errorf("failed to get existing %s job %v", osdClusterReadyJobName, err)
	}

	newJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: osdClusterReadyJobName,
			Annotations:  job.Annotations,
			Namespace:    job.Namespace,
		},
		Spec: job.Spec,
	}

	newJob.Spec.Selector.MatchLabels = map[string]string{}
	newJob.Spec.Template.ObjectMeta.Name = newJob.GetGenerateName()
	newJob.Spec.Template.ObjectMeta.Labels = map[string]string{}
	newJob.Spec.Template.Spec.Containers[0].Name = newJob.GetGenerateName()

	if err = clusterClient.Create(ctx, newJob); err != nil {
		return fmt.Errorf("failed to create %s job: %v", newJob.GetName(), err)
	}

	defer func() {
		_ = clusterClient.Delete(ctx, newJob)
	}()

	return clusterClient.OSDClusterHealthy(ctx, newJob.GetName(), reportDir, osdClusterReadyJobTimeout)
}

// getKubernetesDynamicClient returns the kubernetes dynamic client
func getKubernetesDynamicClient(client *openshiftclient.Client) (*dynamic.DynamicClient, error) {
	dynamicClient, err := dynamic.NewForConfig(client.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes dynamic client: %w", err)
	}
	return dynamicClient, nil
}

type upgradeError struct {
	err error
}

// Error returns the formatted error message when upgradeError is invoked
func (e *upgradeError) Error() string {
	return fmt.Sprintf("osd upgrade failed: %v", e.err)
}

//	func getClusterVersionConfig(ctx context.Context, dynamicClient *dynamic.DynamicClient) (*unstructured.Unstructured, error) {
//		gvr := schema.GroupVersionResource{
//			Group:    "config.openshift.io",
//			Version:  "v1",
//			Resource: "clusterversions",
//		}
//		return dynamicClient.Resource(gvr).Get(ctx, "version", metav1.GetOptions{})
//	}
func getOpenshiftConfig(ctx context.Context, dynamicClient *dynamic.DynamicClient, resource string, name string) (*unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: resource,
	}
	return dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
}

func checkMigrationStatus(ctx context.Context, client *openshiftclient.Client) error {
	var (
		dynamicClient *dynamic.DynamicClient
		err           error
	)

	if dynamicClient, err = getKubernetesDynamicClient(client); err != nil {
		return &upgradeError{err: err}
	}
	for i := 1; i <= upgradeMaxAttempts; i++ {
		// Get the current network configuration
		networkConfig, err := getOpenshiftConfig(ctx, dynamicClient, "networks", "cluster")
		if err != nil {
			return fmt.Errorf("failed to get network configuration: %v", err)
		}

		// Extract the status conditions
		status, found, err := unstructured.NestedMap(networkConfig.Object, "status")
		if err != nil || !found {
			return fmt.Errorf("failed to find status in network configuration: %v", err)
		}

		conditions, found, err := unstructured.NestedSlice(status, "conditions")
		if err != nil || !found {
			return fmt.Errorf("failed to find conditions in status: %v", err)
		}

		// Check for the migration condition
		migrationInProgress := false
		for _, cond := range conditions {
			conditionMap, ok := cond.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid type for condition entry")
			}

			conditionType, found, err := unstructured.NestedString(conditionMap, "type")
			if err != nil || !found {
				return fmt.Errorf("failed to find condition type: %v", err)
			}

			if conditionType == "NetworkTypeMigrationInProgress" {
				status, found, err := unstructured.NestedString(conditionMap, "status")
				if err != nil || !found {
					return fmt.Errorf("failed to find condition status: %v", err)
				}

				reason, found, err := unstructured.NestedString(conditionMap, "reason")
				if err != nil || !found {
					return fmt.Errorf("failed to find condition reason: %v", err)
				}

				// Check if the migration is in progress
				if status == "True" && reason == "NetworkTypeMigrationStarted" {
					migrationInProgress = true
					break
				}

				// Check if the migration is complete
				if status == "False" && reason == "NetworkTypeMigrationCompleted" {
					fmt.Println("Network migration completed successfully!")
					return nil
				}
			}
		}

		if migrationInProgress {
			fmt.Println("Network migration is still in progress...")
		} else {
			fmt.Println("Migration status is unknown or not in progress.")
		}

		time.Sleep(upgradeDelay * time.Second)
	}

	return fmt.Errorf("network migration is still in progress after %d attempts", upgradeMaxAttempts)

}

// checkUpgradeStatus
func checkUpgradeStatus(ctx context.Context, client *openshiftclient.Client, upgradeVersion string) error {
	startTime := time.Now() // Start timing

	var (
		conditionMessage string
		dynamicClient    *dynamic.DynamicClient
		upgradeState     string
		err              error
	)

	if dynamicClient, err = getKubernetesDynamicClient(client); err != nil {
		return &upgradeError{err: err}
	}

	for i := 1; i <= upgradeMaxAttempts; i++ {
		// Get the current cluster version configuration
		upgradeConfig, err := getOpenshiftConfig(ctx, dynamicClient, "clusterversions", "version")
		if err != nil {
			logger.Printf("Failed to get cluster version config: %v", err)
			time.Sleep(upgradeDelay * time.Second)
			continue
		}

		// Extract the status map from the configuration
		status, found, err := unstructured.NestedMap(upgradeConfig.Object, "status")
		if err != nil || !found {
			logger.Printf("Failed to find status in cluster version config: %v", err)
			time.Sleep(upgradeDelay * time.Second)
			continue
		}

		// Extract the history slice from the status
		histories, found, err := unstructured.NestedSlice(status, "history")
		if err != nil || !found {
			log.Printf("Failed to find history in status")
			time.Sleep(upgradeDelay * time.Second)
			continue
		}

		for _, h := range histories {
			historyMap, ok := h.(map[string]interface{})
			if !ok {
				err = fmt.Errorf("invalid type for history entry")
				log.Printf(err.Error(), "Invalid history entry type")
				continue
			}

			// Extract the version for each history entry
			version, found, err := unstructured.NestedString(historyMap, "version")
			if err != nil || !found {
				log.Printf("Failed to find version in history entry")
				continue
			}

			// Check if the version matches the desired upgrade version
			if version == upgradeVersion {
				// Extract the state for the matching version
				state, found, err := unstructured.NestedString(historyMap, "state")
				if err != nil || !found {
					log.Printf("Failed to find state in history entry")
					continue
				}

				upgradeState = state
				break
			}
		}

		// Extract the conditions from the status
		conditions, found, err := unstructured.NestedSlice(status, "conditions")
		if err != nil || !found {
			log.Printf("Failed to find conditions in status")
		} else {
			// Filter for the condition message that starts with "Working towards"
			for _, cond := range conditions {
				conditionMap, ok := cond.(map[string]interface{})
				if !ok {
					err = fmt.Errorf("invalid type for condition entry")
					log.Printf(err.Error(), "Invalid condition entry type")
					continue
				}

				// Extract the condition message
				message, found, err := unstructured.NestedString(conditionMap, "message")
				if err != nil || !found {
					log.Printf("Failed to find message in condition entry")
					continue
				}

				// Check if the message starts with "Working towards"
				if strings.HasPrefix(message, "Working towards") {
					conditionMessage = message
					break
				}
			}
		}

		// Determine the appropriate action based on the upgrade state
		switch upgradeState {
		case "":
			log.Printf("Upgrade has not started yet...")
		case "Partial":
			log.Printf("Upgrade is in progress. Conditions: %v", conditionMessage)
		case "Completed":
			log.Printf("Upgrade complete!")
			duration := time.Since(startTime)
			log.Printf("upgraded to 4.15.8 duration: %v", duration)
			return nil
		case "Failed":
			log.Printf("Upgrade failed! Conditions: %v", conditionMessage)
			return &upgradeError{err: fmt.Errorf("upgrade failed")}
		default:
			log.Printf("Unknown upgrade state: %s", upgradeState)
		}

		// Wait before the next poll attempt
		time.Sleep(upgradeDelay * time.Second)
	}

	return fmt.Errorf("upgrade is still in progress, failed to finish within max wait attempts")
}

func patchVersionConfig(ctx context.Context, client *openshiftclient.Client, channel string, image string, version string) error {
	clusterVersionConfing := configv1.ClusterVersion{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{Name: "version"}}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{
			"channel": channel,
			"desiredUpdate": map[string]interface{}{
				"version": version, // Replace with your desired version
				"image":   image,   // Replace with your desired image repository
				"force":   true,    // Specify force as true
			},
		},
	})

	if err != nil {
		panic(err)
	}

	if err = client.Patch(
		ctx,
		&clusterVersionConfing,
		k8s.Patch{PatchType: types.MergePatchType, Data: mergePatch},
	); err != nil {
		panic(err)
	}
	return nil

}

func patchNetworkConfig(ctx context.Context, client *openshiftclient.Client) error {

	networkConfig := configv1.Network{ObjectMeta: v1.ObjectMeta{Name: "cluster"}}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"network.openshift.io/network-type-migration": "", // Empty string value for the annotation
			},
		},
		"spec": map[string]interface{}{
			"networkType": "OVNKubernetes", // Setting the network typ
		},
	})

	if err != nil {
		panic(err)
	}

	if err = client.Patch(
		ctx,
		&networkConfig,
		k8s.Patch{PatchType: types.MergePatchType, Data: mergePatch},
	); err != nil {
		panic(err)
	}
	return nil
}
func patchNetworkConfigv1(ctx context.Context, client *openshiftclient.Client) error {

	networkConfig := configv1.Network{ObjectMeta: v1.ObjectMeta{Name: "cluster"}}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"unsupported-red-hat-internal-testing": "true", // Empty string value for the annotation
			},
		},
	})

	if err != nil {
		panic(err)
	}

	if err = client.Patch(
		ctx,
		&networkConfig,
		k8s.Patch{PatchType: types.MergePatchType, Data: mergePatch},
	); err != nil {
		panic(err)
	}
	return nil
}
