// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

type Secret struct {
	FileLocation string
	Key          string
}

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

	// JobType is the type of job according to prow for this run
	// Env: JOB_TYPE
	JobType = "jobType"

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

	// CanaryChance
	CanaryChance = "canaryChance"

	// Default network provider for OSD
	DefaultNetworkProvider = "OVNKubernetes"

	// NonOSDe2eSecrets is an internal-only Viper Key.
	// End users should not be using this key, there may be unforeseen consequences.
	NonOSDe2eSecrets = "nonOSDe2eSecrets"

	// JobStartedAt tracks when the job began running.
	JobStartedAt = "JobStartedAt"
)

// This is a config key to secret file mapping. We will attempt to read in from secret files before loading anything else.

var (
	keyToSecretMapping      = []Secret{}
	keyToSecretMappingMutex = sync.Mutex{}
)

// This is a list of OSD-specific namespaces to include in the post-E2E cleanup must-gather
// that takes place.
var defaultInspectNamespaces = []string{
	"openshift-managed-upgrade-operator",
	"openshift-velero",
	"openshift-build-test",
	"openshift-sre-pruning",
	"openshift-cloud-ingress-operator",
	"openshift-rbac-permissions",
	"openshift-route-monitor-operator",
	"openshift-validation-webhook",
	"openshift-backplane",
	"openshift-custom-domains-operator",
	"openshift-must-gather-operator",
	"openshift-splunk-forwarder-operator",
	"openshift-rbac-permissions",
}

// Upgrade config keys.
var Upgrade = struct {
	// UpgradeToLatest will look for the newest-possible version and select that
	// Env: UPGRADE_TO_LATEST
	UpgradeToLatest string

	// UpgradeToLatestY will look for the latest Y version for the cluster and select that
	// Env: UPGRADE_TO_LATEST_Y
	UpgradeToLatestY string

	// UpgradeToLatestZ will look for the latest Z version for the cluster and select that
	// Env: UPGRADE_TO_LATEST_Z
	UpgradeToLatestZ string

	// ReleaseName is the name of the release in a release stream.
	// Env: UPGRADE_RELEASE_NAME
	ReleaseName string

	// Image is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.
	// Env: UPGRADE_IMAGE
	Image string

	// Type of upgrader to use when upgrading (OSD or ARO)
	// ENV: UPGRADE_TYPE
	Type string

	// UpgradeVersionEqualToInstallVersion is true if the install version and upgrade versions are the same.
	UpgradeVersionEqualToInstallVersion string

	// MonitorRoutesDuringUpgrade will monitor the availability of routes whilst an upgrade takes place
	// Env: UPGRADE_MONITOR_ROUTES
	MonitorRoutesDuringUpgrade string

	// Create disruptive Pod Disruption Budget workloads to test the Managed Upgrade Operator's ability to handle them.
	ManagedUpgradeTestPodDisruptionBudgets string

	// Create disruptive Node Drain workload to test the Managed Upgrade Operator's ability to handle them.
	ManagedUpgradeTestNodeDrain string

	// Reschedule the upgrade via provider before commence
	ManagedUpgradeRescheduled string
}{
	UpgradeToLatest:                        "upgrade.toLatest",
	UpgradeToLatestZ:                       "upgrade.ToLatestZ",
	UpgradeToLatestY:                       "upgrade.ToLatestY",
	ReleaseName:                            "upgrade.releaseName",
	Image:                                  "upgrade.image",
	Type:                                   "upgrade.type",
	UpgradeVersionEqualToInstallVersion:    "upgrade.upgradeVersionEqualToInstallVersion",
	MonitorRoutesDuringUpgrade:             "upgrade.monitorRoutesDuringUpgrade",
	ManagedUpgradeTestPodDisruptionBudgets: "upgrade.managedUpgradeTestPodDisruptionBudgets",
	ManagedUpgradeTestNodeDrain:            "upgrade.managedUpgradeTestNodeDrain",
	ManagedUpgradeRescheduled:              "upgrade.managedUpgradeRescheduled",
}

