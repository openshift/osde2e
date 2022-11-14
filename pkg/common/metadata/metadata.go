package metadata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

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
	Region               string `json:"region"`
	UpgradeVersion       string `json:"upgrade-version,omitempty"`
	UpgradeVersionSource string `json:"upgrade-version-source,omitempty"`

	// Metrics
	TimeToOCMReportingInstalled float64            `json:"time-to-ocm-reporting-installed,string"`
	TimeToClusterReady          float64            `json:"time-to-cluster-ready,string"`
	TimeToUpgradedCluster       float64            `json:"time-to-upgraded-cluster,string"`
	TimeToUpgradedClusterReady  float64            `json:"time-to-upgraded-cluster-ready,string"`
	TimeToCertificateIssued     float64            `json:"time-to-certificate-issued,string"`
	InstallPhasePassRate        float64            `json:"install-phase-pass-rate,string"`
	UpgradePhasePassRate        float64            `json:"upgrade-phase-pass-rate,string"`
	LogMetrics                  map[string]int     `json:"log-metrics"`
	BeforeSuiteMetrics          map[string]int     `json:"before-suite-metrics"`
	RouteLatencies              map[string]float64 `json:"route-latencies"`
	RouteThroughputs            map[string]float64 `json:"route-throughputs"`
	RouteAvailabilities         map[string]float64 `json:"route-availabilities"`

	// Real Time Data
	HealthChecks         map[string][]string `json:"healthchecks"`
	HealthCheckIteration float64             `json:"healthcheckIteration"`
	Status               string              `json:"status"`

	// Internal variables
	ReportDir string `json:"-"`
}

// Instance is the global metadata instance
var Instance *Metadata

func init() {
	Instance = &Metadata{}
	Instance.InstallPhasePassRate = -1.0
	Instance.UpgradePhasePassRate = -1.0
	Instance.LogMetrics = make(map[string]int)
	Instance.BeforeSuiteMetrics = make(map[string]int)
	Instance.RouteLatencies = make(map[string]float64)
	Instance.RouteThroughputs = make(map[string]float64)
	Instance.RouteAvailabilities = make(map[string]float64)
	Instance.HealthChecks = make(map[string][]string)
}

// Next are a bunch of setter functions that allow us
// to track/trap changes to metadata and then flush
// the changes to a file.

// SetReportDir sets the report dir for the metadata.
func (m *Metadata) SetReportDir(reportDir string) {
	m.ReportDir = reportDir
}

// SetClusterID sets the cluster id
func (m *Metadata) SetClusterID(id string) {
	m.ClusterID = id
	m.WriteToJSON(m.ReportDir)
}

// SetClusterName sets the cluster name
func (m *Metadata) SetClusterName(name string) {
	m.ClusterName = name
	m.WriteToJSON(m.ReportDir)
}

// SetClusterVersion sets the cluster version
func (m *Metadata) SetClusterVersion(version string) {
	m.ClusterVersion = version
	m.WriteToJSON(m.ReportDir)
}

// SetEnvironment sets the cluster environment
func (m *Metadata) SetEnvironment(env string) {
	m.Environment = env
	m.WriteToJSON(m.ReportDir)
}

// SetRegion sets the cluster environment
func (m *Metadata) SetRegion(region string) {
	m.Region = region
	m.WriteToJSON(m.ReportDir)
}

// SetUpgradeVersion sets the cluster upgrade version
func (m *Metadata) SetUpgradeVersion(ver string) {
	m.UpgradeVersion = ver
	m.WriteToJSON(m.ReportDir)
}

// SetUpgradeVersionSource sets the cluster upgrade version source
func (m *Metadata) SetUpgradeVersionSource(src string) {
	m.UpgradeVersionSource = src
	m.WriteToJSON(m.ReportDir)
}

// SetTimeToOCMReportingInstalled sets the time it took for OCM to report a cluster provisioned
func (m *Metadata) SetTimeToOCMReportingInstalled(timeToOCMReportingInstalled float64) {
	m.TimeToOCMReportingInstalled = timeToOCMReportingInstalled
	m.WriteToJSON(m.ReportDir)
}

// SetTimeToClusterReady sets the time it took for the cluster to appear healthy on install
func (m *Metadata) SetTimeToClusterReady(timeToClusterReady float64) {
	m.TimeToClusterReady = timeToClusterReady
	m.WriteToJSON(m.ReportDir)
}

// SetTimeToUpgradedCluster sets the time it took for the cluster to install an upgrade
func (m *Metadata) SetTimeToUpgradedCluster(timeToUpgradedCluster float64) {
	m.TimeToUpgradedCluster = timeToUpgradedCluster
	m.WriteToJSON(m.ReportDir)
}

