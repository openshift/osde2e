package metadata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/phase"
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
	TimeToOCMReportingInstalled float64        `json:"time-to-ocm-reporting-installed,string"`
	TimeToClusterReady          float64        `json:"time-to-cluster-ready,string"`
	TimeToUpgradedCluster       float64        `json:"time-to-upgraded-cluster,string"`
	TimeToUpgradedClusterReady  float64        `json:"time-to-upgraded-cluster-ready,string"`
	InstallPhasePassRate        float64        `json:"install-phase-pass-rate,string"`
	UpgradePhasePassRate        float64        `json:"upgrade-phase-pass-rate,string"`
	LogMetrics                  map[string]int `json:"log-metrics"`
}

// Instance is the global metadata instance
var Instance *Metadata

func init() {
	Instance = &Metadata{}
	Instance.InstallPhasePassRate = -1.0
	Instance.UpgradePhasePassRate = -1.0
	Instance.LogMetrics = make(map[string]int)
}

// Next are a bunch of setter functions that allow us
// to track/trap changes to metadata and then flush
// the changes to a file.

// SetClusterID sets the cluster id
func (m *Metadata) SetClusterID(id string) {
	m.ClusterID = id
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetClusterName sets the cluster name
func (m *Metadata) SetClusterName(name string) {
	m.ClusterName = name
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetClusterVersion sets the cluster version
func (m *Metadata) SetClusterVersion(version string) {
	m.ClusterVersion = version
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetEnvironment sets the cluster environment
func (m *Metadata) SetEnvironment(env string) {
	m.Environment = env
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetUpgradeVersion sets the cluster upgrade version
func (m *Metadata) SetUpgradeVersion(ver string) {
	m.UpgradeVersion = ver
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetUpgradeVersionSource sets the cluster upgrade version source
func (m *Metadata) SetUpgradeVersionSource(src string) {
	m.UpgradeVersionSource = src
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetTimeToOCMReportingInstalled sets the time it took for OCM to report a cluster provisioned
func (m *Metadata) SetTimeToOCMReportingInstalled(timeToOCMReportingInstalled float64) {
	m.TimeToOCMReportingInstalled = timeToOCMReportingInstalled
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetTimeToClusterReady sets the time it took for the cluster to appear healthy on install
func (m *Metadata) SetTimeToClusterReady(timeToClusterReady float64) {
	m.TimeToClusterReady = timeToClusterReady
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetTimeToUpgradedCluster sets the time it took for the cluster to install an upgrade
func (m *Metadata) SetTimeToUpgradedCluster(timeToUpgradedCluster float64) {
	m.TimeToUpgradedCluster = timeToUpgradedCluster
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetTimeToUpgradedClusterReady sets the time it took for the cluster to appear healthy on upgrade
func (m *Metadata) SetTimeToUpgradedClusterReady(timeToUpgradedClusterReady float64) {
	m.TimeToUpgradedClusterReady = timeToUpgradedClusterReady
	m.WriteToJSON(config.Instance.ReportDir)
}

// SetPassRate sets the passrate metadata metric for the given phase
func (m *Metadata) SetPassRate(currentPhase string, passRate float64) {
	if currentPhase == phase.InstallPhase {
		m.InstallPhasePassRate = passRate
	} else if currentPhase == phase.UpgradePhase {
		m.UpgradePhasePassRate = passRate
	} else {
		// This is a developer issue, so this should fail ungracefully.
		panic(fmt.Sprintf("Invalid phase: %s, couldn't set pass rate.", currentPhase))
	}
}

// IncrementLogMetric adds a supplied number to a log metric or sets the metric to
// the value if it doesn't exist already
func (m *Metadata) IncrementLogMetric(metric string, value int) {
	if _, ok := m.LogMetrics[metric]; ok {
		m.LogMetrics[metric] += value
	} else {
		m.LogMetrics[metric] = value
	}
	m.WriteToJSON(config.Instance.ReportDir)
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
