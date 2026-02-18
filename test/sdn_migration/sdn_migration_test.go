package sdn_migration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	openshiftclient "github.com/openshift/osde2e-common/pkg/clients/openshift"
	"github.com/openshift/osde2e-common/pkg/clouds/aws"
	osdprovider "github.com/openshift/osde2e-common/pkg/openshift/osd"
	rosaprovider "github.com/openshift/osde2e-common/pkg/openshift/rosa"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

const (
	osdClusterReadyJobName    = "osd-cluster-ready"
	osdClusterReadyJobTimeout = 45 * time.Minute
	version4_15               = "4.15.8"
	version4_16               = "4.16.0-rc.0"
	image4_15                 = "quay.io/openshift-release-dev/ocp-release@sha256:8032c4248d7ae032d5e79debf975d08683cc34d5f08ab2e937ce2d1e940c007b"
	image4_16                 = "quay.io/openshift-release-dev/ocp-release@sha256:c56b01191de4cbb4b97c6eeaf61c5c122fcd465d1d0d671db640d877638ed790"
	channel4_15               = "14.15.8"
	channel4_16               = "4.16.0-rc.0"
	upgradeMaxAttempts        = 3080
	upgradeDelay              = 10
)

type rosaCluster struct {
	id             string
	name           string
	channelGroup   string
	version        string
	reportDir      string
	kubeconfigFile string

	client *openshiftclient.Client
}

