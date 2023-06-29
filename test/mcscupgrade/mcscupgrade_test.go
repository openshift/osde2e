package mcscupgrade_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ocmclient "github.com/openshift/osde2e-common/pkg/clients/ocm"
	openshiftclient "github.com/openshift/osde2e-common/pkg/clients/openshift"
	prometheusclient "github.com/openshift/osde2e-common/pkg/clients/prometheus"
	osdprovider "github.com/openshift/osde2e-common/pkg/openshift/osd"
	rosaprovider "github.com/openshift/osde2e-common/pkg/openshift/rosa"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	osdClusterReadyJobName    = "osd-cluster-ready"
	osdClusterReadyJobTimeout = 45 * time.Minute
)

type managementCluster struct {
	id             string
	name           string
	kubeconfigFile string
	reportDir      string

	osdFleetMgmtID string

	clusterVersion *semver.Version
	upgradeVersion *semver.Version

	client *openshiftclient.Client
}

type serviceCluster struct {
	id             string
	name           string
	kubeconfigFile string
	reportDir      string

	osdFleetMgmtID string

	clusterVersion *semver.Version
	upgradeVersion *semver.Version

	client *openshiftclient.Client
}

type rosaHCPCluster struct {
	id               string
	name             string
	channelGroup     string
	version          string
	kubeconfigFile   string
	provisionShardID string
	reportDir        string

	client *openshiftclient.Client
}

