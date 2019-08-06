package config

import (
	"reflect"

	testgrid "k8s.io/test-infra/testgrid/metadata"
)

var (
	// sensitiveFields removed from output
	sensitiveFields = []string{
		"UHC_TOKEN",
		"TESTGRID_SERVICE_ACCOUNT",
		"TEST_KUBECONFIG",
	}
)

// TestGrid returns a version of c suitable for reporting with any secrets removed.
func (c *Config) TestGrid() testgrid.Metadata {
	v := reflect.ValueOf(c).Elem()
	metadata := make(testgrid.Metadata, v.Type().NumField())
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		if env, ok := f.Tag.Lookup(EnvVarTag); ok {
			if isSensitive(env) {
				continue
			}

			field := v.Field(i)
			metadata[env] = field.Interface()
		}
	}
	return metadata
}

// returns true if sensitive config
func isSensitive(s string) bool {
	for _, sStr := range sensitiveFields {
		if s == sStr {
			return true
		}
	}
	return false
}