var _ = Describe("SDN migration", Ordered, func() {
	var (
		clusterName        = os.Getenv("CLUSTER_NAME")
		testRosaCluster    *rosaCluster
		clusterOptions     *rosaprovider.CreateClusterOptions
		reportDir          = os.Getenv("REPORT_DIR")
		ocmToken           = os.Getenv("OCM_TOKEN")
		clientID           = os.Getenv("CLIENT_ID")
		clientSecret       = os.Getenv("CLIENT_SECRET")
		region             = os.Getenv("AWS_REGION")
		replicas, _        = strconv.Atoi(os.Getenv("REPLICAS"))
		ocmEnv             = ocm.Stage
		logger             = GinkgoLogr
		rosaProvider       *rosaprovider.Provider
		createRosaCluster  = Label("CreateRosaCluster")
		enableClusterProxy = Label("EnableClusterProxy")
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
			Region:          region,
			SecretAccessKey: secretAccessKey,
			AccessKeyID:     accessKeyID,
		})
		Expect(err).ShouldNot(HaveOccurred(), "failed to construct rosa provider")
		osdProvider, err := osdprovider.New(ctx, ocmToken, clientID, clientSecret, ocmEnv, logger)
		Expect(err).ShouldNot(HaveOccurred(), "failed to construct osd provider")
		DeferCleanup(osdProvider.Close)

		if createRosaCluster.MatchesLabelFilter(GinkgoLabelFilter()) {
			clusterOptions = &rosaprovider.CreateClusterOptions{
				ClusterName:                  clusterName,
				Version:                      "4.14.14",
				UseDefaultAccountRolesPrefix: true,
				STS:                          true,
				Mode:                         "auto",
				ChannelGroup:                 "stable",
				ComputeMachineType:           "m5.xlarge",
				Replicas:                     replicas,
				MultiAZ:                      true,
				ETCDEncryption:               true,
				NetworkType:                  "OpenShiftSDN",
				SkipHealthCheck:              true,
			}
		}
		if os.Getenv("CLUSTER_ID") == "" {
			if enableClusterProxy.MatchesLabelFilter(GinkgoLabelFilter()) {
				var (
					httpProxyVar  string
					httpsProxyVar string
					subnets       string
				)
				err = os.Setenv("TF_VAR_aws_access_key_id", accessKeyID)
				Expect(err).ShouldNot(HaveOccurred(), "failed to set TF_VAR_aws_access_key_id")
				err = os.Setenv("TF_VAR_aws_secret_access_key", secretAccessKey)
				Expect(err).ShouldNot(HaveOccurred(), "failed to set TF_VAR_aws_secret_access_key")
				err = os.Setenv("TF_VAR_region", region)
				Expect(err).ShouldNot(HaveOccurred(), "failed to set TF_VAR_region")

				out, err := runTerraformCommand("apply")
				Expect(err).ShouldNot(HaveOccurred(), "failed to apply terraform")
				httpProxyVarMeta, httpsProxyVarMeta, subnetsMeta := out["http_proxy_var"], out["https_proxy_var"], out["subnets"]

				err = json.Unmarshal(httpProxyVarMeta.Value, &httpProxyVar)
				Expect(err).ShouldNot(HaveOccurred(), "failed to unmarshal terraform output http_proxy_var")
				err = json.Unmarshal(httpsProxyVarMeta.Value, &httpsProxyVar)
				Expect(err).ShouldNot(HaveOccurred(), "failed to unmarshal terraform output https_proxy_var")
				err = json.Unmarshal(subnetsMeta.Value, &subnets)
				Expect(err).ShouldNot(HaveOccurred(), "failed to unmarshal terraform output subnets")

				clusterOptions.HTTPSProxy = httpsProxyVar
				clusterOptions.HTTPProxy = httpProxyVar
				clusterOptions.AdditionalTrustBundleFile = "terraform/ca.pem"
				clusterOptions.SubnetIDs = subnets
				clusterOptions.NoProxy = "api.stage.openshift.com"
			}
			testRosaCluster.id, err = rosaProvider.CreateCluster(ctx, clusterOptions)
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
			err := rosaProvider.DeleteCluster(ctx, &rosaprovider.DeleteClusterOptions{
				ClusterName:        testRosaCluster.name,
				WorkingDir:         testRosaCluster.reportDir,
				STS:                true,
				DeleteOidcConfigID: true,
			})
			Expect(err).Should(BeNil(), "failed to delete rosa cluster")
			if enableClusterProxy.MatchesLabelFilter(GinkgoLabelFilter()) {
				_, err = runTerraformCommand("destroy")
				Expect(err).Should(BeNil(), "failed to delete proxy resources in AWS")
			}

		}
	})

	It("rosa cluster is upgraded to 4.15.8 successfully", rosaUpgrade, func(ctx context.Context) {
		err := patchVersionConfig(ctx, testRosaCluster.client, channel4_15, image4_15, version4_15)
		Expect(err).ShouldNot(HaveOccurred(), "rosa cluster upgrade failed")
		err = checkUpgradeStatus(ctx, testRosaCluster.client, version4_15, logger)
		Expect(err).ShouldNot(HaveOccurred(), err)
	})

	It("rosa cluster is healthy post 4.15.8 upgrade", postUpgradeCheck, func(ctx context.Context) {
		err := cluterOperatorHealthCheck(ctx, testRosaCluster.client, logger)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")
	})

	It("rosa cluster is upgraded to 4.16.0-rc.0 successfully", rosaUpgrade, func(ctx context.Context) {
		err := patchVersionConfig(ctx, testRosaCluster.client, channel4_16, image4_16, version4_16)
		Expect(err).ShouldNot(HaveOccurred(), "rosa cluster upgrade failed")
	})

	It("rosa cluster is healthy post 4.16.0-rc.0 upgrade", postUpgradeCheck, func(ctx context.Context) {
		err := cluterOperatorHealthCheck(ctx, testRosaCluster.client, logger)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")
	})

	It("rosa cluster migrated from sdn to ovn successfully", sdnToOvn, func(ctx context.Context) {
		err := addIntenalTestingAnnotation(ctx, testRosaCluster.client)
		Expect(err).ShouldNot(HaveOccurred(), "Rosa Cluster failed to patch network")
		err = patchNetworkConfig(ctx, testRosaCluster.client)
		Expect(err).ShouldNot(HaveOccurred(), "Rosa Cluster failed to patch network")
		err = checkMigrationStatus(ctx, testRosaCluster.client, logger)
		Expect(err).ShouldNot(HaveOccurred(), "Rosa Cluster failed to patch network")
	})
	It("rosa cluster has no critical alerts firing post sdn to ovn migration", postMigrationCheck, func(ctx context.Context) {
		err := cluterOperatorHealthCheck(ctx, testRosaCluster.client, logger)
		Expect(err).ShouldNot(HaveOccurred(), err)
		err = osdClusterReadyHealthCheck(ctx, testRosaCluster.client, testRosaCluster.reportDir)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")
	})
})