// Kubeconfig configBUILD_NUMBER keys.
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

	// GinkgoLogLevel contrls the logging level used by ginkgo when providing test output
	// Env: GINKGO_LOG_LEVEL
	GinkgoLogLevel string

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

	// ClusterHealthChecksTimeout defines the duration for which the harness will
	// wait for the cluster to indicate it is healthy before cancelling the test
	// run. This value should be formatted for use with time.ParseDuration.
	// Env: CLUSTER_HEALTH_CHECKS_TIMEOUT
	ClusterHealthChecksTimeout string

	// MetricsBucket is the bucket that metrics data will be uploaded to.
	// Env: METRICS_BUCKET
	MetricsBucket string

	// ServiceAccount defines what user the tests should run as. By default, osde2e uses system:admin
	// Env: SERVICE_ACCOUNT
	ServiceAccount string

	// EnableFips enables the FIPS test suite
	// Env: ENABLE_FIPS
	EnableFips string
}{
	PollingTimeout:             "tests.pollingTimeout",
	GinkgoSkip:                 "tests.ginkgoSkip",
	GinkgoFocus:                "tests.focus",
	GinkgoLogLevel:             "tests.ginkgoLogLevel",
	TestsToRun:                 "tests.testsToRun",
	SuppressSkipNotifications:  "tests.suppressSkipNotifications",
	CleanRuns:                  "tests.cleanRuns",
	OperatorSkip:               "tests.operatorSkip",
	SkipClusterHealthChecks:    "tests.skipClusterHealthChecks",
	MetricsBucket:              "tests.metricsBucket",
	ServiceAccount:             "tests.serviceAccount",
	ClusterHealthChecksTimeout: "tests.clusterHealthChecksTimeout",
	EnableFips:                 "tests.enableFips",
}

