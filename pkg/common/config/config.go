// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	"time"
)

// Instance is the configuration used for end to end testing.
var Instance = new(Config)

// Config dictates the behavior of cluster tests.
type Config struct {
	Upgrade UpgradeConfig `yaml:"upgrade"`

	Kubeconfig KubeConfig `yaml:"kubeconfig"`

	Tests TestConfig `yaml:"tests"`

	Cluster ClusterConfig `yaml:"cluster"`

	OCM OCMConfig `yaml:"ocm"`

	Addons AddonConfig `yaml:"addons"`

	Scale ScaleConfig `yaml:"scale"`

	Prometheus PrometheusConfig `yaml:"prometheus"`

	Weather WeatherConfig `yaml:"weather"`

	// Provider is what provider to use to create/delete clusters.
	Provider string `json:"provider" env:"PROVIDER" sect:"tests" default:"ocm" yaml:"provider"`

	// JobName lets you name the current e2e job run
	JobName string `json:"job_name" env:"JOB_NAME" sect:"tests" yaml:"jobName"`

	// JobID is the ID designated by prow for this run
	JobID int `json:"job_id" env:"BUILD_NUMBER" sect:"tests" yaml:"jobID"`

	// BaseJobURL is the root location for all job artifacts
	// For example, https://storage.googleapis.com/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/61/build-log.txt would be
	// https://storage.googleapis.com/origin-ci-test/logs -- This is also our default
	BaseJobURL string `jon:"baseJobURL" env:"BASE_JOB_URL" sect:"test" yaml:"baseJobURL" default:"https://storage.googleapis.com/origin-ci-test/logs"`

	// ReportDir is the location JUnit XML results are written.
	ReportDir string `json:"report_dir,omitempty" env:"REPORT_DIR" sect:"tests" default:"__TMP_DIR__" yaml:"reportDir"`

	// Suffix is used at the end of test names to identify them.
	Suffix string `json:"suffix,omitempty" env:"SUFFIX" sect:"tests" default:"__RND_3__" yaml:"suffix"`

	// DryRun lets you run osde2e all the way up to the e2e tests then skips them.
	DryRun bool `json:"dry_run,omitempty" env:"DRY_RUN" sect:"tests"  yaml:"dryRun"`

	// LogMetrics is a collection of LogMetric structs used to crudely analyze test logs
	LogMetrics LogMetrics `json:"log-metrics" yaml:"logMetrics"`

	// MustGather will run a Must-Gather process upon completion of the tests.
	MustGather bool `json:"must_gather,omitempty" env:"MUST_GATHER" sect:"tests" default:"true" yaml:"mustGather"`
}

// KubeConfig stores information required to talk to the Kube API
type KubeConfig struct {
	// Path is the filepath of an existing Kubeconfig
	Path string `env:"TEST_KUBECONFIG" sect:"cluster" yaml:"path"`
}

// OCMConfig contains connect info for the OCM API
type OCMConfig struct {
	// Token is used to authenticate with OCM.
	Token string `json:"ocm_token" env:"OCM_TOKEN" sect:"required" yaml:"token"`

	// Env is the OpenShift Dedicated environment used to provision clusters.
	Env string `env:"OSD_ENV" sect:"environment" default:"prod" yaml:"env"`

	// Debug shows debug level messages when enabled.
	Debug bool `env:"DEBUG_OSD" sect:"environment" default:"false" yaml:"debug"`

	// NumRetries is the number of times to retry each OCM call.
	NumRetries int `env:"NUM_RETRIES" sect:"ocm" default:"3" yaml:"numRetries"`
}

// UpgradeConfig stores information required to perform OSDe2e upgrade testing
type UpgradeConfig struct {
	// UpgradeToCISIfPossible will upgrade to the most recent cluster image set if it's newer than the install version
	UpgradeToCISIfPossible bool `env:"UPGRADE_TO_CIS_IF_POSSIBLE" sect:"upgrade" default:"false" yaml:"upgradeToCISIfPossible"`

	// OnlyUpgradeToZReleases will restrict upgrades to selecting Z releases on stage/prod.
	OnlyUpgradeToZReleases bool `env:"ONLY_UPGRADE_TO_Z_RELEASES" sect:"upgrade" default:"false" yaml:"onlyUpgradeToZReleases"`

	// NextReleaseAfterProdDefault will select the cluster image set that the given number of releases away from the the production default.
	NextReleaseAfterProdDefaultForUpgrade int `env:"NEXT_RELEASE_AFTER_PROD_DEFAULT_FOR_UPGRADE" sect:"upgrade" default:"-1" yaml:"nextReleaseAfterProdDefaultForUpgrade"`

	// ReleaseStream used to retrieve latest release images. If set, it will be used to perform an upgrade.
	ReleaseStream string `env:"UPGRADE_RELEASE_STREAM" sect:"upgrade" yaml:"releaseStream"`
}

