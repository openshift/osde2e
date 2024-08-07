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
	//"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	//v1 "k8s.io/client-go/applyconfigurations/meta/v1"
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
)

type rosaCluster struct {
	id           string
	name         string
	channelGroup string
	version      *semver.Version
	//provisionShardID string
	reportDir      string
	upgradeVersion *semver.Version
	kubeconfigFile string

	client *openshiftclient.Client
}

var _ = Describe("SDN migration", ginkgo.Ordered, func() {
	const clusterName = "creed-sdn-ovn-1"
	var (
		testRosaCluster *rosaCluster
		reportDir       = getEnvVar("REPORT_DIR", envconf.RandomName(fmt.Sprintf("%s/sdn_migration", os.TempDir()), 25))
		ocmToken        = os.Getenv("OCM_TOKEN")
		clientID        = os.Getenv("CLIENT_ID")
		clientSecret    = os.Getenv("CLIENT_SECRET")
		ocmEnv          = ocm.Stage
		upgradeType     = os.Getenv("UPGRADE_TYPE")
		logger          = GinkgoLogr
		rosaProvider    *rosaprovider.Provider
		//osdProvider       *osdprovider.Provider
		createRosaCluster  = Label("CreateRosaCluster")
		removeRosaCluster  = Label("RemoveRosaCluster")
		postMigrationCheck = Label("PostMigrationCheck")
		rosaUpgrade        = Label("RosaUpgrade")
		postUpgradeCheck   = Label("PostUpgradeCheck")
		sdnToOvn           = Label("SdnToOvn")
	)

	_ = BeforeAll(func(ctx context.Context) {
		var err error
		testRosaCluster = &rosaCluster{}

		Expect(ocmToken).ShouldNot(BeEmpty(), "ocm token is undefined")

		rosaProvider, err = rosaprovider.New(ctx, ocmToken, clientID, clientSecret, ocmEnv, logger, &aws.AWSCredentials{
			//Profile:         "",
			Region:          "us-east-1",
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
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
		version := rosaCluster.Body().Version().RawID()
		testRosaCluster.version, _ = semver.NewVersion(version)
		testRosaCluster.channelGroup = rosaCluster.Body().Version().ChannelGroup()

		availableVersions := rosaCluster.Body().Version().AvailableUpgrades()
		totalUpgradeVersionsAvailable := len(availableVersions)
		Expect(totalUpgradeVersionsAvailable).ToNot(BeNumerically("==", 0), "rosa cluster has no available supported upgrade versions")

		for i := 0; i < totalUpgradeVersionsAvailable; i++ {
			version, err := semver.NewVersion(availableVersions[totalUpgradeVersionsAvailable-i-1])
			Expect(err).ShouldNot(HaveOccurred(), "failed to parse service cluster upgrade version to semantic version")
			if (testRosaCluster.version.Minor() == version.Minor()) && upgradeType == "Z" {
				testRosaCluster.upgradeVersion = version
				break
			} else if (testRosaCluster.version.Minor() < version.Minor()) && upgradeType == "Y" {
				testRosaCluster.upgradeVersion = version
				break
			}
		}

		testRosaCluster.kubeconfigFile, err = rosaProvider.KubeconfigFile(ctx, testRosaCluster.id, os.TempDir())
		Expect(err).ShouldNot(HaveOccurred())

		testRosaCluster.client, err = openshiftclient.NewFromKubeconfig(testRosaCluster.kubeconfigFile, logger)
		Expect(err).ShouldNot(HaveOccurred(), "failed to construct service cluster client")

		testRosaCluster.reportDir = fmt.Sprintf("%s/%s", reportDir, testRosaCluster.name)
		//Expect(os.MkdirAll(testRosaCluster.reportDir, os.ModePerm)).ShouldNot(HaveOccurred(), "failed to create service cluster report directory")

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
			Expect(err).Should(BeNil(), "FAILED HELP")
		}
	})

	//It("Rosa 14.14.14 cluster healthy pre-upgrade", preUpgradeCheck, func(ctx context.Context) {
	//	err := testRosaCluster.client.OSDClusterHealthy(ctx, osdClusterReadyJobName, testRosaCluster.reportDir, osdClusterReadyJobTimeout)
	//	Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed pre upgrade")
	//
	//})
	It("rosa cluster is upgraded successfully", rosaUpgrade, func(ctx context.Context) {
		osdProvider, err := osdprovider.New(ctx, ocmToken, clientID, clientSecret, ocmEnv, logger)
		err = osdProvider.OCMUpgrade(ctx, testRosaCluster.client, testRosaCluster.id, *testRosaCluster.version, *testRosaCluster.upgradeVersion)
		Expect(err).ShouldNot(HaveOccurred(), "rosa cluster upgrade failed")

	})

	It("rosa cluster is healthy post upgrade", postUpgradeCheck, func(ctx context.Context) {
		criticalAlerts, _, err := queryPrometheusAlerts(ctx, testRosaCluster.client, fmt.Sprintf("%s/prometheus-alerts-pre-upgrade.log", testRosaCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing pre upgrade")

		err = osdClusterReadyHealthCheck(ctx, testRosaCluster.client, "post-upgrade", testRosaCluster.reportDir)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")

	})
	It("rosa cluster migrated from sdn to ovn successfully", sdnToOvn, func(ctx context.Context) {
		err := patchNetworkConfig(ctx, testRosaCluster.client)
		Expect(err).ShouldNot(HaveOccurred(), "Rosa Cluster failed to patch network")
		//TODO add check to verify the migration was completed

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

func patchNetworkConfig(ctx context.Context, client *openshiftclient.Client) error {

	networkConfig := configv1.Network{ObjectMeta: v1.ObjectMeta{Name: "cluster"}}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"network.openshift.io/network-type-migration": "", // Empty string value for the annotation
			},
		},
		"spec": map[string]interface{}{
			"networkType": "OVNKubernetes", // Setting the network type
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
