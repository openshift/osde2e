// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

const (
	// Provider is what provider to use to create/delete clusters.
	// Env: PROVIDER
	Provider = "provider"

	// JobName lets you name the current e2e job run
	// Env: JOB_NAME
	JobName = "jobName"

	// JobID is the ID designated by prow for this run
	// Env: BUILD_NUMBER
	JobID = "jobID"

	// BaseJobURL is the root location for all job artifacts
	// For example, https://storage.googleapis.com/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/61/build-log.txt would be
	// https://storage.googleapis.com/origin-ci-test/logs -- This is also our default
	// Env: BASE_JOB_URL
	BaseJobURL = "baseJobURL"

	// BaseProwURL is the root location of Prow
	// Env: BASE_PROW_URL
	BaseProwURL = "baseProwURL"

	// Artifacts is the artifacts location on prow. It is an alias for report dir.
	// Env: ARTIFACTS
	Artifacts = "artifacts"

	// ReportDir is the location JUnit XML results are written.
	// Env: REPORT_DIR
	ReportDir = "reportDir"

	// Suffix is used at the end of test names to identify them.
	// Env: SUFFIX
	Suffix = "suffix"

	// DryRun lets you run osde2e all the way up to the e2e tests then skips them.
	// Env: DRY_RUN
	DryRun = "dryRun"

	// MustGather will run a Must-Gather process upon completion of the tests.
	// Env: MUST_GATHER
	MustGather = "mustGather"

	// InstalledWorkloads is an internal variable used to track currently installed workloads in this test run.
	InstalledWorkloads = "installedWorkloads"

	// Phase is an internal variable used to track the current set of tests being run (install, upgrade).
	Phase = "phase"

	// Project is both the project and SA automatically created to house all objects created during an osde2e-run
	Project = "project"
)

// This is a config key to secret file mapping. We will attempt to read in from secret files before loading anything else.
var keyToSecretMapping = map[string]string{}
var keyToSecretMappingMutex = sync.Mutex{}

// Upgrade config keys.
var Upgrade = struct {
	// UpgradeToCISIfPossible will upgrade to the most recent cluster image set if it's newer than the install version
	// Env: UPGRADE_TO_CIS_IF_POSSIBLE
	UpgradeToCISIfPossible string

	// OnlyUpgradeToZReleases will restrict upgrades to selecting Z releases on stage/prod.
	// Env: ONLY_UPGRADE_TO_Z_RELEASES
	OnlyUpgradeToZReleases string

	// NextReleaseAfterProdDefaultForUpgrade will select the cluster image set that the given number of releases away from the the production default.
	// Env: NEXT_RELEASE_AFTER_PROD_DEFAULT_FOR_UPGRADE
	NextReleaseAfterProdDefaultForUpgrade string

	// ReleaseStream used to retrieve latest release images. If set, it will be used to perform an upgrade.
	// Env: UPGRADE_RELEASE_STREAM
	ReleaseStream string

	// ReleaseName is the name of the release in a release stream.
	// Env: UPGRADE_RELEASE_NAME
	ReleaseName string

	// Image is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.
	// Env: UPGRADE_IMAGE
	Image string

	// UpgradeVersionEqualToInstallVersion is true if the install version and upgrade versions are the same.
	UpgradeVersionEqualToInstallVersion string

	// MonitorRoutesDuringUpgrade will monitor the availability of routes whilst an upgrade takes place
	// Env: UPGRADE_MONITOR_ROUTES
	MonitorRoutesDuringUpgrade string

	// Perform an upgrade using the Managed Upgrade Operator
	ManagedUpgrade string

	// Wait for workers to upgrade before considering upgrade complete
	WaitForWorkersToManagedUpgrade string

	// Create disruptive Pod Disruption Budget workloads to test the Managed Upgrade Operator's ability to handle them.
	ManagedUpgradeTestPodDisruptionBudgets string

	// Create disruptive Node Drain workload to test the Managed Upgrade Operator's ability to handle them.
	ManagedUpgradeTestNodeDrain string
}{
	UpgradeToCISIfPossible:                 "upgrade.upgradeToCISIfPossible",
	OnlyUpgradeToZReleases:                 "upgrade.onlyUpgradeToZReleases",
	NextReleaseAfterProdDefaultForUpgrade:  "upgrade.nextReleaseAfterProdDefaultForUpgrade",
	ReleaseStream:                          "upgrade.releaseStream",
	ReleaseName:                            "upgrade.releaseName",
	Image:                                  "upgrade.image",
	UpgradeVersionEqualToInstallVersion:    "upgrade.upgradeVersionEqualToInstallVersion",
	MonitorRoutesDuringUpgrade:             "upgrade.monitorRoutesDuringUpgrade",
	ManagedUpgrade:                         "upgrade.managedUpgrade",
	WaitForWorkersToManagedUpgrade:         "upgrade.waitForWorkersToManagedUpgrade",
	ManagedUpgradeTestPodDisruptionBudgets: "upgrade.managedUpgradeTestPodDisruptionBudgets",
	ManagedUpgradeTestNodeDrain:            "upgrade.managedUpgradeTestNodeDrain",
}