// ClusterConfig contains config information pertaining to an OSD cluster
type ClusterConfig struct {
	// MultiAZ deploys a cluster across multiple availability zones.
	MultiAZ bool `env:"MULTI_AZ" sect:"cluster" default:"false" yaml:"multiAZ"`

	// DestroyClusterAfterTest set to true if you want to the cluster to be explicitly deleted after the test.
	DestroyAfterTest bool `env:"DESTROY_CLUSTER" sect:"cluster" default:"false" yaml:"destroyAfterTest"`

	// ExpiryInMinutes is how long before a cluster expires and is deleted by OSD.
	ExpiryInMinutes int64 `env:"CLUSTER_EXPIRY_IN_MINUTES" sect:"cluster" default:"210" yaml:"expiryInMinutes"`

	// AfterTestWait is how long to keep a cluster around after tests have run.
	AfterTestWait int64 `env:"AFTER_TEST_CLUSTER_WAIT" sect:"environment" default:"60" yaml:"afterTestWait"`

	// InstallTimeout is how long to wait before failing a cluster launch.
	InstallTimeout int64 `env:"CLUSTER_UP_TIMEOUT" sect:"environment" default:"135" yaml:"installTimeout"`

	// UseLatestVersionForInstall will select the latest cluster image set available for a fresh install.
	UseLatestVersionForInstall bool `env:"USE_LATEST_VERSION_FOR_INSTALL" sect:"version" default:"false" yaml:"useLatestVersionForInstall"`

	// UseMiddleClusterImageSetForInstall will select the cluster image set that is in the middle of the list of ordered cluster versions known to OCM.
	UseMiddleClusterImageSetForInstall bool `env:"USE_MIDDLE_CLUSTER_IMAGE_SET_FOR_INSTALL" sect:"version" default:"false" yaml:"useMiddleClusterVersionForInstall"`

	// UseOldestClusterImageSetForInstall will select the cluster image set that is in the end of the list of ordered cluster versions known to OCM.
	UseOldestClusterImageSetForInstall bool `env:"USE_OLDEST_CLUSTER_IMAGE_SET_FOR_INSTALL" sect:"version" default:"false" yaml:"useOldestClusterVersionForInstall"`

	// PreviousReleaseFromDefault will select the clsuter image set that is the given number of releases before the current default.
	PreviousReleaseFromDefault int `env:"PREVIOUS_RELEASE_FROM_DEFAULT" sect:"version" default:"0" yaml:"previousReleaseFromDefault"`

	// NextReleaseAfterProdDefault will select the cluster image set that the given number of releases away from the the production default.
	NextReleaseAfterProdDefault int `env:"NEXT_RELEASE_AFTER_PROD_DEFAULT" sect:"version" default:"-1" yaml:"nextReleaseAfterProdDefault"`

	// MajorTarget is the major version to target. If specified, it is used in version selection.
	MajorTarget int64 `env:"MAJOR_TARGET" sect:"version" yaml:"majorTarget"`

	// MinorTarget is the minor version to target. If specified, it is used in version selection.
	MinorTarget int64 `env:"MINOR_TARGET" sect:"version" yaml:"minorTarget"`

	// CleanCheckRuns lets us set the number of osd-verify checks we want to run before deeming a cluster "healthy"
	CleanCheckRuns int `env:"CLEAN_CHECK_RUNS" sect:"environment" default:"20" yaml:"cleanCheckRuns"`
}

// AddonConfig options for addon testing
type AddonConfig struct {
	// IDs is an array of Addon IDs to install
	IDs []string `env:"ADDON_IDS" sect:"addons" yaml:"ids"`
	// TestHarnesses is an array of container images that will test the addon
	TestHarnesses []string `env:"ADDON_TEST_HARNESSES" sect:"addons" yaml:"testHarnesses"`
}

