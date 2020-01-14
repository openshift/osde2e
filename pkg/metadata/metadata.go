package metadata

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// CustomMetadataFile is the name of the custom metadata file generated for spyglass visualization.
	CustomMetadataFile string = "custom-prow-metadata.json"

	// MetadataFile is the name of the custom metadata file generated for spyglass visualization.
	MetadataFile string = "metadata.json"

	// AddonMetadataFile he name of the addon metadata file
	AddonMetadataFile string = "addon-metadata.json"
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
func (m *Metadata) WriteToJSON(reportDir string) (err error) {
	var data []byte
	if data, err = json.Marshal(m); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(reportDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file != nil {
			// This directory name is the name of the current phase we're in, so record it and iterate through these results
			if file.IsDir() {
				phase := file.Name()
				phaseDir := filepath.Join(reportDir, phase)
				phaseFiles, err := ioutil.ReadDir(phaseDir)
				if err != nil {
					return err
				}

				for _, phaseFile := range phaseFiles {
					// Process the jUnit XML result files
					if phaseFile.Name() == AddonMetadataFile {
						// Unmarshal raw metadata to map
						var rawMetadataJSON = map[string]interface{}{}
						if err := json.Unmarshal(data, &rawMetadataJSON); err != nil {
							return err
						}

						// Unmarshal addon metadata to map
						addonData, err := ioutil.ReadFile(filepath.Join(phaseDir, phaseFile.Name()))
						if err != nil {
							return err
						}

						var rawAddonMetadataJSON = map[string]interface{}{}
						if err := json.Unmarshal(addonData, &rawAddonMetadataJSON); err != nil {
							return err
						}

						rawMetadataJSON["addon."+phase] = rawAddonMetadataJSON

						if data, err = json.Marshal(rawMetadataJSON); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	if err = ioutil.WriteFile(filepath.Join(reportDir, CustomMetadataFile), data, os.FileMode(0644)); err != nil {
		return err
	}

	if err = ioutil.WriteFile(filepath.Join(reportDir, MetadataFile), data, os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}
