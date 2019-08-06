// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	"time"
)

const (
	// EnvVarTag is the Go struct tag containing the environment variable that sets the option.
	EnvVarTag = "env"

	// SectionTag is the Go struct tag containing the documentation section of the option.
	SectionTag = "sect"
)

// Cfg is the configuration used for end to end testing.
var Cfg = new(Config)

// Config dictates the behavior of cluster tests.
type Config struct {
	// ReportDir is the location JUnit XML results are written.
	ReportDir string `env:"REPORT_DIR"`

	// Suffix is used at the end of test names to identify them.
	Suffix string

	// UHCToken is used to authenticate with UHC.
	UHCToken string `env:"UHC_TOKEN"`

	// ClusterID identifies the cluster. If set at start, an existing cluster is tested.
	ClusterID string `env:"CLUSTER_ID"`

	// ClusterName is the name of the cluster being created.
	ClusterName string `env:"CLUSTER_NAME"`

	// ClusterVersion is the version of the cluster being deployed.
	ClusterVersion string `env:"CLUSTER_VERSION"`

	// MajorTarget is the major version to target. If specified, it is used in version selection.
	MajorTarget int64 `env:"MAJOR_TARGET"`

	// MinorTarget is the minor version to target. If specified, it is used in version selection.
	MinorTarget int64 `env:"MINOR_TARGET"`

	// ClusterUpTimeout is how long to wait before failing a cluster launch.
	ClusterUpTimeout time.Duration

	// TestGridBucket is the Google Cloud Storage bucket where results are reported for TestGrid.
	TestGridBucket string `env:"TESTGRID_BUCKET"`

	// TestGridPrefix is used to namespace reports.
	TestGridPrefix string `env:"TESTGRID_PREFIX"`

	// TestGridServiceAccount is a Base64 encoded Google Cloud Service Account used to access the TestGridBucket.
	TestGridServiceAccount []byte `env:"TESTGRID_SERVICE_ACCOUNT"`

	// UseProd sends requests to production OSD.
	//
	// Deprecated: Use OSD_ENV=prod instead.
	UseProd bool `env:"USE_PROD"`

	// MultiAZ deploys a cluster across multiple availability zones.
	MultiAZ bool `env:"MULTI_AZ"`

	// NoDestroy leaves the cluster running after testing.
	NoDestroy bool `env:"NO_DESTROY"`

	// NoTestGrid disables reporting to TestGrid.
	NoTestGrid bool `env:"NO_TESTGRID"`

	// Kubeconfig is used to access a cluster.
	Kubeconfig []byte `env:"TEST_KUBECONFIG"`

	// OSDEnv is the OpenShift Dedicated environment used to provision clusters.
	OSDEnv string `env:"OSD_ENV"`

	// DebugOSD shows debug level messages when enabled.
	DebugOSD bool `env:"DEBUG_OSD"`

	// CleanRuns is the number of times the test-version is run before skipping.
	CleanRuns int `env:"CLEAN_RUNS"`

	// UpgradeReleaseStream used to retrieve latest release images. If set, it will be used to perform an upgrade.
	UpgradeReleaseStream string `env:"UPGRADE_RELEASE_STREAM"`

	// UpgradeReleaseName is the name of the release in a release stream. UpgradeReleaseStream must be set.
	UpgradeReleaseName string `env:"UPGRADE_RELEASE_NAME"`

	// UpgradeImage is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.
	UpgradeImage string `env:"UPGRADE_IMAGE"`
}