func runTerraformCommand(command string) (map[string]tfexec.OutputMeta, error) {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.2.1")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error installing Terraform: %s", err.Error())
	}

	// Initialize the Terraform executable
	tf, err := tfexec.NewTerraform("terraform", execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Terraform: %v", err.Error())
	}

	switch command {
	case "apply":
		err = tf.Init(context.Background(), tfexec.Reconfigure(true))
		if err != nil {
			return nil, fmt.Errorf("error running Init: %s", err.Error())
		}
		err = tf.Apply(context.Background())
		if err != nil {
			cleanupErr := tf.Destroy(context.Background())
			if cleanupErr != nil {
				return nil, fmt.Errorf("failed to cleanup resources after failed apply: %v", err.Error())
			}

			return nil, fmt.Errorf("terraform apply failed: %v", err.Error())
		}

		outputs, err := tf.Output(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get Terraform output: %v", err.Error())
		}
		return outputs, nil

	case "destroy":
		err = tf.Destroy(context.Background())
		if err != nil {
			return nil, fmt.Errorf("terraform destroy failed: %v", err.Error())
		}
		return map[string]tfexec.OutputMeta{}, nil
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

func cluterOperatorHealthCheck(ctx context.Context, clusterClient *openshiftclient.Client, logger logr.Logger) error {
	var (
		err    error
		coList configv1.ClusterOperatorList
		state  string
	)

	for i := 1; i <= upgradeMaxAttempts; i++ {
		state = "healthy"
		err = clusterClient.List(ctx, &coList)
		if err != nil {
			return fmt.Errorf("failed to get cluster operators: %v", err)
		}

		// Iterate over the list of ClusterOperators
		if coList.Items == nil {
			logger.Info("Failed to find cluster operators")
		} else {
			for _, co := range coList.Items {
				// Check if the "Progressing" condition exists and is set to "false"
				progressingCondition := findConditionByType(co.Status.Conditions, "Progressing")
				availableCondition := findConditionByType(co.Status.Conditions, "Available")
				if progressingCondition != nil && progressingCondition.Status == "True" || availableCondition != nil && availableCondition.Status == "False" {
					logger.Info("waiting for cluster operators to be healthy")
					state = "unhealthy"
					break
				}
			}
		}

		nodes := &corev1.NodeList{}
		err = clusterClient.List(ctx, nodes)
		if err != nil {
			return fmt.Errorf("failed to get nodes: %v", err)
		}
		if nodes.Items == nil {
			logger.Info("failed to list nodes")
		} else {
			for _, node := range nodes.Items {
				if node.Spec.Unschedulable == true {
					logger.Info("waiting for the nodes to become schedulable")
					state = "unhealthy"
					break
				}
			}
		}
		switch state {
		case "unhealthy":
			logger.Info("Health check in progress")
		case "healthy":
			logger.Info("Health check complete!")
			return nil

		}
		// Wait before the next poll attempt
		time.Sleep(upgradeDelay * time.Second)

	}
	return errors.New("cluster failed health check and did not become healthy within the maximum wait attempts")
}

// osdClusterReadyHealthCheck verifies the osd-cluster-ready health check job is passing
func osdClusterReadyHealthCheck(ctx context.Context, clusterClient *openshiftclient.Client, reportDir string) error {
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
	newJob.Spec.Template.Name = newJob.GetGenerateName()
	newJob.Spec.Template.Labels = map[string]string{}
	newJob.Spec.Template.Spec.Containers[0].Name = newJob.GetGenerateName()

	if err = clusterClient.Create(ctx, newJob); err != nil {
		return fmt.Errorf("failed to create %s job: %v", newJob.GetName(), err)
	}

	defer func() {
		_ = clusterClient.Delete(ctx, newJob)
	}()

	return clusterClient.OSDClusterHealthy(ctx, reportDir, osdClusterReadyJobTimeout)
}

type upgradeError struct {
	err error
}

// Error returns the formatted error message when upgradeError is invoked
func (e *upgradeError) Error() string {
	return fmt.Sprintf("osd upgrade failed: %v", e.err)
}

// checkMigrationStatus probes the status of the SDN-to-OVN migration
func checkMigrationStatus(ctx context.Context, client *openshiftclient.Client, logger logr.Logger) error {
	var (
		err     error
		network configv1.Network
	)

	for i := 1; i <= upgradeMaxAttempts; i++ {
		// Get the current network configuration
		err = client.Get(ctx, "cluster", "", &network)
		if err != nil {
			return fmt.Errorf("failed to get network configuration: %v", err)
		}

		// Check for the migration condition
		migrationInProgress := false
		for _, cond := range network.Status.Conditions {
			if cond.Type == "NetworkTypeMigrationInProgress" {
				if cond.Status == "True" && cond.Reason == "NetworkTypeMigrationStarted" {
					migrationInProgress = true
					break
				}

				if cond.Status == "False" && cond.Reason == "NetworkTypeMigrationCompleted" {
					logger.Info("Network migration completed successfully!")
					return nil
				}
			}
		}

		if migrationInProgress {
			logger.Info("Network migration is still in progress...")
		} else {
			logger.Info("Migration status is unknown or not in progress.")
		}

		time.Sleep(upgradeDelay * time.Second)
	}

	return fmt.Errorf("network migration is still in progress after %d attempts", upgradeMaxAttempts)
}

// checkUpgradeStatus probes the status of the cluster upgrade
func checkUpgradeStatus(ctx context.Context, client *openshiftclient.Client, upgradeVersion string, logger logr.Logger) error {
	var (
		err error
		cv  configv1.ClusterVersion
	)

	for i := 1; i <= upgradeMaxAttempts; i++ {

		err = client.Get(ctx, "version", "", &cv)
		if err != nil {
			logger.Info("Failed to get cluster version config", "error", err)
			time.Sleep(upgradeDelay * time.Second)
			continue
		}

		// Extract the status map from the ClusterVersion configuration
		status := cv.Status
		if status.History == nil {
			logger.Info("Failed to find history in cluster version config: status history is nil")
			time.Sleep(upgradeDelay * time.Second)
			continue
		}

		// Extract the history slice from the status
		var upgradeState string
		for _, h := range status.History {
			// Check if the version matches the desired upgrade version
			if h.Version == upgradeVersion {
				// Extract the state for the matching version
				upgradeState = string(h.State)
				break
			}
		}

		// Extract the conditions from the status
		conditions := status.Conditions
		if conditions == nil {
			logger.Info("Failed to find conditions in status: conditions are nil")
		} else {
			// Filter for the condition message that starts with "Working towards"
			for _, cond := range conditions {
				// Extract the condition message
				message := cond.Message
				if strings.HasPrefix(message, "Working towards") {
					break
				}
			}
		}

		// Determine the appropriate action based on the upgrade state
		switch upgradeState {
		case "":
			logger.Info("Upgrade has not started yet...")
		case "Partial":
			logger.Info("Upgrade is in progress.")
		case "Completed":
			logger.Info("Upgrade complete!")
			return nil
		case "Failed":
			logger.Info("Upgrade failed!")
			return &upgradeError{err: fmt.Errorf("upgrade failed")}
		default:
			logger.Info("Unknown upgrade state", "state", upgradeState)
		}

		// Wait before the next poll attempt
		time.Sleep(upgradeDelay * time.Second)
	}

	return fmt.Errorf("upgrade is still in progress, failed to finish within max wait attempts")
}

// findConditionByType finds a specific condition by type
func findConditionByType(conditions []configv1.ClusterOperatorStatusCondition, conditionType configv1.ClusterStatusConditionType) *configv1.ClusterOperatorStatusCondition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}