var _ = Describe("HyperShift", Ordered, func() {
	var (
		applyHCPWorkloads     = Label("ApplyHCPWorkloads")
		mcUpgrade             = Label("MCUpgrade")
		mcUpgradeHealthChecks = Label("MCUpgradeHealthChecks")
		removeHCPWorkloads    = Label("RemoveHCPWorkloads")
		scUpgrade             = Label("SCUpgrade")
		scUpgradeHealthChecks = Label("SCUpgradeHealthChecks")

		err error

		osdProvider  *osdprovider.Provider
		rosaProvider *rosaprovider.Provider

		hcpCluster *rosaHCPCluster
		mcCluster  *managementCluster
		scCluster  *serviceCluster

		logger = GinkgoLogr
	)

	_ = BeforeAll(func(ctx context.Context) {
		var (
			ocmEnv = ocmclient.Integration

			ocmToken    = os.Getenv("OCM_TOKEN")
			upgradeType = os.Getenv("UPGRADE_TYPE")
			reportDir   = getEnvVar("REPORT_DIR", envconf.RandomName(fmt.Sprintf("%s/mcscupgrade", os.TempDir()), 25))
		)

		hcpCluster = &rosaHCPCluster{
			name:             getEnvVar("CLUSTER_NAME", envconf.RandomName("hcp", 8)),
			channelGroup:     getEnvVar("CLUSTER_CHANNEL_GROUP", "candidate"),
			provisionShardID: os.Getenv("PROVISION_SHARD_ID"),
			version:          getEnvVar("CLUSTER_VERSION", "4.12.20"),
		}
		hcpCluster.reportDir = fmt.Sprintf("%s/%s", reportDir, hcpCluster.name)

		mcCluster = &managementCluster{
			osdFleetMgmtID: os.Getenv("OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID"),
		}

		scCluster = &serviceCluster{
			osdFleetMgmtID: os.Getenv("OSD_FLEET_MGMT_SERVICE_CLUSTER_ID"),
		}

		Expect(os.MkdirAll(reportDir, os.ModePerm)).ShouldNot(HaveOccurred(), "failed to create report directory")
		Expect(os.MkdirAll(hcpCluster.reportDir, os.ModePerm)).ShouldNot(HaveOccurred(), "failed to create hosted control plane cluster report directory")

		Expect(ocmToken).ShouldNot(BeEmpty(), "ocm token is undefined")
		Expect(scCluster.osdFleetMgmtID).ShouldNot(BeEmpty(), "osd fleet manager service cluster id is undefined")
		Expect(mcCluster.osdFleetMgmtID).ShouldNot(BeEmpty(), "osd fleet manager management cluster id is undefined")
		Expect(hcpCluster.provisionShardID).ShouldNot(BeEmpty(), "osd fleet manager service cluster provision shard id is undefined")
		Expect(upgradeType).Should(BeElementOf([]string{"Y", "Z"}), "upgrade type is invalid")

		rosaProvider, err = rosaprovider.New(ctx, ocmToken, ocmEnv, logger)
		Expect(err).ShouldNot(HaveOccurred(), "failed to construct rosa provider")

		osdProvider, err = osdprovider.New(ctx, ocmToken, ocmEnv, logger)
		Expect(err).ShouldNot(HaveOccurred(), "failed to construct osd provider")
		DeferCleanup(osdProvider.Client.Close)

		if scUpgrade.MatchesLabelFilter(GinkgoLabelFilter()) || scUpgradeHealthChecks.MatchesLabelFilter(GinkgoLabelFilter()) {
			osdFleetManagerSC, err := osdProvider.OSDFleetMgmt().V1().ServiceClusters().ServiceCluster(scCluster.osdFleetMgmtID).Get().SendContext(ctx)
			Expect(err).ShouldNot(HaveOccurred(), "osd fleet manager api request failed to get service cluster: %q", scCluster.osdFleetMgmtID)
			Expect(osdFleetManagerSC).NotTo(BeNil(), "osd fleet manager service cluster: %q does not exist", scCluster.osdFleetMgmtID)

			serviceCluster, err := osdProvider.ClustersMgmt().V1().Clusters().Cluster(osdFleetManagerSC.Body().ClusterManagementReference().ClusterId()).Get().SendContext(ctx)
			Expect(err).ShouldNot(HaveOccurred(), "failed to get get service cluster: %s", scCluster.osdFleetMgmtID)

			scCluster.id = serviceCluster.Body().ID()
			scCluster.name = serviceCluster.Body().Name()
			scCluster.clusterVersion, err = semver.NewVersion(serviceCluster.Body().Version().RawID())
			Expect(err).ShouldNot(HaveOccurred(), "failed to parse service cluster installed version to semantic version")

			availableVersions := serviceCluster.Body().Version().AvailableUpgrades()
			totalUpgradeVersionsAvailable := len(availableVersions)
			Expect(totalUpgradeVersionsAvailable).ToNot(BeNumerically("==", 0), "service cluster has no available supported upgrade versions")

			for i := 0; i < totalUpgradeVersionsAvailable; i++ {
				version, err := semver.NewVersion(availableVersions[totalUpgradeVersionsAvailable-i-1])
				Expect(err).ShouldNot(HaveOccurred(), "failed to parse service cluster upgrade version to semantic version")
				if (scCluster.clusterVersion.Minor() == version.Minor()) && upgradeType == "Z" {
					scCluster.upgradeVersion = version
					break
				} else if (scCluster.clusterVersion.Minor() < version.Minor()) && upgradeType == "Y" {
					scCluster.upgradeVersion = version
					break
				}
			}
			Expect(scCluster.upgradeVersion).ToNot(BeNil(), "failed to identify service cluster %q upgrade version", scCluster.osdFleetMgmtID)

			scCluster.kubeconfigFile, err = osdProvider.KubeconfigFile(ctx, scCluster.id, os.TempDir())
			Expect(err).ShouldNot(HaveOccurred(), "failed to get service cluster %q kubeconfig file", scCluster.osdFleetMgmtID)

			scCluster.client, err = openshiftclient.NewFromKubeconfig(scCluster.kubeconfigFile, logger)
			Expect(err).ShouldNot(HaveOccurred(), "failed to construct service cluster client")

			scCluster.reportDir = fmt.Sprintf("%s/%s", reportDir, scCluster.name)
			Expect(os.MkdirAll(scCluster.reportDir, os.ModePerm)).ShouldNot(HaveOccurred(), "failed to create service cluster report directory")
		}

		if mcUpgrade.MatchesLabelFilter(GinkgoLabelFilter()) || mcUpgradeHealthChecks.MatchesLabelFilter(GinkgoLabelFilter()) {
			osdFleetManagerMC, err := osdProvider.OSDFleetMgmt().V1().ManagementClusters().ManagementCluster(mcCluster.osdFleetMgmtID).Get().SendContext(ctx)
			Expect(err).ShouldNot(HaveOccurred(), "osd fleet manager api request failed to get management cluster: %q", mcCluster.osdFleetMgmtID)
			Expect(osdFleetManagerMC).NotTo(BeNil(), "osd fleet manager management cluster: %q does not exist", mcCluster.osdFleetMgmtID)

			ocmMC, err := osdProvider.ClustersMgmt().V1().Clusters().Cluster(osdFleetManagerMC.Body().ClusterManagementReference().ClusterId()).Get().SendContext(ctx)
			Expect(err).ShouldNot(HaveOccurred(), "failed to get get management cluster: %s", mcCluster.osdFleetMgmtID)

			mcCluster.id = ocmMC.Body().ID()
			mcCluster.name = ocmMC.Body().Name()
			mcCluster.clusterVersion, err = semver.NewVersion(ocmMC.Body().Version().RawID())
			Expect(err).ShouldNot(HaveOccurred(), "failed to parse management cluster installed version to semantic version")

			availableVersions := ocmMC.Body().Version().AvailableUpgrades()
			totalUpgradeVersionsAvailable := len(availableVersions)
			Expect(totalUpgradeVersionsAvailable).ToNot(BeNumerically("==", 0), "management cluster has no available supported upgrade versions")

			for i := 0; i < totalUpgradeVersionsAvailable; i++ {
				version, err := semver.NewVersion(availableVersions[totalUpgradeVersionsAvailable-i-1])
				Expect(err).ShouldNot(HaveOccurred(), "failed to parse management cluster upgrade version to semantic version")
				if (mcCluster.clusterVersion.Minor() == version.Minor()) && upgradeType == "Z" {
					mcCluster.upgradeVersion = version
					break
				} else if (mcCluster.clusterVersion.Minor() < version.Minor()) && upgradeType == "Y" {
					mcCluster.upgradeVersion = version
					break
				}
			}
			Expect(mcCluster.upgradeVersion).ToNot(BeNil(), "failed to identify service cluster %q upgrade version", mcCluster.osdFleetMgmtID)

			mcCluster.kubeconfigFile, err = osdProvider.KubeconfigFile(ctx, mcCluster.id, os.TempDir())
			Expect(err).ShouldNot(HaveOccurred(), "failed to get management cluster %q kubeconfig file", mcCluster.osdFleetMgmtID)

			mcCluster.client, err = openshiftclient.NewFromKubeconfig(mcCluster.kubeconfigFile, logger)
			Expect(err).ShouldNot(HaveOccurred(), "failed to construct management cluster client")

			mcCluster.reportDir = fmt.Sprintf("%s/%s", reportDir, mcCluster.name)
			Expect(os.MkdirAll(mcCluster.reportDir, os.ModePerm)).ShouldNot(HaveOccurred(), "failed to create management cluster report directory")
		}

		if applyHCPWorkloads.MatchesLabelFilter(GinkgoLabelFilter()) {
			hcpCluster.id, err = rosaProvider.CreateCluster(ctx, &rosaprovider.CreateClusterOptions{
				ClusterName:  hcpCluster.name,
				Version:      hcpCluster.version,
				ChannelGroup: hcpCluster.channelGroup,
				HostedCP:     true,
				Properties:   map[string]string{"provision_shard_id": hcpCluster.provisionShardID},
				WorkingDir:   hcpCluster.reportDir,
			})
			Expect(err).ShouldNot(HaveOccurred(), "failed to create hosted control plane cluster")

			hcpCluster.kubeconfigFile, err = rosaProvider.KubeconfigFile(ctx, hcpCluster.id, os.TempDir())
			Expect(err).ShouldNot(HaveOccurred(), "failed to get hosted control plane cluster %q kubeconfig file", hcpCluster.id)

			hcpCluster.client, err = openshiftclient.NewFromKubeconfig(hcpCluster.kubeconfigFile, logger)
			Expect(err).ShouldNot(HaveOccurred(), "failed to construct hosted control plane cluster client")
		}
	})

	_ = AfterAll(func(ctx context.Context) {
		if removeHCPWorkloads.MatchesLabelFilter(GinkgoLabelFilter()) && hcpCluster.id != "" {
			err := rosaProvider.DeleteCluster(ctx, &rosaprovider.DeleteClusterOptions{
				ClusterName:        hcpCluster.name,
				ClusterID:          hcpCluster.id,
				HostedCP:           true,
				WorkingDir:         hcpCluster.reportDir,
				DeleteHostedCPVPC:  true,
				DeleteOidcConfigID: true,
			})
			Expect(err).ShouldNot(HaveOccurred(), "failed to delete hosted control plane cluster")
		}
	})

	It("service cluster is healthy pre upgrade", scUpgrade, func(ctx context.Context) {
		criticalAlerts, _, err := queryPrometheusAlerts(ctx, scCluster.client, fmt.Sprintf("%s/prometheus-alerts-pre-upgrade.log", scCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing pre upgrade")

		err = scCluster.client.OSDClusterHealthy(ctx, osdClusterReadyJobName, scCluster.reportDir, osdClusterReadyJobTimeout)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed pre upgrade")
	})

	It("service cluster is upgraded successfully", scUpgrade, func(ctx context.Context) {
		err = osdProvider.OCMUpgrade(ctx, scCluster.client, scCluster.id, *scCluster.clusterVersion, *scCluster.upgradeVersion)
		Expect(err).ShouldNot(HaveOccurred(), "service cluster upgrade failed")
	})

	It("service cluster is healthy post upgrade", scUpgradeHealthChecks, func(ctx context.Context) {
		criticalAlerts, _, err := queryPrometheusAlerts(ctx, scCluster.client, fmt.Sprintf("%s/prometheus-alerts-post-upgrade.log", scCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing post upgrade")

		err = osdClusterReadyHealthCheck(ctx, scCluster.client, "post-upgrade", scCluster.reportDir)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")
	})

	It("hcp cluster is healthy post service cluster upgrade", scUpgradeHealthChecks, func(ctx context.Context) {
		if hcpCluster.kubeconfigFile == "" {
			Skip("Unable to locate hosted control plane cluster kubeconfig, skipping health checks")
		}

		criticalAlerts, _, err := queryPrometheusAlerts(ctx, hcpCluster.client, fmt.Sprintf("%s/prometheus-alerts-post-sc-upgrade.log", hcpCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing post upgrade")
	})

	It("management cluster is healthy pre upgrade", mcUpgrade, func(ctx context.Context) {
		criticalAlerts, _, err := queryPrometheusAlerts(ctx, mcCluster.client, fmt.Sprintf("%s/prometheus-alerts-pre-upgrade.log", mcCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing pre upgrade")

		err = mcCluster.client.OSDClusterHealthy(ctx, osdClusterReadyJobName, mcCluster.reportDir, osdClusterReadyJobTimeout)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed pre upgrade")
	})

	It("management cluster is upgraded successfully", mcUpgrade, func(ctx context.Context) {
		err = osdProvider.OCMUpgrade(ctx, mcCluster.client, mcCluster.id, *mcCluster.clusterVersion, *mcCluster.upgradeVersion)
		Expect(err).ShouldNot(HaveOccurred(), "management cluster upgrade failed")
	})

	It("management cluster is healthy post upgrade", mcUpgradeHealthChecks, func(ctx context.Context) {
		criticalAlerts, _, err := queryPrometheusAlerts(ctx, mcCluster.client, fmt.Sprintf("%s/prometheus-alerts-post-upgrade.log", mcCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing post upgrade")

		err = osdClusterReadyHealthCheck(ctx, mcCluster.client, "post-upgrade", mcCluster.reportDir)
		Expect(err).ShouldNot(HaveOccurred(), "osd-cluster-ready health check job failed post upgrade")
	})

	It("hcp cluster has no critical alerts firing post management cluster upgrade", mcUpgradeHealthChecks, func(ctx context.Context) {
		if hcpCluster.kubeconfigFile == "" {
			Skip("Unable to locate hosted control plane cluster kubeconfig, skipping health checks")
		}

		criticalAlerts, _, err := queryPrometheusAlerts(ctx, hcpCluster.client, fmt.Sprintf("%s/prometheus-alerts-post-mc-upgrade.log", hcpCluster.reportDir))
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve prometheus alerts")
		Expect(criticalAlerts).ToNot(BeNumerically(">", 0), "critical alerts are firing post upgrade")
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

// getEnvVar returns the env variable value and if unset returns default provided
func getEnvVar(key, value string) string {
	result, exist := os.LookupEnv(key)
	if exist {
		return result
	}
	return value
}
