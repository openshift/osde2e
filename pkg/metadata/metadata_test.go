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

// generateExpected marshals/unmarshals a supplied metadata object.
// Rather than hand-curating our own expected data, this functionally
// achieves the same test result but it's cleaner.
func generateExpected(m *Metadata) map[string]interface{} {
	tmp, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return nil
	}

	data := make(map[string]interface{})

	err = json.Unmarshal(tmp, &data)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return nil
	}

	return data
}

func TestMetadata(t *testing.T) {
	tests := []struct {
		name string
		m    *Metadata
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
		},
		{
			name: "log metrics exists",
			m: &Metadata{
				ClusterID:                   "test-id",
				ClusterName:                 "test-name",
				ClusterVersion:              "test-version",
				Environment:                 "test-environment",
				UpgradeVersion:              "test-upgrade",
				TimeToOCMReportingInstalled: 123.45,
				TimeToClusterReady:          456.78,
				LogMetrics: map[string]int{
					"some-metric": 5,
				},
			},
		},
	}

	for _, test := range tests {
		if err := writeAndTestMetadata(test.m); err != nil {
			t.Errorf("%s: error while testing metadata: %v", test.name, err)
		}
	}
}

func writeAndTestMetadata(m *Metadata) (err error) {
	var tempDir string
	if tempDir, err = ioutil.TempDir("", ""); err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	if err = m.WriteToJSON(tempDir); err != nil {
		return err
	}

	var data []byte
	if data, err = ioutil.ReadFile(filepath.Join(tempDir, MetadataFile)); err != nil {
		return err
	}

	readData := map[string]interface{}{}
	if err = json.Unmarshal(data, &readData); err != nil {
		return err
	}

	if !reflect.DeepEqual(generateExpected(m), readData) {
		return fmt.Errorf("expected and generated JSON do not match")
	}

	return nil
}