// Kubeconfig config keys.
var Kubeconfig = struct {
	// Path is the filepath of an existing Kubeconfig
	// Env: TEST_KUBECONFIG
	Path string

	// Contents is the actual contents of a valid Kubeconfig
	Contents string
}{
	Path:     "kubeconfig.path",
	Contents: "kubeconfig.contents",
}

// Tests config keys

// Tests config keys.
var Tests = struct {
	// PollingTimeout is how long (in seconds) to wait for an object to be created before failing the test.
	// Env: POLLING_TIMEOUT
	PollingTimeout string

	// GinkgoSkip is a regex passed to Ginkgo that skips any test suites matching the regex. ex. "Operator"
	// Env: GINKGO_SKIP
	GinkgoSkip string

	// GinkgoFocus is a regex passed to Ginkgo that focus on any test suites matching the regex. ex. "Operator"
	// Env: GINKGO_FOCUS
	GinkgoFocus string

	// TestsToRun is a list of files which should be executed as part of a test suite
	// Env: TESTS_TO_RUN
	TestsToRun string

	// SuppressSkipNotifications suppresses the notifications of skipped tests
	// Env: SUPPRESS_SKIP_NOTIFICATIONS
	SuppressSkipNotifications string

	// CleanRuns is the number of times the test-version is run before skipping.
	// Env: CLEAN_RUNS
	CleanRuns string

	// OperatorSkip is a comma-delimited list of operator names to ignore health checks from. ex. "insights,telemetry"
	// Env: OPERATOR_SKIP
	OperatorSkip string

	// SkipClusterHealthChecks skips the cluster health checks. Useful when developing against a running cluster.
	// Env: SKIP_CLUSTER_HEALTH_CHECKS
	SkipClusterHealthChecks string

	// MetricsBucket is the bucket that metrics data will be uploaded to.
	// Env: METRICS_BUCKET
	MetricsBucket string

	// ServiceAccount defines what user the tests should run as. By default, osde2e uses system:admin
	// Env: SERVICE_ACCOUNT
	ServiceAccount string
}{

	PollingTimeout:            "tests.pollingTimeout",
	GinkgoSkip:                "tests.ginkgoSkip",
	GinkgoFocus:               "tests.focus",
	TestsToRun:                "tests.testsToRun",
	SuppressSkipNotifications: "tests.suppressSkipNotifications",
	CleanRuns:                 "tests.cleanRuns",
	OperatorSkip:              "tests.operatorSkip",
	SkipClusterHealthChecks:   "tests.skipClusterHealthChecks",
	MetricsBucket:             "tests.metricsBucket",
	ServiceAccount:            "tests.serviceAccount",
}