// Cluster config keys.
var Cluster = struct {
	// MultiAZ deploys a cluster across multiple availability zones.
	// Env: MULTI_AZ
	MultiAZ string

	// Channel dictates which install/upgrade edges will be available to the cluster
	// Env: CHANNEL
	Channel string

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

	// ReleaseImageLatest is used when we're testing versions not-yet-accepted from the release controller.
	ReleaseImageLatest string

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

	// LatestYReleaseAfterProdDefault will select the next minor version CIS for an environment given the production default
	LatestYReleaseAfterProdDefault string

	// LatestZReleaseAfterProdDefault will select the next patch version CIS for an environment given the production default
	LatestZReleaseAfterProdDefault string

	// InstallSpecificNightly will select a nightly using a specific nightly given an "X.Y" formatted string
	InstallSpecificNightly string

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

	// NumWorkerNodes overrides the flavour's number of worker nodes specified
	NumWorkerNodes string

	// NetworkProvider chooses the network driver powering the cluster.
	NetworkProvider string

	// Specify a key in the pre-defined imageContentSource array in the ocmprovider
	// Blank will default to a randomized option
	ImageContentSource string

	// InstallConfig overrides merges on top of the installer's default OCP installer config
	// Blank will do nothing
	// Cannot specify imageContentSources within this config
	InstallConfig string

	// HibernateAfterUse will tell the provider to attempt to hibernate the cluster after
	// the test run, assuming the provider supports hibernation
	HibernateAfterUse string

	// UseExistingCluster will allow the test run to use an existing cluster if available
	// ENV: USE_EXISTING_CLUSTER
	// Default: True
	UseExistingCluster string

	// Passing tracks the internal status of the tests: Pass or Fail
	Passing string

	// Reused tracks whether this cluster's test run used a new or recycled cluster
	Reused string

	// InspectNamespaces is a comma-delimited list of namespaces to perform an inspect on during test cleanup
	InspectNamespaces string

	// UseProxyForInstall will attempt to use a cluster-wide proxy for cluster installation, provided that a cluster-wide proxy config is supplied
	UseProxyForInstall string
}{
	MultiAZ:                             "cluster.multiAZ",
	Channel:                             "cluster.channel",
	DestroyAfterTest:                    "cluster.destroyAfterTest",
	ExpiryInMinutes:                     "cluster.expiryInMinutes",
	AfterTestWait:                       "cluster.afterTestWait",
	InstallTimeout:                      "cluster.installTimeout",
	ReleaseImageLatest:                  "cluster.releaseImageLatest",
	UseProxyForInstall:                  "cluster.useProxyForInstall",
	UseLatestVersionForInstall:          "cluster.useLatestVersionForInstall",
	UseMiddleClusterImageSetForInstall:  "cluster.useMiddleClusterVersionForInstall",
	UseOldestClusterImageSetForInstall:  "cluster.useOldestClusterVersionForInstall",
	DeltaReleaseFromDefault:             "cluster.deltaReleaseFromDefault",
	NextReleaseAfterProdDefault:         "cluster.nextReleaseAfterProdDefault",
	LatestYReleaseAfterProdDefault:      "cluster.latestYReleaseAfterProdDefault",
	LatestZReleaseAfterProdDefault:      "cluster.latestZReleaseAfterProdDefault",
	InstallSpecificNightly:              "cluster.installLatestNightly",
	CleanCheckRuns:                      "cluster.cleanCheckRuns",
	ID:                                  "cluster.id",
	Name:                                "cluster.name",
	Version:                             "cluster.version",
	EnoughVersionsForOldestOrMiddleTest: "cluster.enoughVersionForOldestOrMiddleTest",
	PreviousVersionFromDefaultFound:     "cluster.previousVersionFromDefaultFound",
	ProvisionShardID:                    "cluster.provisionshardID",
	NumWorkerNodes:                      "cluster.numWorkerNodes",
	NetworkProvider:                     "cluster.networkProvider",
	ImageContentSource:                  "cluster.imageContentSource",
	InstallConfig:                       "cluster.installConfig",
	HibernateAfterUse:                   "cluster.hibernateAfterUse",
	UseExistingCluster:                  "cluster.useExistingCluster",
	Passing:                             "cluster.passing",
	Reused:                              "cluster.rused",
	InspectNamespaces:                   "cluster.inspectNamespaces",
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

	// SlackChannel is the name of a slack channel in the CoreOS slack workspace that will
	// receive an alert if the tests fail.
	// Env: ADDON_SLACK_CHANNEL
	SlackChannel string

	// Parameters is a nested json object. Top-level keys should be addon
	// IDs provided in the IDs field. The values should be objects with
	// string key-value pairs of parameters to provide to the addon with
	// the associated top-level ID.
	// An example:
	// {"AddonA": {"paramName":"paramValue"}, "AddonB": {"paramName": "paramValue"}}
	// Env: ADDON_PARAMETERS
	Parameters string

	// SkipAddonList is a boolean to indicate whether the listing of addons has to be disabled or not.
	// Env: SKIP_ADDON_LIST
	SkipAddonList string

	// PollingTimeout is how long (in seconds) to wait for the add-on test to complete running.
	// Env: ADDON_POLLING_TIMEOUT
	PollingTimeout string
}{
	IDsAtCreation:    "addons.idsAtCreation",
	IDs:              "addons.ids",
	TestHarnesses:    "addons.testHarnesses",
	TestUser:         "addons.testUser",
	RunCleanup:       "addons.runCleanup",
	CleanupHarnesses: "addons.cleanupHarnesses",
	SlackChannel:     "addons.slackChannel",
	SkipAddonList:    "addons.skipAddonlist",
	Parameters:       "addons.parameters",
	PollingTimeout:   "addons.pollingTimeout",
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
	// EnableAlerts is a boolean to indicate whether alerts should be enabled or not.
	// Env: ENABLE_ALERTS
	EnableAlerts string

	// SlackAPIToken is a bot slack token
	// Env: SLACK_API_TOKEN
	SlackAPIToken string

	// PagerDutyAPIToken is a pagerduty token
	// Env: PAGERDUTY_API_TOKEN
	PagerDutyAPIToken string

	// PagerDutyUserToken is a pagerduty token for a user account with full access to the v2 API
	// Env: PAGERDUTY_API_TOKEN
	PagerDutyUserToken string
}{
	EnableAlerts:       "alert.EnableAlerts",
	SlackAPIToken:      "alert.slackAPIToken",
	PagerDutyAPIToken:  "alert.pagerDutyAPIToken",
	PagerDutyUserToken: "alert.pagerDutyUserToken",
}