// ScaleConfig options for scale testing
type ScaleConfig struct {
	WorkloadsRepository string `env:"WORKLOADS_REPO" sect:"scale" default:"https://github.com/openshift-scale/workloads" yaml:"workloadsRepository"`

	WorkloadsRepositoryBranch string `env:"WORKLOADS_REPO_BRANCH" sect:"scale" default:"master" yaml:"workloadsRepositoryBranch"`
}

// TestConfig changes the behavior of how and what tests are run.
type TestConfig struct {
	// PollingTimeout is how long (in mimutes) to wait for an object to be created
	// before failing the test.
	PollingTimeout int64 `env:"POLLING_TIMEOUT" sect:"tests" default:"30" yaml:"pollingTimeout"`

	// GinkgoSkip is a regex passed to Ginkgo that skips any test suites matching the regex. ex. "Operator"
	GinkgoSkip string `env:"GINKGO_SKIP" sect:"tests" yaml:"ginkgoSkip"`

	// GinkgoFocus is a regex passed to Ginkgo that focus on any test suites matching the regex. ex. "Operator"
	GinkgoFocus string `env:"GINKGO_FOCUS" sect:"tests" yaml:"focus"`

	// TestsToRun is a list of files which should be executed as part of a test suite
	TestsToRun []string `env:"TESTS_TO_RUN" sect:"tests" yaml:"testsToRun"`

	// SuppressSkipNotifications suppresses the notifications of skipped tests
	SuppressSkipNotifications bool `env:"SUPPRESS_SKIP_NOTIFICATIONS" sect:"tests" default:"true" yaml:"suppressSkipNotifications"`

	// CleanRuns is the number of times the test-version is run before skipping.
	CleanRuns int `env:"CLEAN_RUNS" sect:"tests" yaml:"cleanRuns"`

	// OperatorSkip is a comma-delimited list of operator names to ignore health checks from. ex. "insights,telemetry"
	OperatorSkip string `env:"OPERATOR_SKIP" sect:"tests" default:"insights" yaml:"ginkgoFocus"`

	// SkipClusterHealthChecks skips the cluster health checks. Useful when developing against a running cluster.
	SkipClusterHealthChecks bool `env:"SKIP_CLUSTER_HEALTH_CHECKS" sect:"tests" default:"false" yaml:"skipClusterHealthChecks"`

	// UploadMetrics tells osde2e whether to try to upload to the S3 metrics bucket.
	UploadMetrics bool `env:"UPLOAD_METRICS" sect:"metrics" default:"false" yaml:"uploadMetrics"`

	// MetricsBucket is the bucket that metrics data will be uploaded to.
	MetricsBucket string `env:"METRICS_BUCKET" sect:"metrics" default:"osde2e-metrics" yaml:"metricsBucket"`

	// ServiceAccount defines what user the tests should run as. By default, osde2e uses system:admin
	ServiceAccount string `env:"SERVICE_ACCOUNT" sect:"tests" yaml:"serviceAccount"`
}

// PrometheusConfig contains configs for connecting to a Prometheus instance for querying.
type PrometheusConfig struct {
	// Address is the address of the Prometheus instance to connect to.
	Address string `env:"PROMETHEUS_ADDRESS" sect:"weather" yaml:"address"`

	// BearerToken is the token needed for communicating with Prometheus.
	BearerToken string `env:"PROMETHEUS_BEARER_TOKEN" sect:"weather" yaml:"bearerToken"`
}

// WeatherConfig describes various config options for weather reports.
type WeatherConfig struct {
	// StartOfTimeWindowInHours is how many hours to look back through results.
	StartOfTimeWindowInHours time.Duration `env:"START_OF_TIME_WINDOW_IN_HOURS" sect:"weather" default:"24" yaml:"startOfTimeWindowInHours"`

	// NumberOfSamplesNecessary is how many samples are necessary for generating a report.
	NumberOfSamplesNecessary int `env:"NUMBER_OF_SAMPLES_NECESSARY" sect:"weather" default:"3" yaml:"numberOfSamplesNecessary"`

	// SlackWebhook is the webhook to use to post the weather report to slack.
	SlackWebhook string `env:"SLACK_WEBHOOK" sect:"weather" yaml:"slackWebhook"`

	// JobWhitelist is a list of job regexes to consider in the weather report.
	JobWhitelist []string `env:"JOB_WHITELIST" sect:"weather" default:"osde2e-.*-aws-e2e-.*" yaml:"jobWhitelist"`
}
