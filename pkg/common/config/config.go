// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

const (
	// Provider is what provider to use to create/delete clusters.
	Provider = "provider"

	// JobName lets you name the current e2e job run
	JobName = "jobName"

	// JobID is the ID designated by prow for this run
	JobID = "jobID"

	// BaseJobURL is the root location for all job artifacts
	// For example, https://storage.googleapis.com/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/61/build-log.txt would be
	// https://storage.googleapis.com/origin-ci-test/logs -- This is also our default
	BaseJobURL = "baseJobURL"

	// BaseProwURL is the root location of Prow
	BaseProwURL = "baseProwURL"

	// Artifacts is the artifacts location on prow. It is an alias for report dir.
	Artifacts = "artifacts"

	// ReportDir is the location JUnit XML results are written.
	ReportDir = "reportDir"

	// Suffix is used at the end of test names to identify them.
	Suffix = "suffix"

	// DryRun lets you run osde2e all the way up to the e2e tests then skips them.
	DryRun = "dryRun"

	// MustGather will run a Must-Gather process upon completion of the tests.
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
	UpgradeToCISIfPossible string

	// OnlyUpgradeToZReleases will restrict upgrades to selecting Z releases on stage/prod.
	OnlyUpgradeToZReleases string

	// NextReleaseAfterProdDefaultForUpgrade will select the cluster image set that the given number of releases away from the the production default.
	NextReleaseAfterProdDefaultForUpgrade string

	// ReleaseStream used to retrieve latest release images. If set, it will be used to perform an upgrade.
	ReleaseStream string

	// ReleaseName is the name of the release in a release stream.
	ReleaseName string

	// Image is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.
	Image string

	// UpgradeVersionEqualToInstallVersion is true if the install version and upgrade versions are the same.
	UpgradeVersionEqualToInstallVersion string

	// MonitorRoutesDuringUpgrade will monitor the availability of routes whilst an upgrade takes place
	MonitorRoutesDuringUpgrade string
}{
	UpgradeToCISIfPossible:                "upgrade.upgradeToCISIfPossible",
	OnlyUpgradeToZReleases:                "upgrade.onlyUpgradeToZReleases",
	NextReleaseAfterProdDefaultForUpgrade: "upgrade.nextReleaseAfterProdDefaultForUpgrade",
	ReleaseStream:                         "upgrade.releaseStream",
	ReleaseName:                           "upgrade.releaseName",
	Image:                                 "upgrade.image",
	UpgradeVersionEqualToInstallVersion:   "upgrade.upgradeVersionEqualToInstallVersion",
	MonitorRoutesDuringUpgrade:            "upgrade.monitorRoutesDuringUpgrade",
}

// Kubeconfig config keys.
var Kubeconfig = struct {
	// Path is the filepath of an existing Kubeconfig
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
	// PollingTimeout is how long (in mimutes) to wait for an object to be created before failing the test.
	PollingTimeout string

	// GinkgoSkip is a regex passed to Ginkgo that skips any test suites matching the regex. ex. "Operator"
	GinkgoSkip string

	// GinkgoFocus is a regex passed to Ginkgo that focus on any test suites matching the regex. ex. "Operator"
	GinkgoFocus string

	// TestsToRun is a list of files which should be executed as part of a test suite
	TestsToRun string

	// SuppressSkipNotifications suppresses the notifications of skipped tests
	SuppressSkipNotifications string

	// CleanRuns is the number of times the test-version is run before skipping.
	CleanRuns string

	// OperatorSkip is a comma-delimited list of operator names to ignore health checks from. ex. "insights,telemetry"
	OperatorSkip string

	// SkipClusterHealthChecks skips the cluster health checks. Useful when developing against a running cluster.
	SkipClusterHealthChecks string

	// MetricsBucket is the bucket that metrics data will be uploaded to.
	MetricsBucket string

	// ServiceAccount defines what user the tests should run as. By default, osde2e uses system:admin
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
	MultiAZ string

	// DestroyClusterAfterTest set to true if you want to the cluster to be explicitly deleted after the test.
	DestroyAfterTest string

	// ExpiryInMinutes is how long before a cluster expires and is deleted by OSD.
	ExpiryInMinutes string

	// AfterTestWait is how long to keep a cluster around after tests have run.
	AfterTestWait string

	// InstallTimeout is how long to wait before failing a cluster launch.
	InstallTimeout string

	// UseLatestVersionForInstall will select the latest cluster image set available for a fresh install.
	UseLatestVersionForInstall string

	// UseMiddleClusterImageSetForInstall will select the cluster image set that is in the middle of the list of ordered cluster versions known to OCM.
	UseMiddleClusterImageSetForInstall string

	// UseOldestClusterImageSetForInstall will select the cluster image set that is in the end of the list of ordered cluster versions known to OCM.
	UseOldestClusterImageSetForInstall string

	// DeltaReleaseFromDefault will select the cluster image set that is the given number of releases from the current default in either direction.
	DeltaReleaseFromDefault string

	// NextReleaseAfterProdDefault will select the cluster image set that the given number of releases away from the the production default.
	NextReleaseAfterProdDefault string

	// CleanCheckRuns lets us set the number of osd-verify checks we want to run before deeming a cluster "healthy"
	CleanCheckRuns string

	// ID identifies the cluster. If set at start, an existing cluster is tested.
	ID string

	// Name is the name of the cluster being created.
	Name string

	// Version is the version of the cluster being deployed.
	Version string

	// EnoughVersionsForOldestOrMiddleTest is true if there were enough versions for an older/middle test.
	EnoughVersionsForOldestOrMiddleTest string

	// PreviousVersionFromDefaultFound is true if a previous version from default was found.
	PreviousVersionFromDefaultFound string
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
}

