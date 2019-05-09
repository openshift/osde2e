// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	"os"
	"reflect"
)

// Cfg is the configuration used for end to end testing.
var Cfg = new(Config)

func init() {
	Cfg.LoadFromEnv()
}

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
	ClusterName string

	// ClusterVersion is the version of the cluster being deployed.
	ClusterVersion string `env:"CLUSTER_VERSION"`

	// AWSKeyID is used by OSD.
	AWSKeyID string `env:"AWS_ACCESS_KEY_ID"`

	// AWSAccessKey is used by OSD.
	AWSAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`

	// TestGridBucket is the Google Cloud Storage bucket where results are reported for TestGrid.
	TestGridBucket string `env:"TESTGRID_BUCKET"`

	// TestGridPrefix is used to namespace reports.
	TestGridPrefix string `env:"TESTGRID_PREFIX"`

	// TestGridServiceAccount is a Base64 encoded Google Cloud Service Account used to access the TestGridBucket.
	TestGridServiceAccount []byte `env:"TESTGRID_SERVICE_ACCOUNT"`

	// UseProd sends requests to production OSD.
	UseProd bool

	// NoDestroy leaves the cluster running after testing.
	NoDestroy bool `env:"NO_DESTROY"`

	// NoTestGrid disables reporting to TestGrid.
	NoTestGrid bool `env:"NO_TESTGRID"`

	// Kubeconfig is used to access a cluster.
	Kubeconfig []byte `env:"TEST_KUBECONFIG"`

	// DebugOSD shows debug level messages when enabled.
	DebugOSD bool `env:"DEBUG_OSD"`
}

// LoadFromEnv sets values from environment variables specified in `env` tags.
func (c *Config) LoadFromEnv() {
	v := reflect.ValueOf(c).Elem()
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		if env, ok := f.Tag.Lookup("env"); ok {
			if envVal, envOk := os.LookupEnv(env); envOk {
				field := v.Field(i)
				switch f.Type.Kind() {
				case reflect.String:
					field.SetString(envVal)
				case reflect.Bool:
					field.SetBool(true)
				case reflect.Slice:
					field.SetBytes([]byte(envVal))
				}
			}
		}
	}
}
