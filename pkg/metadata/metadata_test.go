package metadata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMetadata(t *testing.T) {
	tests := []struct {
		name     string
		m        *Metadata
		expected map[string]interface{}
	}{
		{
			name: "test all fields",
			m: &Metadata{
				ClusterID:                   "test-id",
				ClusterName:                 "test-name",
				ClusterVersion:              "test-version",
				Environment:                 "test-environment",
				UpgradeVersion:              "test-upgrade",
				TimeToOCMReportingInstalled: 123.45,
				TimeToClusterReady:          456.78,
			},
			expected: map[string]interface{}{
				"cluster-id":                      "test-id",
				"cluster-name":                    "test-name",
				"cluster-version":                 "test-version",
				"environment":                     "test-environment",
				"upgrade-version":                 "test-upgrade",
				"time-to-ocm-reporting-installed": "123.45",
				"time-to-cluster-ready":           "456.78",
			},
		},
		{
			name: "omit upgrade version",
			m: &Metadata{
				ClusterID:                   "test-id",
				ClusterName:                 "test-name",
				ClusterVersion:              "test-version",
				Environment:                 "test-environment",
				TimeToOCMReportingInstalled: 123.45,
				TimeToClusterReady:          456.78,
			},
			expected: map[string]interface{}{
				"cluster-id":                      "test-id",
				"cluster-name":                    "test-name",
				"cluster-version":                 "test-version",
				"environment":                     "test-environment",
				"time-to-ocm-reporting-installed": "123.45",
				"time-to-cluster-ready":           "456.78",
			},
		},
	}

	for _, test := range tests {
		if err := writeAndTestMetadata(test.m, test.expected); err != nil {
			t.Errorf("%s: error while testing metadata: %v", test.name, err)
		}
	}
}

func writeAndTestMetadata(m *Metadata, expected map[string]interface{}) (err error) {
	var tempDir string
	if tempDir, err = ioutil.TempDir("", ""); err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	outputFilename := filepath.Join(tempDir, "test.json")
	if err = m.WriteToJSON(outputFilename); err != nil {
		return err
	}

	var data []byte
	if data, err = ioutil.ReadFile(outputFilename); err != nil {
		return err
	}

	readData := map[string]interface{}{}
	if err = json.Unmarshal(data, &readData); err != nil {
		return err
	}

	if !reflect.DeepEqual(expected, readData) {
		return fmt.Errorf("expected and generated JSON do not match")
	}

	return nil
}