// Cluster config keys.
var Cluster = struct {
	// MultiAZ deploys a cluster across multiple availability zones.
	// Env: MULTI_AZ
	MultiAZ string

	// DestroyClusterAfterTest set to true if you want to the cluster to be explicitly deleted after the test.
	// Env: DESTROY_CLUSTER
	DestroyAfterTest string

	// ExpiryInMinutes is how long before a cluster expires and is deleted by OSD.
	// Env: CLUSTER_EXPIRY_IN_MINUTES
	ExpiryInMinutes string

	// AfterTestWait is how long to keep a cluster around after tests have run.
	// Env: AFTER_TEST_CLUSTER_WAIT
	AfterTestWait string

	// InstallTimeout is how long to wait before failing a cluster launch.
	// Env: CLUSTER_UP_TIMEOUT
	InstallTimeout string

	// UseLatestVersionForInstall will select the latest cluster image set available for a fresh install.
	// Env: USE_LATEST_VERSION_FOR_INSTALL
	UseLatestVersionForInstall string

	// UseMiddleClusterImageSetForInstall will select the cluster image set that is in the middle of the list of ordered cluster versions known to OCM.
	// Env: USE_MIDDLE_CLUSTER_IMAGE_SET_FOR_INSTALL
	UseMiddleClusterImageSetForInstall string

	// UseOldestClusterImageSetForInstall will select the cluster image set that is in the end of the list of ordered cluster versions known to OCM.
	// Env: USE_OLDEST_CLUSTER_IMAGE_SET_FOR_INSTALL
	UseOldestClusterImageSetForInstall string

	// DeltaReleaseFromDefault will select the cluster image set that is the given number of releases from the current default in either direction.
	// Env: DELTA_RELEASE_FROM_DEFAULT
	DeltaReleaseFromDefault string

	// NextReleaseAfterProdDefault will select the cluster image set that the given number of releases away from the the production default.
	// Env: NEXT_RELEASE_AFTER_PROD_DEFAULT
	NextReleaseAfterProdDefault string

	// CleanCheckRuns lets us set the number of osd-verify checks we want to run before deeming a cluster "healthy"
	// Env: CLEAN_CHECK_RUNS
	CleanCheckRuns string

	// ID identifies the cluster. If set at start, an existing cluster is tested.
	// Env: CLUSTER_ID
	ID string

	// Name is the name of the cluster being created.
	// Env: CLUSTER_NAME
	Name string

	// Version is the version of the cluster being deployed.
	// Env: CLUSTER_VERSION
	Version string

	// EnoughVersionsForOldestOrMiddleTest is true if there were enough versions for an older/middle test.
	EnoughVersionsForOldestOrMiddleTest string

	// PreviousVersionFromDefaultFound is true if a previous version from default was found.
	PreviousVersionFromDefaultFound string

	// ProvisionShardID is the shard ID that is set to provision a shard for the cluster.
	ProvisionShardID string
}{
	MultiAZ:                             "cluster.multiAZ",
	DestroyAfterTest:                    "cluster.destroyAfterTest",
	ExpiryInMinutes:                     "cluster.expiryInMinutes",
	AfterTestWait:                       "cluster.afterTestWait",
	InstallTimeout:                      "cluster.installTimeout",
	UseLatestVersionForInstall:          "cluster.useLatestVersionForInstall",
	UseMiddleClusterImageSetForInstall:  "cluster.useMiddleClusterVersionForInstall",
	UseOldestClusterImageSetForInstall:  "cluster.useOldestClusterVersionForInstall",
	DeltaReleaseFromDefault:             "cluster.deltaReleaseFromDefault",
	NextReleaseAfterProdDefault:         "cluster.nextReleaseAfterProdDefault",
	CleanCheckRuns:                      "cluster.cleanCheckRuns",
	ID:                                  "cluster.id",
	Name:                                "cluster.name",
	Version:                             "cluster.version",
	EnoughVersionsForOldestOrMiddleTest: "cluster.enoughVersionForOldestOrMiddleTest",
	PreviousVersionFromDefaultFound:     "cluster.previousVersionFromDefaultFound",
	ProvisionShardID:                    "cluster.provisionshardID",
}

// CloudProvider config keys.
var CloudProvider = struct {
	// CloudProviderID is the cloud provider ID to use to provision the cluster.
	// Env: CLOUD_PROVIDER_ID
	CloudProviderID string

	// Region is the cloud provider region to use to provision the cluster.
	// Env: CLOUD_PROVIDER_REGION
	Region string
}{
	CloudProviderID: "cloudProvider.providerId",
	Region:          "cloudProvider.region",
}

