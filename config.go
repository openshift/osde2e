package osde2e

import (
	"os"
	"reflect"

	"github.com/openshift/osde2e/pkg/cluster"
)

// Cfg is the configuration being used for end to end testing.
var Cfg = new(Config)

// Config dictates the behavior of cluster tests.
type Config struct {
	// ReportDir is the location JUnit XML results are written.
	ReportDir string `env:"REPORT_DIR"`

	// Prefix is used at the beginning of tests to identify them.
	Prefix string

	// UHCToken is used to authenticate with UHC.
	UHCToken string `env:"UHC_TOKEN"`

	// ClusterName is the name of the cluster being created.
	ClusterName string

	// AWSKeyId is used by UHC.
	AWSKeyId string `env:"AWS_ACCESS_KEY_ID"`

	// AWSAccessKey is used by UHC.
	AWSAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`

	// UseProd sends requests to production UHC.
	UseProd bool

	// runtime vars
	clusterId  string
	uhc        *cluster.UHC
	kubeconfig []byte
}

func (c *Config) LoadFromEnv() {
	v := reflect.ValueOf(c).Elem()
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		if env, ok := f.Tag.Lookup("env"); ok {
			if envVal, envOk := os.LookupEnv(env); envOk {
				switch f.Type.Kind() {
				case reflect.String:
					v.Field(i).SetString(envVal)
				case reflect.Bool:
					v.Field(i).SetBool(true)
				}
			}
		}
	}
}
