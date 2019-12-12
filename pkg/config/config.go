// Package config provides the configuration for tests run as part of the osde2e suite.
package config

const (
	// EnvVarTag is the Go struct tag containing the environment variable that sets the option.
	EnvVarTag = "env"

	// SectionTag is the Go struct tag containing the documentation section of the option.
	SectionTag = "sect"

	// DefaultTag is the Go struct tag containing the default value of the option.
	DefaultTag = "default"
)

// Cfg is the configuration used for end to end testing.
var Cfg = new(Config)

// Config dictates the behavior of cluster tests.
type Config struct {
	Upgrade UpgradeConfig `yaml:"upgrade"`

	Kubeconfig KubeConfig `yaml:"kubeconfig"`

	Tests TestConfig `yaml:"tests"`

	Cluster ClusterConfig `yaml:"cluster"`

	OCM OCMConfig `yaml:"ocm"`

	// Name lets you name the current e2e job run
	JobName string `json:"job_name" env:"JOB_NAME" sect:"tests" yaml:"jobName"`

	// ReportDir is the location JUnit XML results are written.
	ReportDir string `json:"report_dir,omitempty" env:"REPORT_DIR" sect:"tests" yaml:"reportDir"`

	// Suffix is used at the end of test names to identify them.
	Suffix string `json:"suffix,omitempty" env:"SUFFIX" sect:"tests" yaml:"suffix"`

	// DryRun lets you run osde2e all the way up to the e2e tests then skips them.
	DryRun bool `json:"dry_run,omitempty" env:"DRY_RUN" sect:"tests"  yaml:"dryRun"`

	// InstalledWorkloads is an internal variable used to track currently installed workloads in this test run.
	InstalledWorkloads map[string]string

	// Phase is an internal variable used to track the current set of tests being run (install, upgrade).
	Phase string
}

// KubeConfig stores information required to talk to the Kube API
type KubeConfig struct {
	// Contents is the actual contents of a valid Kubeconfig
	Contents []byte `yaml:"contents"`

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
}

// UpgradeConfig stores information required to perform OSDe2e upgrade testing
type UpgradeConfig struct {
	// MajorTarget is the major version to target. If specified, it is used in version selection.
	MajorTarget int64 `env:"MAJOR_TARGET" sect:"version" yaml:"majorTarget"`

	// MinorTarget is the minor version to target. If specified, it is used in version selection.
	MinorTarget int64 `env:"MINOR_TARGET" sect:"version" yaml:"minorTarget"`

	// ReleaseStream used to retrieve latest release images. If set, it will be used to perform an upgrade.
	ReleaseStream string `env:"UPGRADE_RELEASE_STREAM" sect:"upgrade" yaml:"releaseStream"`

	// ReleaseName is the name of the release in a release stream. UpgradeReleaseStream must be set.
	ReleaseName string `env:"UPGRADE_RELEASE_NAME" sect:"upgrade" yaml:"releaseName"`

	// Image is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.
	Image string `env:"UPGRADE_IMAGE" sect:"upgrade" yaml:"image"`
}

// ClusterConfig contains config information pertaining to an OSD cluster
type ClusterConfig struct {
	// ID identifies the cluster. If set at start, an existing cluster is tested.
	ID string `json:"cluster_id,omitempty" env:"CLUSTER_ID" sect:"cluster" yaml:"id"`

	// Name is the name of the cluster being created.
	Name string `json:"cluster_name,omitempty" env:"CLUSTER_NAME" sect:"cluster" yaml:"name"`

	// Version is the version of the cluster being deployed.
	Version string `json:"cluster_version,omitempty" env:"CLUSTER_VERSION" sect:"version"  yaml:"version"`

	// AddOns
	AddOns []string `json:"addons,omitempty"  yaml:"addons"`

	// MultiAZ deploys a cluster across multiple availability zones.
	MultiAZ bool `env:"MULTI_AZ" sect:"cluster" default:"false" yaml:"multiAZ"`

	// DestroyAfterTest set to false if you want OCM to clean up the cluster itself after the test completes.
	DestroyAfterTest bool `env:"DESTROY_CLUSTER" sect:"cluster" default:"true" yaml:"destroyAfterTest"`

	// ExpiryInMinutes is how long before a cluster expires and is deleted by OSD.
	ExpiryInMinutes int64 `env:"CLUSTER_EXPIRY_IN_MINUTES" sect:"cluster" default:"210" yaml:"expiryInMinutes"`

	// AfterTestWait is how long to keep a cluster around after tests have run.
	AfterTestWait int64 `env:"AFTER_TEST_CLUSTER_WAIT" sect:"environment" default:"60" yaml:"afterTestWait"`

	// InstallTimeout is how long to wait before failing a cluster launch.
	InstallTimeout int64 `env:"CLUSTER_UP_TIMEOUT" sect:"environment" default:"135" yaml:"installTimeout"`
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

	// CleanRuns is the number of times the test-version is run before skipping.
	CleanRuns int `env:"CLEAN_RUNS" sect:"tests" yaml:"cleanRuns"`

	// OperatorSkip is a comma-delimited list of operator names to ignore health checks from. ex. "insights,telemetry"
	OperatorSkip string `env:"OPERATOR_SKIP" sect:"tests" default:"insights" yaml:"ginkgoFocus"`

	// UploadMetrics tells osde2e whether to try to upload to the S3 metrics bucket.
	UploadMetrics bool `env:"UPLOAD_METRICS" sect:"metrics" default:"false" yaml:"uploadMetrics"`

	// MetricsBucket is the bucket that metrics data will be uploaded to.
	MetricsBucket string `env:"METRICS_BUCKET" sect:"metrics" default:"osde2e-metrics" yaml:"metricsBucket"`
}