// Addons config keys.
var Addons = struct {
	// IDsAtCreation is a comma separated list of IDs to create at cluster creation time.
	// Env: ADDON_IDS_AT_CREATION
	IDsAtCreation string

	// IDs is a comma separated list of IDs to install after a cluster is created.
	// Env: ADDON_IDS
	IDs string

	// TestHarnesses is a comma separated list of container images that will test the addon
	// Env: ADDON_TEST_HARNESSES
	TestHarnesses string

	// TestUser is the OpenShift user that the tests will run as
	// If "%s" is detected in the TestUser string, it will evaluate that as the project namespace
	// Example: "system:serviceaccount:%s:dedicated-admin"
	// Evaluated: "system:serviceaccount:osde2e-abc123:dedicated-admin"
	// Env: ADDON_TEST_USER
	TestUser string

	// RunCleanup is a boolean to specify whether the testHarnesses should have a separate
	// cleanup phase. This phase would run at the end of all e2e testing
	// Env: ADDON_RUN_CLEANUP
	RunCleanup string

	// CleanupHarnesses is a comma separated list of container images that will clean up any
	// artifacts created after test harnesses have run
	// Env: ADDON_CLEANUP_HARNESSES
	CleanupHarnesses string
}{
	IDsAtCreation:    "addons.idsAtCreation",
	IDs:              "addons.ids",
	TestHarnesses:    "addons.testHarnesses",
	TestUser:         "addons.testUser",
	RunCleanup:       "addons.runCleanup",
	CleanupHarnesses: "addons.cleanupHarnesses",
}

// Scale config keys.
var Scale = struct {
	// WorkloadsRepository is the git repository where the openshift-scale workloads are located.
	// Env: WORKLOADS_REPO
	WorkloadsRepository string

	// WorkloadsRepositoryBranch is the branch of the git repository to use.
	// Env: WORKLOADS_REPO_BRANCH
	WorkloadsRepositoryBranch string
}{
	WorkloadsRepository:       "scale.workloadsRepository",
	WorkloadsRepositoryBranch: "scale.workloadsRepositoryBranch",
}

// Prometheus config keys.
var Prometheus = struct {
	// Address is the address of the Prometheus instance to connect to.
	// Env: PROMETHEUS_ADDRESS
	Address string

	// BearerToken is the token needed for communicating with Prometheus.
	// Env: PROMETHEUS_BEARER_TOKEN
	BearerToken string
}{
	Address:     "prometheus.address",
	BearerToken: "prometheus.bearerToken",
}

// Alert config keys.
var Alert = struct {
	// SlackAPIToken is a bot slack token
	// Env: SLACK_API_TOKEN
	SlackAPIToken string
}{
	SlackAPIToken: "alert.slackAPIToken",
}

