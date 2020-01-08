package metadata

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Metadata houses the metadata that will be written to the report directory after
type Metadata struct {
	// Cluster information
	ClusterID            string `json:"cluster-id"`
	ClusterName          string `json:"cluster-name"`
	ClusterVersion       string `json:"cluster-version"`
	Environment          string `json:"environment"`
	UpgradeVersion       string `json:"upgrade-version,omitempty"`
	UpgradeVersionSource string `json:"upgrade-version-source,omitempty"`

	// Metrics
	TimeToOCMReportingInstalled float64 `json:"time-to-ocm-reporting-installed,string"`
	TimeToClusterReady          float64 `json:"time-to-cluster-ready,string"`
}

// Instance is the global metadata instance
var Instance *Metadata

func init() {
	Instance = &Metadata{}
}

// WriteToJSON will marshall the metadata struct and write it into the given file.
func (m *Metadata) WriteToJSON(outputFilename string) (err error) {
	var data []byte
	if data, err = json.Marshal(m); err != nil {
		return err
	}

	if err = ioutil.WriteFile(outputFilename, data, os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}