// CloudProvider config keys.
var CloudProvider = struct {
	// CloudProviderID is the cloud provider ID to use to provision the cluster.
	CloudProviderID string

	// Region is the cloud provider region to use to provision the cluster.
	Region string
}{
	CloudProviderID: "cloudProvider.providerId",
	Region:          "cloudProvider.region",
}

// Addons config keys.
var Addons = struct {
	// IDsAtCreation is a comma separated list of IDs to create at cluster creation time.
	IDsAtCreation string

	// IDs is a comma separated list of IDs to install after a cluster is created.
	IDs string

	// TestHarnesses is a comma separated list of container images that will test the addon
	TestHarnesses string

	// TestUser is the OpenShift user that the tests will run as
	// If "%s" is detected in the TestUser string, it will evaluate that as the project namespace
	// Example: "system:serviceaccount:%s:dedicated-admin"
	// Evaluated: "system:serviceaccount:osde2e-abc123:dedicated-admin"
	TestUser string
}{
	IDsAtCreation: "addons.idsAtCreation",
	IDs:           "ids",
	TestHarnesses: "testHarnesses",
	TestUser:      "addons.testUser",
}

// Scale config keys.
var Scale = struct {
	// WorkloadsRepository is the git repository where the openshift-scale workloads are located.
	WorkloadsRepository string

	// WorkloadsRepositoryBranch is the branch of the git repository to use.
	WorkloadsRepositoryBranch string
}{
	WorkloadsRepository:       "scale.workloadsRepository",
	WorkloadsRepositoryBranch: "scale.workloadsRepositoryBranch",
}

// Prometheus config keys.
var Prometheus = struct {
	// Address is the address of the Prometheus instance to connect to.
	Address string

	// BearerToken is the token needed for communicating with Prometheus.
	BearerToken string
}{
	Address:     "prometheus.address",
	BearerToken: "prometheus.bearerToken",
}

// Weather config keys.
var Weather = struct {
	// StartOfTimeWindowInHours is how many hours to look back through results.
	StartOfTimeWindowInHours string

	// NumberOfSamplesNecessary is how many samples are necessary for generating a report.
	NumberOfSamplesNecessary string

	// SlackWebhook is the webhook to use to post the weather report to slack.
	SlackWebhook string

	// JobAllowlist is a list of job regexes to consider in the weather report.
	JobAllowlist string
}{
	StartOfTimeWindowInHours: "weather.startOfTimeWindowInHours",
	NumberOfSamplesNecessary: "weather.numberOfSamplesNecessary",
	SlackWebhook:             "weather.slackWebhook",
	JobAllowlist:             "weather.jobAllowlist",
}

var Alert = struct {
	// SlackAPIToken is a bot slack token
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

	// ----- Cloud Provider -----
	viper.SetDefault(CloudProvider.CloudProviderID, "aws")
	viper.BindEnv(CloudProvider.CloudProviderID, "CLOUD_PROVIDER_ID")

	viper.SetDefault(CloudProvider.Region, "us-east-1")
	viper.BindEnv(CloudProvider.Region, "CLOUD_PROVIDER_REGION")

	// ----- Addons -----
	viper.BindEnv(Addons.IDsAtCreation, "ADDON_IDS_AT_CREATION")

	viper.BindEnv(Addons.IDs, "ADDON_IDS")

	viper.BindEnv(Addons.TestHarnesses, "ADDON_TEST_HARNESSES")

	viper.BindEnv(Addons.TestUser, "ADDON_TEST_USER")
	viper.SetDefault(Addons.TestUser, "system:serviceaccount:%s:cluster-admin")

	// ----- Scale -----
	viper.SetDefault(Scale.WorkloadsRepository, "https://github.com/openshift-scale/workloads")
	viper.BindEnv(Scale.WorkloadsRepository, "WORKLOADS_REPO")

	viper.SetDefault(Scale.WorkloadsRepositoryBranch, "master")
	viper.BindEnv(Scale.WorkloadsRepositoryBranch, "WORKLOADS_REPO_BRANCH")

	// ----- Prometheus -----
	viper.BindEnv(Prometheus.Address, "PROMETHEUS_ADDRESS")

	viper.BindEnv(Prometheus.BearerToken, "PROMETHEUS_BEARER_TOKEN")

	// ----- Weather -----
	viper.SetDefault(Weather.StartOfTimeWindowInHours, 24)
	viper.BindEnv(Weather.StartOfTimeWindowInHours, "START_OF_TIME_WINDOW_IN_HOURS")

	viper.SetDefault(Weather.NumberOfSamplesNecessary, 3)
	viper.BindEnv(Weather.NumberOfSamplesNecessary, "NUMBER_OF_SAMPLES_NECESSARY")

	viper.BindEnv(Weather.SlackWebhook, "SLACK_WEBHOOK")

	viper.SetDefault(Weather.JobAllowlist, "osde2e-.*-aws-e2e-.*")
	viper.BindEnv(Weather.JobAllowlist, "JOB_ALLOWLIST")

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