// patchVersionConfig updates the version config to the desired version to initiate an upgrade
func patchVersionConfig(ctx context.Context, client *openshiftclient.Client, channel string, image string, version string) error {
	clusterVersionConfing := configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "version"},
	}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{
			"channel": channel,
			"desiredUpdate": map[string]interface{}{
				"version": version,
				"image":   image,
				"force":   true,
			},
		},
	})
	if err != nil {
		return err
	}

	if err = client.Patch(
		ctx,
		&clusterVersionConfing,
		k8s.Patch{PatchType: types.MergePatchType, Data: mergePatch},
	); err != nil {
		return err
	}
	return nil
}

// patchNetworkConfig updates network type to OVN
func patchNetworkConfig(ctx context.Context, client *openshiftclient.Client) error {
	networkConfig := configv1.Network{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"network.openshift.io/network-type-migration": "",
			},
		},
		"spec": map[string]interface{}{
			"networkType": "OVNKubernetes",
		},
	})
	if err != nil {
		return err
	}

	if err = client.Patch(
		ctx,
		&networkConfig,
		k8s.Patch{PatchType: types.MergePatchType, Data: mergePatch},
	); err != nil {
		return err
	}
	return nil
}

// addAnnotation adds the internal testing annotation to the network config
func addIntenalTestingAnnotation(ctx context.Context, client *openshiftclient.Client) error {
	networkConfig := configv1.Network{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"unsupported-red-hat-internal-testing": "true",
			},
		},
	})
	if err != nil {
		return err
	}

	if err = client.Patch(
		ctx,
		&networkConfig,
		k8s.Patch{PatchType: types.MergePatchType, Data: mergePatch},
	); err != nil {
		return err
	}
	return nil
}