func init() {
	// Here's where we bind environment variables to config options and set defaults

	viper.SetConfigType("yaml") // Our configs are all in yaml.

	// ----- Top Level Configs -----
	viper.SetDefault(Provider, "ocm")
	viper.BindEnv(Provider, "PROVIDER")

	viper.BindEnv(JobName, "JOB_NAME")

	viper.SetDefault(JobID, -1)
	viper.BindEnv(JobID, "BUILD_NUMBER")

	viper.SetDefault(BaseJobURL, "https://storage.googleapis.com/origin-ci-test/logs")
	viper.BindEnv(BaseJobURL, "BASE_JOB_URL")

	viper.SetDefault(BaseProwURL, "https://deck-ci.apps.ci.l2s4.p1.openshiftapps.com")
	viper.BindEnv(BaseProwURL, "BASE_PROW_URL")

	// ARTIFACTS and REPORT_DIR are basically the same, but ARTIFACTS is used on prow.
	viper.BindEnv(Artifacts, "ARTIFACTS")

	viper.BindEnv(ReportDir, "REPORT_DIR")

	viper.BindEnv(Suffix, "SUFFIX")

	viper.SetDefault(DryRun, false)
	viper.BindEnv(DryRun, "DRY_RUN")

	viper.SetDefault(MustGather, true)
	viper.BindEnv(MustGather, "MUST_GATHER")

	// ----- Upgrade -----
	viper.SetDefault(Upgrade.UpgradeToCISIfPossible, false)
	viper.BindEnv(Upgrade.UpgradeToCISIfPossible, "UPGRADE_TO_CIS_IF_POSSIBLE")

	viper.SetDefault(Upgrade.OnlyUpgradeToZReleases, false)
	viper.BindEnv(Upgrade.OnlyUpgradeToZReleases, "ONLY_UPGRADE_TO_Z_RELEASES")

	viper.SetDefault(Upgrade.NextReleaseAfterProdDefaultForUpgrade, -1)
	viper.BindEnv(Upgrade.NextReleaseAfterProdDefaultForUpgrade, "NEXT_RELEASE_AFTER_PROD_DEFAULT_FOR_UPGRADE")

	viper.BindEnv(Upgrade.ReleaseStream, "UPGRADE_RELEASE_STREAM")

	viper.BindEnv(Upgrade.ReleaseName, "UPGRADE_RELEASE_NAME")

	viper.BindEnv(Upgrade.Image, "UPGRADE_IMAGE")

	viper.SetDefault(Upgrade.UpgradeVersionEqualToInstallVersion, false)

	viper.BindEnv(Upgrade.MonitorRoutesDuringUpgrade, "UPGRADE_MONITOR_ROUTES")
	viper.SetDefault(Upgrade.MonitorRoutesDuringUpgrade, true)

	viper.BindEnv(Upgrade.ManagedUpgrade, "UPGRADE_MANAGED")
	viper.SetDefault(Upgrade.ManagedUpgrade, true)

	viper.BindEnv(Upgrade.ManagedUpgradeTestPodDisruptionBudgets, "UPGRADE_MANAGED_TEST_PDBS")
	viper.SetDefault(Upgrade.ManagedUpgradeTestPodDisruptionBudgets, true)

	viper.BindEnv(Upgrade.ManagedUpgradeTestNodeDrain, "UPGRADE_MANAGED_TEST_DRAIN")
	viper.SetDefault(Upgrade.ManagedUpgradeTestNodeDrain, true)

	viper.BindEnv(Upgrade.WaitForWorkersToManagedUpgrade, "UPGRADE_WAIT_FOR_WORKERS")
	viper.SetDefault(Upgrade.WaitForWorkersToManagedUpgrade, true)

	// ----- Kubeconfig -----
	viper.BindEnv(Kubeconfig.Path, "TEST_KUBECONFIG")

	// ----- Tests -----
	viper.SetDefault(Tests.PollingTimeout, 30)
	viper.BindEnv(Tests.PollingTimeout, "POLLING_TIMEOUT")

	viper.BindEnv(Tests.GinkgoSkip, "GINKGO_SKIP")

	viper.BindEnv(Tests.GinkgoFocus, "GINKGO_FOCUS")

	viper.BindEnv(Tests.TestsToRun, "TESTS_TO_RUN")

	viper.SetDefault(Tests.SuppressSkipNotifications, true)
	viper.BindEnv(Tests.SuppressSkipNotifications, "SUPPRESS_SKIP_NOTIFICATIONS")

	viper.BindEnv(Tests.CleanRuns, "CLEAN_RUNS")

	viper.SetDefault(Tests.OperatorSkip, "insights")
	viper.BindEnv(Tests.OperatorSkip, "OPERATOR_SKIP")

	viper.SetDefault(Tests.SkipClusterHealthChecks, false)
	viper.BindEnv(Tests.OperatorSkip, "SKIP_CLUSTER_HEALTH_CHECKS")

	viper.SetDefault(Tests.MetricsBucket, "osde2e-metrics")
	viper.BindEnv(Tests.MetricsBucket, "METRICS_BUCKET")

	viper.BindEnv(Tests.ServiceAccount, "SERVICE_ACCOUNT")

	// ----- Cluster -----
	viper.SetDefault(Cluster.MultiAZ, false)
	viper.BindEnv(Cluster.MultiAZ, "MULTI_AZ")

	viper.SetDefault(Cluster.DestroyAfterTest, false)
	viper.BindEnv(Cluster.DestroyAfterTest, "DESTROY_CLUSTER")

	viper.SetDefault(Cluster.ExpiryInMinutes, 210)
	viper.BindEnv(Cluster.ExpiryInMinutes, "CLUSTER_EXPIRY_IN_MINUTES")

	viper.SetDefault(Cluster.AfterTestWait, 60)
	viper.BindEnv(Cluster.AfterTestWait, "AFTER_TEST_CLUSTER_WAIT")

	viper.SetDefault(Cluster.InstallTimeout, 135)
	viper.BindEnv(Cluster.InstallTimeout, "CLUSTER_UP_TIMEOUT")

	viper.SetDefault(Cluster.UseLatestVersionForInstall, false)
	viper.BindEnv(Cluster.UseLatestVersionForInstall, "USE_LATEST_VERSION_FOR_INSTALL")

	viper.SetDefault(Cluster.UseMiddleClusterImageSetForInstall, false)
	viper.BindEnv(Cluster.UseMiddleClusterImageSetForInstall, "USE_MIDDLE_CLUSTER_IMAGE_SET_FOR_INSTALL")

	viper.SetDefault(Cluster.UseOldestClusterImageSetForInstall, false)
	viper.BindEnv(Cluster.UseOldestClusterImageSetForInstall, "USE_OLDEST_CLUSTER_IMAGE_SET_FOR_INSTALL")

	viper.SetDefault(Cluster.DeltaReleaseFromDefault, 0)
	viper.BindEnv(Cluster.DeltaReleaseFromDefault, "DELTA_RELEASE_FROM_DEFAULT")

	viper.SetDefault(Cluster.NextReleaseAfterProdDefault, -1)
	viper.BindEnv(Cluster.NextReleaseAfterProdDefault, "NEXT_RELEASE_AFTER_PROD_DEFAULT")

	viper.SetDefault(Cluster.CleanCheckRuns, 20)
	viper.BindEnv(Cluster.CleanCheckRuns, "CLEAN_CHECK_RUNS")

	viper.BindEnv(Cluster.ID, "CLUSTER_ID")

	viper.BindEnv(Cluster.Name, "CLUSTER_NAME")

	viper.BindEnv(Cluster.Version, "CLUSTER_VERSION")

	viper.SetDefault(Cluster.EnoughVersionsForOldestOrMiddleTest, true)

	viper.SetDefault(Cluster.PreviousVersionFromDefaultFound, true)

	viper.SetDefault(Cluster.ProvisionShardID, "")
	viper.BindEnv(Cluster.ProvisionShardID, "PROVISION_SHARD_ID")

	// ----- Cloud Provider -----
	viper.SetDefault(CloudProvider.CloudProviderID, "aws")
	viper.BindEnv(CloudProvider.CloudProviderID, "CLOUD_PROVIDER_ID")

	viper.SetDefault(CloudProvider.Region, "us-east-1")
	viper.BindEnv(CloudProvider.Region, "CLOUD_PROVIDER_REGION")

	// ----- Addons -----
	viper.BindEnv(Addons.IDsAtCreation, "ADDON_IDS_AT_CREATION")

	viper.BindEnv(Addons.IDs, "ADDON_IDS")

	viper.BindEnv(Addons.TestHarnesses, "ADDON_TEST_HARNESSES")
	viper.BindEnv(Addons.CleanupHarnesses, "ADDON_CLEANUP_HARNESSES")

	viper.SetDefault(Addons.TestUser, "system:serviceaccount:%s:cluster-admin")
	viper.BindEnv(Addons.TestUser, "ADDON_TEST_USER")

	viper.SetDefault(Addons.RunCleanup, false)
	viper.BindEnv(Addons.RunCleanup, "ADDON_RUN_CLEANUP")

	// ----- Scale -----
	viper.SetDefault(Scale.WorkloadsRepository, "https://github.com/openshift-scale/workloads")
	viper.BindEnv(Scale.WorkloadsRepository, "WORKLOADS_REPO")

	viper.SetDefault(Scale.WorkloadsRepositoryBranch, "master")
	viper.BindEnv(Scale.WorkloadsRepositoryBranch, "WORKLOADS_REPO_BRANCH")

	// ----- Prometheus -----
	viper.BindEnv(Prometheus.Address, "PROMETHEUS_ADDRESS")

	viper.BindEnv(Prometheus.BearerToken, "PROMETHEUS_BEARER_TOKEN")

	// ----- Alert ----
	viper.BindEnv(Alert.SlackAPIToken, "SLACK_API_TOKEN")
}

// PostProcess is a variety of post-processing commands that is intended to be run after a config is loaded.
func PostProcess() {
	// Set REPORT_DIR to ARTIFACTS if ARTIFACTS is set.
	artifacts := viper.GetString(Artifacts)
	if artifacts != "" {
		log.Printf("Found an ARTIFACTS directory, using that for the REPORT_DIR.")
		viper.Set(ReportDir, artifacts)
	}
}

// RegisterSecret will register the secret filename that will be used for the corresponding Viper string.
func RegisterSecret(key string, secretFileName string) {
	keyToSecretMappingMutex.Lock()
	keyToSecretMapping[key] = secretFileName
	keyToSecretMappingMutex.Unlock()
}

// GetAllSecrets will return Viper config keys and their corresponding secret filenames.
func GetAllSecrets() map[string]string {
	return keyToSecretMapping
}