// SetTimeToUpgradedClusterReady sets the time it took for the cluster to appear healthy on upgrade
func (m *Metadata) SetTimeToUpgradedClusterReady(timeToUpgradedClusterReady float64) {
	m.TimeToUpgradedClusterReady = timeToUpgradedClusterReady
	m.WriteToJSON(m.ReportDir)
}

// SetTimeToCertificateIssued sets the time it took for a certificate to be issued to the cluster
func (m *Metadata) SetTimeToCertificateIssued(timeToCertificateIssued float64) {
	m.TimeToCertificateIssued = timeToCertificateIssued
	m.WriteToJSON(m.ReportDir)
}

// SetHealthcheckValue sets an arbitrary string value to a healthcheck
func (m *Metadata) SetHealthcheckValue(key string, value []string) {
	if !reflect.DeepEqual(m.HealthChecks[key], value) {
		m.HealthChecks[key] = value
		m.WriteToJSON(m.ReportDir)
	}
}

// ClearHealthcheckValue removes a pending healthcheck
func (m *Metadata) ClearHealthcheckValue(key string) {
	if _, ok := m.HealthChecks[key]; ok {
		delete(m.HealthChecks, key)
		m.WriteToJSON(m.ReportDir)
	}
}

// IncrementHealthcheckIteration increments the healthcheck counter
func (m *Metadata) IncrementHealthcheckIteration() {
	m.HealthCheckIteration++
	m.WriteToJSON(m.ReportDir)
}

// ZeroHealthcheckIteration zeroes out the healthcheck counter
func (m *Metadata) ZeroHealthcheckIteration() {
	m.HealthCheckIteration = 0
	m.WriteToJSON(m.ReportDir)
}

// SetStatus stores the status of an osde2e cluster
func (m *Metadata) SetStatus(status string) {
	m.Status = status
	m.WriteToJSON(m.ReportDir)
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

// ResetLogMetrics zeroes out old results to be used before a new run.
func (m *Metadata) ResetLogMetrics() {
	for metric := range m.LogMetrics {
		m.LogMetrics[metric] = 0
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

	m.WriteToJSON(m.ReportDir)
}

// ResetBeforeSuiteMetrics zeroes out old results to be used before a new run.
func (m *Metadata) ResetBeforeSuiteMetrics() {
	for metric := range m.BeforeSuiteMetrics {
		m.BeforeSuiteMetrics[metric] = 0
	}
}

// IncrementBeforeSuiteMetric adds a supplied number to a before suite metric or sets the metric to
// the value if it doesn't exist already
func (m *Metadata) IncrementBeforeSuiteMetric(metric string, value int) {
	if _, ok := m.BeforeSuiteMetrics[metric]; ok {
		m.BeforeSuiteMetrics[metric] += value
	} else {
		m.BeforeSuiteMetrics[metric] = value
	}

	m.WriteToJSON(m.ReportDir)
}

// SetRouteLatency sets the mean latency for the given route
// (measured in milliseconds)
func (m *Metadata) SetRouteLatency(route string, latency float64) {
	m.RouteLatencies[route] = latency
	m.WriteToJSON(m.ReportDir)
}

// SetRouteThroughput sets the throughput for the given route
// (rate of successful requests per second)
func (m *Metadata) SetRouteThroughput(route string, throughput float64) {
	m.RouteThroughputs[route] = throughput
	m.WriteToJSON(m.ReportDir)
}

// SetRouteAvailability sets the availability for the given route
// (ratio of successful requests)
func (m *Metadata) SetRouteAvailability(route string, availability float64) {
	m.RouteAvailabilities[route] = availability
	m.WriteToJSON(m.ReportDir)
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
						rawMetadataJSON := map[string]interface{}{}
						if err := json.Unmarshal(data, &rawMetadataJSON); err != nil {
							return err
						}

						// Unmarshal addon metadata to map
						addonData, err := ioutil.ReadFile(filepath.Join(phaseDir, phaseFile.Name()))
						if err != nil {
							return err
						}

						rawAddonMetadataJSON := map[string]interface{}{}
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

	if err = ioutil.WriteFile(filepath.Join(reportDir, CustomMetadataFile), data, os.FileMode(0o644)); err != nil {
		return err
	}

	if err = ioutil.WriteFile(filepath.Join(reportDir, MetadataFile), data, os.FileMode(0o644)); err != nil {
		return err
	}

	return nil
}
