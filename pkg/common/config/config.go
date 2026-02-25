// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"gopkg.in/yaml.v3"
)

// channel id for #hcm-cicd-notifications
const defaultNotificationsChannel = "C06HQR8HN0L"

type Secret struct {
	FileLocation string
	Key          string
}

// TestSuite represents a test image with optional slack channel
type TestSuite struct {
	Image        string `yaml:"image" json:"image" mapstructure:"image"`
	SlackChannel string `yaml:"slackChannel,omitempty" json:"slackChannel,omitempty" mapstructure:"slackChannel,omitempty"`
}

const (
	Success = 0
	Failure = 1
	Aborted = 130

	// KrknAIModeDiscover is the mode for discover mode
	KrknAIModeDiscover = "discover"

	// KrknAIModeRun is the mode for run mode
	KrknAIModeRun = "run"

	// KrknAIVerboseLevel is the verbosity level for krkn-ai output
	KrknAIVerboseLevel = "2"

	// Provider is what provider to use to create/delete clusters.
	// Env: PROVIDER
	Provider = "provider"

	// OcmConfig is the path for the ocm.json file.
	// Env: OCM_CONFIG
	OcmConfig = "ocmConfig"

	// JobName lets you name the current e2e job run
	// Env: JOB_NAME
	JobName = "jobName"

	// JobID is the ID designated by prow for this run
	// Env: BUILD_ID
	JobID = "jobID"

	// ProwJobId is the ID designated by prow for this run
	// Env: PROW_JOB_ID
	ProwJobId = "prowJobId"

	// JobType is the type of job according to prow for this run
	// Env: JOB_TYPE
	JobType = "jobType"

	// BaseJobURL is the root location for all job artifacts
	// For example, https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/61/build-log.txt would be
	// https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs -- This is also our default
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

	// SkipMustGather will skip the Must-Gather process upon completion of the tests.
	// Env: SKIP_MUST_GATHER
	SkipMustGather = "skipMustGather"

	// InstalledWorkloads is an internal variable used to track currently installed workloads in this test run.
	InstalledWorkloads = "installedWorkloads"

	// Phase is an internal variable used to track the current set of tests being run (install, upgrade).
	Phase = "phase"

	// Project is both the project and SA automatically created to house all objects created during an osde2e-run
	Project = "project"

	CanaryChance = "canaryChance"

	// DefaultNetworkProvider Default network provider for OSD
	// env: CLUSTER_NETWORK_PROVIDER
	DefaultNetworkProvider = "OVNKubernetes"

	// NonOSDe2eSecrets is an internal-only Viper Key.
	// End users should not be using this key, there may be unforeseen consequences.
	NonOSDe2eSecrets = "nonOSDe2eSecrets"

	// JobStartedAt tracks when the job began running.
	JobStartedAt = "JobStartedAt"

	// Hypershift enables the use of hypershift for cluster creation.
	Hypershift = "Hypershift"

	// SharedDir is the location where files to be used by other processes/programs are stored.
	// This is primarily used when running within Prow and using additional steps after osde2e finishes.
	SharedDir = "sharedDir"

	KonfluxTestOutputFile = "konfluxResultsPath"

	// SlackMessageLength TotalSlackMessageLength is about 10000 characters
	// Summary: 1500 Characters
	// Build file comment: 500 Characters
	// Other comments(s3, ec2, elasticIP, iam): 2000 * 4 = 8000
	SlackMessageLength int = 2000
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

	// Create disruptive Pod Disruption Budget workloads to test the Managed Upgrade Operator's ability to handle them.
	ManagedUpgradeTestPodDisruptionBudgets string

	// Create disruptive Node Drain workload to test the Managed Upgrade Operator's ability to handle them.
	ManagedUpgradeTestNodeDrain string

	// Reschedule the upgrade via provider before commence
	ManagedUpgradeRescheduled string

	// Toggle on/off running pre upgrade tests
	RunPreUpgradeTests string

	// Toggle on/off running post upgrade tests
	RunPostUpgradeTests string
}{
	UpgradeToLatest:                        "upgrade.toLatest",
	UpgradeToLatestZ:                       "upgrade.ToLatestZ",
	UpgradeToLatestY:                       "upgrade.ToLatestY",
	ReleaseName:                            "upgrade.releaseName",
	Image:                                  "upgrade.image",
	Type:                                   "upgrade.type",
	UpgradeVersionEqualToInstallVersion:    "upgrade.upgradeVersionEqualToInstallVersion",
	ManagedUpgradeTestPodDisruptionBudgets: "upgrade.managedUpgradeTestPodDisruptionBudgets",
	ManagedUpgradeTestNodeDrain:            "upgrade.managedUpgradeTestNodeDrain",
	ManagedUpgradeRescheduled:              "upgrade.managedUpgradeRescheduled",
	RunPreUpgradeTests:                     "upgrade.runPreUpgradeTests",
	RunPostUpgradeTests:                    "upgrade.runPostUpgradeTests",
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
var Tests = struct {
	// SuiteTimeout is how long (in hours) to wait for the entire suite to finish before timing out
	// Env: SUITE_TIMEOUT
	SuiteTimeout string

	// AdHocTestContainerTimeout is how long (in seconds) to wait for the individual adHocTestImage to finish before timing out. If unspecified, POLLING_TIMEOUT is used.
	// Env: AD_HOC_TEST_CONTAINER_TIMEOUT
	AdHocTestContainerTimeout string

	// AdHocTestImages is a list of test adHocTestImages to run (DEPRECATED - use TestSuites).
	// Env: AD_HOC_TEST_IMAGES
	AdHocTestImages string

	// TestSuites is a list of test suites to run with optional slack channels.
	// Env: TEST_SUITES_YAML
	TestSuites string

	// PollingTimeout is how long (in seconds) to wait for an object to be created before failing the test.
	// Env: POLLING_TIMEOUT
	PollingTimeout string

	// TestUser is the OpenShift user that the tests will run as
	// If "%s" is detected in the TestUser string, it will evaluate that as the project namespace
	// Example: "system:serviceaccount:%s:dedicated-admin"
	// Evaluated: "system:serviceaccount:osde2e-abc123:dedicated-admin"
	// Env: TEST_USER
	TestUser string

	// SlackChannel is the name of a slack channel in the Internal Red hat slack workspace that will
	// receive an alert if the tests fail.
	// Env: SLACK_CHANNEL
	SlackChannel string

	// Slack Webhook is the URL to osde2e owner channel for Cloud Account Cleanup Report workflow to send notifications.
	// Env: SLACK_WEBHOOK
	SlackWebhook string

	// SlackNotify is a boolean that determines if Slack notifications should be sent.
	// Env: SLACK_NOTIFY
	EnableSlackNotify string

	// GinkgoSkip is a regex passed to Ginkgo that skips any test suites matching the regex. ex. "Operator"
	// Env: GINKGO_SKIP
	GinkgoSkip string

	// GinkgoFocus is a regex passed to Ginkgo that focus on any test suites matching the regex. ex. "Operator"
	// Env: GINKGO_FOCUS
	GinkgoFocus string

	// GinkgoLogLevel controls the logging level used by ginkgo when providing test output
	// Env: GINKGO_LOG_LEVEL
	GinkgoLogLevel string

	// GinkgoLabelFilter controls which test suites or tests to run
	// Env: GINKGO_LABEL_FILTER
	GinkgoLabelFilter string

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

	// ClusterHealthChecksTimeout defines the duration for which the test will
	// wait for the cluster to indicate it is healthy before cancelling the test
	// run. This value should be formatted for use with time.ParseDuration.
	// Env: CLUSTER_HEALTH_CHECKS_TIMEOUT
	ClusterHealthChecksTimeout string

	// LogBucket is the s3 bucket that log file/s will be uploaded to.
	// Env: LOG_BUCKET
	LogBucket string

	// ServiceAccount defines what user the tests should run as. By default, osde2e uses system:admin
	// Env: SERVICE_ACCOUNT
	ServiceAccount string

	// OnlyHealthcheckNodes focuses pre-install validation only on the nodes
	// Env: ONLY_HEALTH_CHECK_NODES
	OnlyHealthCheckNodes string
}{
	AdHocTestImages:            "tests.adHocTestImages",
	TestSuites:                 "tests.testSuites",
	SuiteTimeout:               "tests.suiteTimeout",
	AdHocTestContainerTimeout:  "tests.adHocTestContainerTimeout",
	PollingTimeout:             "tests.pollingTimeout",
	ServiceAccount:             "tests.serviceAccount",
	SlackChannel:               "tests.slackChannel",
	SlackWebhook:               "tests.slackWebhook",
	EnableSlackNotify:          "tests.enableSlackNotify",
	GinkgoSkip:                 "tests.ginkgoSkip",
	GinkgoFocus:                "tests.focus",
	GinkgoLogLevel:             "tests.ginkgoLogLevel",
	GinkgoLabelFilter:          "tests.ginkgoLabelFilter",
	TestsToRun:                 "tests.testsToRun",
	SuppressSkipNotifications:  "tests.suppressSkipNotifications",
	CleanRuns:                  "tests.cleanRuns",
	OperatorSkip:               "tests.operatorSkip",
	SkipClusterHealthChecks:    "tests.skipClusterHealthChecks",
	LogBucket:                  "tests.logBucket",
	ClusterHealthChecksTimeout: "tests.clusterHealthChecksTimeout",
	OnlyHealthCheckNodes:       "tests.onlyHealthCheckNodes",
}

// Cluster config keys.
var Cluster = struct {
	// Reserve  creates a reserve of testing-ready cluster and skips all tests.
	// Env: RESERVE
	// Arg --reserve
	Reserve string

	// MultiAZ deploys a cluster across multiple availability zones.
	// Env: MULTI_AZ
	MultiAZ string

	// Channel dictates which install/upgrade edges will be available to the cluster
	// Env: CHANNEL
	Channel string

	// SkipDestroyCluster indicates whether cluster should be destroyed after test completion.
	// Env: SKIP_DESTROY_CLUSTER
	SkipDestroyCluster string

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

	// InstallLatestXY will select the latest version available given an "X.Y" formatted string
	InstallLatestXY string

	// InstallLatestYFromDelta will select the latest Y from the delta (+/-) given
	InstallLatestYFromDelta string

	// InstallLatestZFromDelta will select the latest Z from the delta (+/-) given
	InstallLatestZFromDelta string

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

	// UseClusterReserve will allow the test run to use an existing cluster if available
	// ENV: USE_CLUSTER_RESERVE, will also accept obsoleted var name: USE_EXISTING_CLUSTER
	// Default: True
	UseClusterReserve string

	// Passing tracks the internal status of the tests: Pass or Fail
	Passing string

	// ClaimedFromReserve tracks whether this cluster's test run used a new or recycled cluster
	ClaimedFromReserve string

	// InspectNamespaces is a comma-delimited list of namespaces to perform an inspect on during test cleanup
	InspectNamespaces string

	// UseProxyForInstall will attempt to use a cluster-wide proxy for cluster installation, provided that a cluster-wide proxy config is supplied
	UseProxyForInstall string

	// EnableFips enables the FIPS test suite
	// Env: ENABLE_FIPS
	EnableFips string

	// FedRamp will enable OSDe2e to run in a FedRamp environment
	// Env: FEDRAMP
	FedRamp string
}{
	MultiAZ:                             "cluster.multiAZ",
	Channel:                             "cluster.channel",
	SkipDestroyCluster:                  "cluster.skipDestroyCluster",
	Reserve:                             "cluster.reserve",
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
	InstallLatestXY:                     "cluster.installLatestXY",
	InstallLatestYFromDelta:             "cluster.installLatestYFromDelta",
	InstallLatestZFromDelta:             "cluster.installLatestZFromDelta",
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
	UseClusterReserve:                   "cluster.useClusterReserve",
	Passing:                             "cluster.passing",
	ClaimedFromReserve:                  "cluster.claimedFromReserve",
	InspectNamespaces:                   "cluster.inspectNamespaces",
	EnableFips:                          "cluster.enableFips",
	FedRamp:                             "cluster.fedRamp",
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
}{
	IDsAtCreation: "addons.idsAtCreation",
	IDs:           "addons.ids",
	SkipAddonList: "addons.skipAddonlist",
	Parameters:    "addons.parameters",
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

// Cad Configuration Anomaly Detection config
var Cad = struct {
	// Env: CAD_PAGERDUTY_ROUTING_KEY
	CADPagerDutyRoutingKey string
}{
	CADPagerDutyRoutingKey: "cad.pagerDutyRoutingKey",
}

var LogAnalysis = struct {
	// EnableAnalysis enables log analysis powered failure analysis
	// Env: LOG_ANALYSIS_ENABLE
	EnableAnalysis string

	// APIKey is the API key for the LLM service (e.g., Gemini)
	// Env: GEMINI_API_KEY
	APIKey string

	// Model specifies which LLM model to use
	// Env: LLM_MODEL
	Model string

	// SlackWebhook is the Slack webhook URL for log analysis notifications
	// Env: LOG_ANALYSIS_SLACK_WEBHOOK
	SlackWebhook string

	// SlackChannel is the default Slack channel for OSDE2E notifications
	// Env: LOG_ANALYSIS_SLACK_CHANNEL
	SlackChannel string
}{
	EnableAnalysis: "logAnalysis.enableAnalysis",
	APIKey:         "logAnalysis.apiKey",
	Model:          "logAnalysis.model",
	SlackWebhook:   "logAnalysis.slackWebhook",
	SlackChannel:   "logAnalysis.slackChannel",
}

// KrknAI config keys for Kraken AI chaos testing.
var KrknAI = struct {
	// Namespace is the target namespace for chaos testing
	// Env: KRKN_NAMESPACE
	Namespace string

	// PodLabel is the label selector for targeting pods
	// Env: KRKN_POD_LABEL
	PodLabel string

	// NodeLabel is the label selector for targeting nodes
	// Env: KRKN_NODE_LABEL
	NodeLabel string

	// SkipPodName is a pattern to skip specific pods
	// Env: KRKN_SKIP_POD_NAME
	SkipPodName string

	// FitnessQuery is the Prometheus query for the fitness function
	// Env: KRKN_FITNESS_QUERY
	FitnessQuery string

	// Scenarios is a comma-separated list of scenarios to enable
	// Env: KRKN_SCENARIOS
	Scenarios string
}{
	Namespace:    "krknAI.namespace",
	PodLabel:     "krknAI.podLabel",
	NodeLabel:    "krknAI.nodeLabel",
	SkipPodName:  "krknAI.skipPodName",
	FitnessQuery: "krknAI.fitnessQuery",
	Scenarios:    "krknAI.scenarios",
}

func InitOSDe2eViper() {
	// Here's where we bind environment variables to config options and set defaults

	viper.SetConfigType("yaml") // Our configs are all in yaml.

	// capture job startup time
	viper.SetDefault(JobStartedAt, time.Now().UTC().Format(time.RFC3339))

	// ----- Top Level Configs -----
	viper.SetDefault(Provider, "ocm")
	_ = viper.BindEnv(Provider, "PROVIDER")

	viper.SetDefault(OcmConfig, fmt.Sprintf("%s/ocm.json", os.TempDir()))
	_ = viper.BindEnv(OcmConfig, "OCM_CONFIG")
	os.Setenv("OCM_CONFIG", viper.GetString(OcmConfig))

	_ = viper.BindEnv(JobName, "JOB_NAME")
	_ = viper.BindEnv(JobType, "JOB_TYPE")

	viper.SetDefault(JobID, -1)
	_ = viper.BindEnv(JobID, "BUILD_ID")

	viper.SetDefault(BaseJobURL, "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/logs")
	_ = viper.BindEnv(BaseJobURL, "BASE_JOB_URL")

	viper.SetDefault(BaseProwURL, "https://deck-ci.apps.ci.l2s4.p1.openshiftapps.com")
	_ = viper.BindEnv(BaseProwURL, "BASE_PROW_URL")

	// ARTIFACTS and REPORT_DIR are basically the same, but ARTIFACTS is used on prow.
	_ = viper.BindEnv(Artifacts, "ARTIFACTS")

	_ = viper.BindEnv(ReportDir, "REPORT_DIR")

	_ = viper.BindEnv(SharedDir, "SHARED_DIR")

	_ = viper.BindEnv(KonfluxTestOutputFile, "KONFLUX_TEST_OUTPUT_FILE")

	_ = viper.BindEnv(Suffix, "SUFFIX")

	viper.SetDefault(DryRun, false)
	_ = viper.BindEnv(DryRun, "DRY_RUN")

	viper.SetDefault(SkipMustGather, false)
	_ = viper.BindEnv(SkipMustGather, "SKIP_MUST_GATHER")

	_ = viper.BindEnv(CanaryChance, "CANARY_CHANCE")

	// ----- Upgrade -----
	_ = viper.BindEnv(Upgrade.UpgradeToLatest, "UPGRADE_TO_LATEST")
	viper.SetDefault(Upgrade.UpgradeToLatest, false)

	_ = viper.BindEnv(Upgrade.UpgradeToLatestZ, "UPGRADE_TO_LATEST_Z")
	viper.SetDefault(Upgrade.UpgradeToLatestZ, false)

	_ = viper.BindEnv(Upgrade.UpgradeToLatestY, "UPGRADE_TO_LATEST_Y")
	viper.SetDefault(Upgrade.UpgradeToLatestY, false)

	_ = viper.BindEnv(Upgrade.ReleaseName, "UPGRADE_RELEASE_NAME")

	_ = viper.BindEnv(Upgrade.Image, "UPGRADE_IMAGE")

	viper.SetDefault(Upgrade.Type, "OSD")
	_ = viper.BindEnv(Upgrade.Type, "UPGRADE_TYPE")

	viper.SetDefault(Upgrade.UpgradeVersionEqualToInstallVersion, false)

	_ = viper.BindEnv(Upgrade.ManagedUpgradeTestPodDisruptionBudgets, "UPGRADE_MANAGED_TEST_PDBS")
	viper.SetDefault(Upgrade.ManagedUpgradeTestPodDisruptionBudgets, true)

	_ = viper.BindEnv(Upgrade.ManagedUpgradeTestNodeDrain, "UPGRADE_MANAGED_TEST_DRAIN")
	viper.SetDefault(Upgrade.ManagedUpgradeTestNodeDrain, true)

	_ = viper.BindEnv(Upgrade.ManagedUpgradeRescheduled, "UPGRADE_MANAGED_TEST_RESCHEDULE")
	viper.SetDefault(Upgrade.ManagedUpgradeRescheduled, false)

	_ = viper.BindEnv(Upgrade.RunPreUpgradeTests, "UPGRADE_RUN_PRE_TESTS")
	viper.SetDefault(Upgrade.RunPreUpgradeTests, false)

	_ = viper.BindEnv(Upgrade.RunPostUpgradeTests, "UPGRADE_RUN_POST_TESTS")
	viper.SetDefault(Upgrade.RunPostUpgradeTests, true)

	// ----- Kubeconfig -----
	_ = viper.BindEnv(Kubeconfig.Path, "TEST_KUBECONFIG")

	// ----- Tests -----
	_ = viper.BindEnv(Tests.AdHocTestImages, "AD_HOC_TEST_IMAGES")
	_ = viper.BindEnv(Tests.TestSuites, "TEST_SUITES_YAML")

	viper.SetDefault(Tests.SuiteTimeout, 6)
	_ = viper.BindEnv(Tests.SuiteTimeout, "SUITE_TIMEOUT")

	viper.SetDefault(Tests.AdHocTestContainerTimeout, "30m")
	_ = viper.BindEnv(Tests.AdHocTestContainerTimeout, "AD_HOC_TEST_CONTAINER_TIMEOUT")

	viper.SetDefault(Tests.PollingTimeout, 300)
	_ = viper.BindEnv(Tests.PollingTimeout, "POLLING_TIMEOUT")

	viper.SetDefault(Tests.TestUser, "system:serviceaccount:%s:cluster-admin")
	_ = viper.BindEnv(Tests.TestUser, "TEST_USER")

	_ = viper.BindEnv(Tests.GinkgoSkip, "GINKGO_SKIP")

	_ = viper.BindEnv(Tests.GinkgoFocus, "GINKGO_FOCUS")

	_ = viper.BindEnv(Tests.GinkgoLogLevel, "GINKGO_LOG_LEVEL")

	_ = viper.BindEnv(Tests.GinkgoLabelFilter, "GINKGO_LABEL_FILTER")

	_ = viper.BindEnv(Tests.TestsToRun, "TESTS_TO_RUN")

	viper.SetDefault(Tests.SuppressSkipNotifications, true)
	_ = viper.BindEnv(Tests.SuppressSkipNotifications, "SUPPRESS_SKIP_NOTIFICATIONS")

	_ = viper.BindEnv(Tests.CleanRuns, "CLEAN_RUNS")

	viper.SetDefault(Tests.OperatorSkip, "insights")
	_ = viper.BindEnv(Tests.OperatorSkip, "OPERATOR_SKIP")

	viper.SetDefault(Tests.SkipClusterHealthChecks, false)
	_ = viper.BindEnv(Tests.SkipClusterHealthChecks, "SKIP_CLUSTER_HEALTH_CHECKS")

	viper.SetDefault(Tests.ClusterHealthChecksTimeout, "2h")
	_ = viper.BindEnv(Tests.ClusterHealthChecksTimeout, "CLUSTER_HEALTH_CHECKS_TIMEOUT")

	_ = viper.BindEnv(Tests.LogBucket, "LOG_BUCKET")

	_ = viper.BindEnv(Tests.ServiceAccount, "SERVICE_ACCOUNT")

	_ = viper.BindEnv(Tests.OnlyHealthCheckNodes, "ONLY_HEALTH_CHECK_NODES")

	viper.SetDefault(Tests.SlackChannel, "hcm-cicd-alerts")
	_ = viper.BindEnv(Tests.SlackChannel, "SLACK_CHANNEL")

	_ = viper.BindEnv(Tests.SlackWebhook, "SLACK_WEBHOOK")
	RegisterSecret(Tests.SlackWebhook, "cleanup-job-notification-webhook")

	viper.SetDefault(Tests.EnableSlackNotify, false)
	_ = viper.BindEnv(Tests.EnableSlackNotify, "SLACK_NOTIFY")

	// ----- Cluster -----
	viper.SetDefault(Cluster.MultiAZ, false)
	_ = viper.BindEnv(Cluster.MultiAZ, "MULTI_AZ")

	viper.SetDefault(Cluster.Channel, "stable")
	_ = viper.BindEnv(Cluster.Channel, "CHANNEL")

	viper.SetDefault(Cluster.SkipDestroyCluster, false)
	_ = viper.BindEnv(Cluster.SkipDestroyCluster, "SKIP_DESTROY_CLUSTER")

	_ = viper.BindEnv(Cluster.Reserve, "RESERVE")

	viper.SetDefault(Cluster.ExpiryInMinutes, 360)
	_ = viper.BindEnv(Cluster.ExpiryInMinutes, "CLUSTER_EXPIRY_IN_MINUTES")

	viper.SetDefault(Cluster.AfterTestWait, 60)
	_ = viper.BindEnv(Cluster.AfterTestWait, "AFTER_TEST_CLUSTER_WAIT")

	viper.SetDefault(Cluster.InstallTimeout, 135)
	_ = viper.BindEnv(Cluster.InstallTimeout, "CLUSTER_UP_TIMEOUT")

	_ = viper.BindEnv(Cluster.ReleaseImageLatest, "RELEASE_IMAGE_LATEST")

	_ = viper.BindEnv(ProwJobId, "PROW_JOB_ID")

	viper.SetDefault(Cluster.UseProxyForInstall, false)
	_ = viper.BindEnv(Cluster.UseProxyForInstall, "USE_PROXY_FOR_INSTALL")

	viper.SetDefault(Hypershift, false)
	_ = viper.BindEnv(Hypershift, "HYPERSHIFT")

	viper.SetDefault(Cluster.UseLatestVersionForInstall, false)
	_ = viper.BindEnv(Cluster.UseLatestVersionForInstall, "USE_LATEST_VERSION_FOR_INSTALL")

	viper.SetDefault(Cluster.UseMiddleClusterImageSetForInstall, false)
	_ = viper.BindEnv(Cluster.UseMiddleClusterImageSetForInstall, "USE_MIDDLE_CLUSTER_IMAGE_SET_FOR_INSTALL")

	viper.SetDefault(Cluster.UseOldestClusterImageSetForInstall, false)
	_ = viper.BindEnv(Cluster.UseOldestClusterImageSetForInstall, "USE_OLDEST_CLUSTER_IMAGE_SET_FOR_INSTALL")

	viper.SetDefault(Cluster.LatestYReleaseAfterProdDefault, false)
	_ = viper.BindEnv(Cluster.LatestYReleaseAfterProdDefault, "LATEST_Y_RELEASE_AFTER_PROD_DEFAULT")

	viper.SetDefault(Cluster.LatestZReleaseAfterProdDefault, false)
	_ = viper.BindEnv(Cluster.LatestZReleaseAfterProdDefault, "LATEST_Z_RELEASE_AFTER_PROD_DEFAULT")

	_ = viper.BindEnv(Cluster.InstallSpecificNightly, "INSTALL_LATEST_NIGHTLY")

	_ = viper.BindEnv(Cluster.InstallLatestXY, "INSTALL_LATEST_XY")

	_ = viper.BindEnv(Cluster.InstallLatestYFromDelta, "INSTALL_LATEST_Y_FROM_DELTA")

	_ = viper.BindEnv(Cluster.InstallLatestZFromDelta, "INSTALL_LATEST_Z_FROM_DELTA")

	viper.SetDefault(Cluster.DeltaReleaseFromDefault, 0)
	_ = viper.BindEnv(Cluster.DeltaReleaseFromDefault, "DELTA_RELEASE_FROM_DEFAULT")

	viper.SetDefault(Cluster.NextReleaseAfterProdDefault, -1)
	_ = viper.BindEnv(Cluster.NextReleaseAfterProdDefault, "NEXT_RELEASE_AFTER_PROD_DEFAULT")

	viper.SetDefault(Cluster.CleanCheckRuns, 20)
	_ = viper.BindEnv(Cluster.CleanCheckRuns, "CLEAN_CHECK_RUNS")

	viper.SetDefault(Cluster.ID, "")
	_ = viper.BindEnv(Cluster.ID, "CLUSTER_ID")

	viper.SetDefault(Cluster.Name, "")
	_ = viper.BindEnv(Cluster.Name, "CLUSTER_NAME")

	viper.SetDefault(Cluster.Version, "")
	_ = viper.BindEnv(Cluster.Version, "CLUSTER_VERSION")

	viper.SetDefault(Cluster.EnoughVersionsForOldestOrMiddleTest, true)

	viper.SetDefault(Cluster.PreviousVersionFromDefaultFound, true)

	viper.SetDefault(Cluster.ProvisionShardID, "")
	_ = viper.BindEnv(Cluster.ProvisionShardID, "PROVISION_SHARD_ID")

	viper.SetDefault(Cluster.NumWorkerNodes, "")
	_ = viper.BindEnv(Cluster.NumWorkerNodes, "NUM_WORKER_NODES")

	_ = viper.BindEnv(Cluster.ImageContentSource, "CLUSTER_IMAGE_CONTENT_SOURCE")
	_ = viper.BindEnv(Cluster.InstallConfig, "CLUSTER_INSTALL_CONFIG")

	viper.SetDefault(Cluster.NetworkProvider, DefaultNetworkProvider)
	_ = viper.BindEnv(Cluster.NetworkProvider, "CLUSTER_NETWORK_PROVIDER")

	viper.SetDefault(Cluster.UseClusterReserve, false)
	_ = viper.BindEnv(Cluster.UseClusterReserve, "USE_EXISTING_CLUSTER", "USE_CLUSTER_RESERVE")

	viper.SetDefault(Cluster.ClaimedFromReserve, false)
	viper.SetDefault(Cluster.Passing, false)

	viper.SetDefault(Cluster.InspectNamespaces, strings.Join(defaultInspectNamespaces, ","))
	_ = viper.BindEnv(Cluster.InspectNamespaces, "INSPECT_NAMESPACES")

	viper.SetDefault(Cluster.EnableFips, false)
	_ = viper.BindEnv(Cluster.EnableFips, "ENABLE_FIPS")

	viper.SetDefault(Cluster.FedRamp, false)
	_ = viper.BindEnv(Cluster.FedRamp, "FEDRAMP")
	RegisterSecret(Cluster.FedRamp, "fedramp")

	// ----- Cloud Provider -----
	viper.SetDefault(CloudProvider.CloudProviderID, "aws")
	_ = viper.BindEnv(CloudProvider.CloudProviderID, "CLOUD_PROVIDER_ID")

	viper.SetDefault(CloudProvider.Region, "us-east-1")
	_ = viper.BindEnv(CloudProvider.Region, "CLOUD_PROVIDER_REGION")

	// ----- Addons -----
	_ = viper.BindEnv(Addons.IDsAtCreation, "ADDON_IDS_AT_CREATION")

	_ = viper.BindEnv(Addons.IDs, "ADDON_IDS")

	viper.SetDefault(Addons.Parameters, "{}")
	_ = viper.BindEnv(Addons.Parameters, "ADDON_PARAMETERS")
	RegisterSecret(Addons.Parameters, "addon-parameters")

	viper.SetDefault(Addons.SkipAddonList, false)
	_ = viper.BindEnv(Addons.SkipAddonList, "SKIP_ADDON_LIST")

	// ----- Proxy ------
	_ = viper.BindEnv(Proxy.HttpProxy, "TEST_HTTP_PROXY")
	RegisterSecret(Proxy.HttpProxy, "test-http-proxy")

	_ = viper.BindEnv(Proxy.HttpsProxy, "TEST_HTTPS_PROXY")
	RegisterSecret(Proxy.HttpsProxy, "test-https-proxy")

	_ = viper.BindEnv(Proxy.UserCABundle, "USER_CA_BUNDLE")
	RegisterSecret(Proxy.UserCABundle, "user-ca-bundle")

	// ------- Configuration Anomaly Detection ------
	viper.SetDefault(Cad.CADPagerDutyRoutingKey, "notprovided")
	_ = viper.BindEnv(Cad.CADPagerDutyRoutingKey, "CAD_PAGERDUTY_ROUTING_KEY")
	RegisterSecret(Cad.CADPagerDutyRoutingKey, "pagerduty-routing-key")

	// ----- LLM Configuration -----
	viper.SetDefault(LogAnalysis.EnableAnalysis, false)
	_ = viper.BindEnv(LogAnalysis.EnableAnalysis, "LOG_ANALYSIS_ENABLE")

	_ = viper.BindEnv(LogAnalysis.APIKey, "GEMINI_API_KEY")
	RegisterSecret(LogAnalysis.APIKey, "gemini-api-key")

	viper.SetDefault(LogAnalysis.Model, "gemini-2.5-pro")
	_ = viper.BindEnv(LogAnalysis.Model, "LLM_MODEL")

	viper.SetDefault(LogAnalysis.SlackWebhook, "")
	_ = viper.BindEnv(LogAnalysis.SlackWebhook, "LOG_ANALYSIS_SLACK_WEBHOOK")

	viper.SetDefault(LogAnalysis.SlackChannel, defaultNotificationsChannel)
	_ = viper.BindEnv(LogAnalysis.SlackChannel, "LOG_ANALYSIS_SLACK_CHANNEL")

	// ----- KrknAI Configuration -----
	viper.SetDefault(KrknAI.Namespace, "default")
	_ = viper.BindEnv(KrknAI.Namespace, "KRKN_NAMESPACE")

	viper.SetDefault(KrknAI.PodLabel, "")
	_ = viper.BindEnv(KrknAI.PodLabel, "KRKN_POD_LABEL")

	viper.SetDefault(KrknAI.NodeLabel, "kubernetes.io/hostname")
	_ = viper.BindEnv(KrknAI.NodeLabel, "KRKN_NODE_LABEL")

	viper.SetDefault(KrknAI.SkipPodName, "")
	_ = viper.BindEnv(KrknAI.SkipPodName, "KRKN_SKIP_POD_NAME")

	viper.SetDefault(KrknAI.FitnessQuery, "")
	_ = viper.BindEnv(KrknAI.FitnessQuery, "KRKN_FITNESS_QUERY")

	viper.SetDefault(KrknAI.Scenarios, "")
	_ = viper.BindEnv(KrknAI.Scenarios, "KRKN_SCENARIOS")
}

func init() {
	InitOSDe2eViper()
	if err := InitAWSViper(); err != nil {
		log.Fatalf("Could not init AWS config: %v", err)
	}
	InitGCPViper()
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

var loadOnce sync.Once

// LoadKubeconfig will, given a path to a kubeconfig, attempt to load it into the Viper config.
func LoadKubeconfig() error {
	var kubeconfigBytes []byte
	var err error
	loadOnce.Do(func() {
		kubeconfigPath := viper.GetString(Kubeconfig.Path)
		if kubeconfigPath != "" && viper.GetString(Kubeconfig.Contents) == "" {
			kubeconfigBytes, err = os.ReadFile(kubeconfigPath)
			if err != nil {
				err = fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", kubeconfigPath, err)
			}
			viper.Set(Kubeconfig.Contents, string(kubeconfigBytes))
		}
	})
	return err
}

// LoadClusterId  given a path to a shared directoru, if a cluster id is written to it, will attempt to load it into the Viper config.
// No error if shared cluster-id file doesn't exist.
func LoadClusterId() error {
	// get cluster id from shared_dir (used in prow multi-step jobs
	if viper.GetString(Cluster.ID) == "" && viper.GetString(SharedDir) != "" {
		sharedClusterIdPath := viper.GetString(SharedDir) + "/cluster-id"
		_, err := os.Stat(sharedClusterIdPath)
		if err == nil {
			clusteridbytes, err := os.ReadFile(sharedClusterIdPath)
			if err == nil {
				clusterID := string(clusteridbytes)
				fmt.Printf("cluster-id found in SHARED_DIR %s", clusterID)
				viper.Set(Cluster.ID, clusterID)
			} else {
				return fmt.Errorf("will not load shared cluster-id: %s", err.Error())
			}
		}
	}
	return nil
}

// GetTestSuites returns test suites, supporting both new TestSuites and legacy AdHocTestImages formats.
// Checks TestSuites first, then falls back to legacy AdHocTestImages string slice.
func GetTestSuites() ([]TestSuite, error) {
	// Priority 1: Try new TestSuites format
	if viper.IsSet(Tests.TestSuites) {
		// Try structured unmarshaling first (config files)
		var testSuites []TestSuite
		if err := viper.UnmarshalKey(Tests.TestSuites, &testSuites); err == nil {
			return testSuites, nil
		}

		// Try parsing as YAML string from environment variable
		strValue := viper.GetString(Tests.TestSuites)
		var suites []TestSuite
		if err := yaml.Unmarshal([]byte(strValue), &suites); err != nil {
			return nil, fmt.Errorf("failed to parse TEST_SUITES_YAML, format should be: '- image: ...\\n  slackChannel: ...'")
		}
		return suites, nil
	}

	// Priority 2: Try legacy AdHocTestImages as string slice (simple image names only)
	if viper.IsSet(Tests.AdHocTestImages) {
		legacyImages := viper.GetStringSlice(Tests.AdHocTestImages)
		var suites []TestSuite
		for _, img := range legacyImages {
			if img != "" {
				suites = append(suites, TestSuite{Image: img})
			}
		}
		return suites, nil
	}

	return []TestSuite{}, nil
}

// GetAdHocTestImagesAsString returns only the images from the test suites configuration as a comma-separated string
func GetAdHocTestImagesAsString() string {
	suites, err := GetTestSuites()
	if err != nil {
		return ""
	}

	var imageNames []string
	for _, suite := range suites {
		imageNames = append(imageNames, suite.Image)
	}

	return strings.Join(imageNames, ",")
}