// Database config keys.
var Database = struct {
	// The Postgres user used to access the database.
	// Env: PG_USER
	User string
	// The Postgres password for the user.
	// Env: PG_PASS
	Pass string
	// The Postgres instance's hostname.
	// Env: PG_HOST
	Host string
	// The Postgres instance's listen port.
	// Env: PG_PORT
	Port string
	// The Postgres database name to connect to.
	// Env: PG_DATABASE
	DatabaseName string
}{
	User:         "database.user",
	Pass:         "database.pass",
	Host:         "database.host",
	Port:         "database.port",
	DatabaseName: "database.name",
}

// Proxy config keys
var Proxy = struct {
	// The HTTPS Proxy address to use for proxy tests,
	HttpsProxy string
	// The HTTP Proxy address to use for proxy tests.
	HttpProxy string
	// The User CA Bundle to use for proxy tests.
	UserCABundle string
}{
	HttpsProxy:   "proxy.https_proxy",
	HttpProxy:    "proxy.http_proxy",
	UserCABundle: "proxy.user_ca_bundle",
}

func InitViper() {
	// Here's where we bind environment variables to config options and set defaults

	viper.SetConfigType("yaml") // Our configs are all in yaml.

	// capture job startup time
	viper.SetDefault(JobStartedAt, time.Now().UTC().Format(time.RFC3339))

	// ----- Top Level Configs -----
	viper.SetDefault(Provider, "ocm")
	viper.BindEnv(Provider, "PROVIDER")

	viper.BindEnv(JobName, "JOB_NAME")
	viper.BindEnv(JobType, "JOB_TYPE")

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

	viper.BindEnv(CanaryChance, "CANARY_CHANCE")

	// ----- Upgrade -----
	viper.BindEnv(Upgrade.UpgradeToLatest, "UPGRADE_TO_LATEST")
	viper.SetDefault(Upgrade.UpgradeToLatest, false)

	viper.BindEnv(Upgrade.UpgradeToLatestZ, "UPGRADE_TO_LATEST_Z")
	viper.SetDefault(Upgrade.UpgradeToLatestZ, false)

	viper.BindEnv(Upgrade.UpgradeToLatestY, "UPGRADE_TO_LATEST_Y")
	viper.SetDefault(Upgrade.UpgradeToLatestY, false)

	viper.BindEnv(Upgrade.ReleaseName, "UPGRADE_RELEASE_NAME")

	viper.BindEnv(Upgrade.Image, "UPGRADE_IMAGE")

	viper.SetDefault(Upgrade.Type, "OSD")
	viper.BindEnv(Upgrade.Type, "UPGRADE_TYPE")

	viper.SetDefault(Upgrade.UpgradeVersionEqualToInstallVersion, false)

	viper.BindEnv(Upgrade.MonitorRoutesDuringUpgrade, "UPGRADE_MONITOR_ROUTES")
	viper.SetDefault(Upgrade.MonitorRoutesDuringUpgrade, true)

	viper.BindEnv(Upgrade.ManagedUpgradeTestPodDisruptionBudgets, "UPGRADE_MANAGED_TEST_PDBS")
	viper.SetDefault(Upgrade.ManagedUpgradeTestPodDisruptionBudgets, true)

	viper.BindEnv(Upgrade.ManagedUpgradeTestNodeDrain, "UPGRADE_MANAGED_TEST_DRAIN")
	viper.SetDefault(Upgrade.ManagedUpgradeTestNodeDrain, true)

	viper.BindEnv(Upgrade.ManagedUpgradeRescheduled, "UPGRADE_MANAGED_TEST_RESCHEDULE")
	viper.SetDefault(Upgrade.ManagedUpgradeRescheduled, false)

	// ----- Kubeconfig -----
	viper.BindEnv(Kubeconfig.Path, "TEST_KUBECONFIG")

	// ----- Tests -----
	viper.SetDefault(Tests.PollingTimeout, 500)
	viper.BindEnv(Tests.PollingTimeout, "POLLING_TIMEOUT")

	viper.BindEnv(Tests.GinkgoSkip, "GINKGO_SKIP")

	viper.BindEnv(Tests.GinkgoFocus, "GINKGO_FOCUS")

	viper.BindEnv(Tests.GinkgoLogLevel, "GINKGO_LOG_LEVEL")

	viper.BindEnv(Tests.TestsToRun, "TESTS_TO_RUN")

	viper.SetDefault(Tests.SuppressSkipNotifications, true)
	viper.BindEnv(Tests.SuppressSkipNotifications, "SUPPRESS_SKIP_NOTIFICATIONS")

	viper.BindEnv(Tests.CleanRuns, "CLEAN_RUNS")

	viper.SetDefault(Tests.OperatorSkip, "insights")
	viper.BindEnv(Tests.OperatorSkip, "OPERATOR_SKIP")

	viper.SetDefault(Tests.SkipClusterHealthChecks, false)
	viper.BindEnv(Tests.OperatorSkip, "SKIP_CLUSTER_HEALTH_CHECKS")

	viper.SetDefault(Tests.ClusterHealthChecksTimeout, "2h")
	viper.BindEnv(Tests.ClusterHealthChecksTimeout, "CLUSTER_HEALTH_CHECKS_TIMEOUT")

	viper.SetDefault(Tests.MetricsBucket, "osde2e-metrics")
	viper.BindEnv(Tests.MetricsBucket, "METRICS_BUCKET")

	viper.BindEnv(Tests.ServiceAccount, "SERVICE_ACCOUNT")

	viper.SetDefault(Tests.EnableFips, false)
	viper.BindEnv(Tests.EnableFips, "ENABLE_FIPS")

	// ----- Cluster -----
	viper.SetDefault(Cluster.MultiAZ, false)
	viper.BindEnv(Cluster.MultiAZ, "MULTI_AZ")

	viper.SetDefault(Cluster.Channel, "candidate")
	viper.BindEnv(Cluster.Channel, "CHANNEL")

	viper.SetDefault(Cluster.DestroyAfterTest, true)
	viper.BindEnv(Cluster.DestroyAfterTest, "DESTROY_CLUSTER")

	viper.SetDefault(Cluster.ExpiryInMinutes, 360)
	viper.BindEnv(Cluster.ExpiryInMinutes, "CLUSTER_EXPIRY_IN_MINUTES")

	viper.SetDefault(Cluster.AfterTestWait, 60)
	viper.BindEnv(Cluster.AfterTestWait, "AFTER_TEST_CLUSTER_WAIT")

	viper.SetDefault(Cluster.InstallTimeout, 135)
	viper.BindEnv(Cluster.InstallTimeout, "CLUSTER_UP_TIMEOUT")

	viper.BindEnv(Cluster.ReleaseImageLatest, "RELEASE_IMAGE_LATEST")

	viper.SetDefault(Cluster.UseProxyForInstall, false)
	viper.BindEnv(Cluster.UseProxyForInstall, "USE_PROXY_FOR_INSTALL")

	viper.SetDefault(Cluster.UseLatestVersionForInstall, false)
	viper.BindEnv(Cluster.UseLatestVersionForInstall, "USE_LATEST_VERSION_FOR_INSTALL")

	viper.SetDefault(Cluster.UseMiddleClusterImageSetForInstall, false)
	viper.BindEnv(Cluster.UseMiddleClusterImageSetForInstall, "USE_MIDDLE_CLUSTER_IMAGE_SET_FOR_INSTALL")

	viper.SetDefault(Cluster.UseOldestClusterImageSetForInstall, false)
	viper.BindEnv(Cluster.UseOldestClusterImageSetForInstall, "USE_OLDEST_CLUSTER_IMAGE_SET_FOR_INSTALL")

	viper.SetDefault(Cluster.LatestYReleaseAfterProdDefault, false)
	viper.BindEnv(Cluster.LatestYReleaseAfterProdDefault, "LATEST_Y_RELEASE_AFTER_PROD_DEFAULT")

	viper.SetDefault(Cluster.LatestZReleaseAfterProdDefault, false)
	viper.BindEnv(Cluster.LatestZReleaseAfterProdDefault, "LATEST_Z_RELEASE_AFTER_PROD_DEFAULT")

	viper.BindEnv(Cluster.InstallSpecificNightly, "INSTALL_LATEST_NIGHTLY")

	viper.SetDefault(Cluster.DeltaReleaseFromDefault, 0)
	viper.BindEnv(Cluster.DeltaReleaseFromDefault, "DELTA_RELEASE_FROM_DEFAULT")

	viper.SetDefault(Cluster.NextReleaseAfterProdDefault, -1)
	viper.BindEnv(Cluster.NextReleaseAfterProdDefault, "NEXT_RELEASE_AFTER_PROD_DEFAULT")

	viper.SetDefault(Cluster.CleanCheckRuns, 20)
	viper.BindEnv(Cluster.CleanCheckRuns, "CLEAN_CHECK_RUNS")

	viper.SetDefault(Cluster.ID, "")
	viper.BindEnv(Cluster.ID, "CLUSTER_ID")

	viper.SetDefault(Cluster.Name, "")
	viper.BindEnv(Cluster.Name, "CLUSTER_NAME")

	viper.SetDefault(Cluster.Version, "")
	viper.BindEnv(Cluster.Version, "CLUSTER_VERSION")

	viper.SetDefault(Cluster.EnoughVersionsForOldestOrMiddleTest, true)

	viper.SetDefault(Cluster.PreviousVersionFromDefaultFound, true)

	viper.SetDefault(Cluster.ProvisionShardID, "")
	viper.BindEnv(Cluster.ProvisionShardID, "PROVISION_SHARD_ID")

	viper.SetDefault(Cluster.NumWorkerNodes, "")
	viper.BindEnv(Cluster.NumWorkerNodes, "NUM_WORKER_NODES")

	viper.BindEnv(Cluster.ImageContentSource, "CLUSTER_IMAGE_CONTENT_SOURCE")
	viper.BindEnv(Cluster.InstallConfig, "CLUSTER_INSTALL_CONFIG")

	viper.SetDefault(Cluster.NetworkProvider, DefaultNetworkProvider)
	viper.BindEnv(Cluster.NetworkProvider, "CLUSTER_NETWORK_PROVIDER")

	viper.SetDefault(Cluster.HibernateAfterUse, true)
	viper.BindEnv(Cluster.HibernateAfterUse, "HIBERNATE_AFTER_USE")

	viper.SetDefault(Cluster.UseExistingCluster, false)
	viper.BindEnv(Cluster.UseExistingCluster, "USE_EXISTING_CLUSTER")

	viper.SetDefault(Cluster.Reused, false)
	viper.SetDefault(Cluster.Passing, false)

	viper.SetDefault(Cluster.InspectNamespaces, strings.Join(defaultInspectNamespaces, ","))
	viper.BindEnv(Cluster.InspectNamespaces, "INSPECT_NAMESPACES")

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

	viper.SetDefault(Addons.SlackChannel, "sd-cicd-alerts")
	viper.BindEnv(Addons.SlackChannel, "ADDON_SLACK_CHANNEL")

	viper.SetDefault(Addons.Parameters, "{}")
	viper.BindEnv(Addons.Parameters, "ADDON_PARAMETERS")
	RegisterSecret(Addons.Parameters, "addon-parameters")

	viper.SetDefault(Addons.SkipAddonList, false)
	viper.BindEnv(Addons.SkipAddonList, "SKIP_ADDON_LIST")

	viper.SetDefault(Addons.PollingTimeout, 3600)
	viper.BindEnv(Addons.PollingTimeout, "ADDON_POLLING_TIMEOUT")

	// ----- Scale -----
	viper.SetDefault(Scale.WorkloadsRepository, "https://github.com/openshift-scale/workloads")
	viper.BindEnv(Scale.WorkloadsRepository, "WORKLOADS_REPO")

	viper.SetDefault(Scale.WorkloadsRepositoryBranch, "master")
	viper.BindEnv(Scale.WorkloadsRepositoryBranch, "WORKLOADS_REPO_BRANCH")

	// ----- Prometheus -----
	viper.BindEnv(Prometheus.Address, "PROMETHEUS_ADDRESS")

	viper.BindEnv(Prometheus.BearerToken, "PROMETHEUS_BEARER_TOKEN")

	// ----- Alert ----
	viper.BindEnv(Alert.EnableAlerts, "ENABLE_ALERTS")
	viper.SetDefault(Alert.EnableAlerts, false)

	viper.BindEnv(Alert.SlackAPIToken, "SLACK_API_TOKEN")
	RegisterSecret(Alert.SlackAPIToken, "slack-api-token")

	// Support Legacy ENV Reference
	viper.BindEnv(Alert.PagerDutyAPIToken, "PAGERDUTY_API_TOKEN", "PAGERDUTY_TOKEN")
	RegisterSecret(Alert.PagerDutyAPIToken, "pagerduty-api-token")

	viper.BindEnv(Alert.PagerDutyUserToken, "PAGERDUTY_USER_TOKEN")
	RegisterSecret(Alert.PagerDutyUserToken, "pagerduty-user-token")

	// ----- Database -----
	viper.SetDefault(Database.User, "postgres")
	viper.BindEnv(Database.User, "PG_USER")
	RegisterSecret(Database.User, "rds-user")

	viper.BindEnv(Database.Pass, "PG_PASS")

	RegisterSecret(Database.Pass, "rds-pass")

	viper.BindEnv(Database.Host, "PG_HOST")
	RegisterSecret(Database.Host, "rds-host")

	viper.SetDefault(Database.Port, "5432")
	viper.BindEnv(Database.Port, "PG_PORT")

	viper.SetDefault(Database.DatabaseName, "cicd_test_data")
	viper.BindEnv(Database.DatabaseName, "PG_DATABASE")
	RegisterSecret(Database.DatabaseName, "rds-database")

	// ----- Proxy ------
	viper.BindEnv(Proxy.HttpProxy, "TEST_HTTP_PROXY")
	RegisterSecret(Proxy.HttpProxy, "test-http-proxy")

	viper.BindEnv(Proxy.HttpsProxy, "TEST_HTTPS_PROXY")
	RegisterSecret(Proxy.HttpsProxy, "test-https-proxy")

	viper.BindEnv(Proxy.UserCABundle, "USER_CA_BUNDLE")
	RegisterSecret(Proxy.UserCABundle, "user-ca-bundle")
}

func init() {
	InitViper()
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
	keyToSecretMapping = append(keyToSecretMapping, Secret{
		Key:          key,
		FileLocation: secretFileName,
	})
	keyToSecretMappingMutex.Unlock()
}

// GetAllSecrets will return Viper config keys and their corresponding secret filenames.
func GetAllSecrets() []Secret {
	return keyToSecretMapping
}

// LoadKubeconfig will, given a path to a kubeconfig, attempt to load it into the Viper config.
func LoadKubeconfig() error {
	kubeconfigPath := viper.GetString(Kubeconfig.Path)
	if kubeconfigPath != "" {
		kubeconfigBytes, err := ioutil.ReadFile(kubeconfigPath)
		if err != nil {
			return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", kubeconfigPath, err)
		}
		viper.Set(Kubeconfig.Contents, string(kubeconfigBytes))
	}
	return nil
}
