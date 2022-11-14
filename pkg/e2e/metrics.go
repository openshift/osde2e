package e2e

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo/v2/reporters"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

const (
	prometheusFileNamePattern string = "%s.%s.metrics.prom"

	cicdPrefix string = "cicd_"

	jUnitMetricName    string = cicdPrefix + "jUnitResult"
	metadataMetricName string = cicdPrefix + "metadata"
	addonMetricName    string = cicdPrefix + "addon_metadata"
	eventMetricName    string = cicdPrefix + "event"
	routeMetricName    string = cicdPrefix + "route"
)

var junitFileRegex, logFileRegex *regexp.Regexp

// Metrics is the metrics object which can parse jUnit and JSON metadata and produce Prometheus metrics.
type Metrics struct {
	metricRegistry   *prometheus.Registry
	jUnitGatherer    *prometheus.GaugeVec
	metadataGatherer *prometheus.GaugeVec
	addonGatherer    *prometheus.GaugeVec
	eventGatherer    *prometheus.CounterVec
	routeGatherer    *prometheus.GaugeVec

	// Provider for getting metrics data
	provider spi.Provider
}

// NewMetrics creates a new metrics object using the given config object.
func NewMetrics() *Metrics {
	// Set up Prometheus metrics registry and gatherers
	metricRegistry := prometheus.NewRegistry()
	jUnitGatherer := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: jUnitMetricName,
		},
		[]string{"install_version", "upgrade_version", "cloud_provider", "environment", "region", "phase", "suite", "testname", "result", "cluster_id", "job_id"},
	)
	metadataGatherer := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metadataMetricName,
		},
		[]string{"install_version", "upgrade_version", "cloud_provider", "environment", "region", "metadata_name", "cluster_id", "job_id"},
	)
	addonGatherer := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: addonMetricName,
		},
		[]string{"install_version", "upgrade_version", "cloud_provider", "environment", "region", "metadata_name", "cluster_id", "job_id", "phase"},
	)
	eventGatherer := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: eventMetricName,
		},
		[]string{"install_version", "upgrade_version", "cloud_provider", "environment", "region", "event", "cluster_id", "job_id"},
	)
	routeGatherer := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: routeMetricName,
		},
		[]string{"install_version", "upgrade_version", "cloud_provider", "environment", "region", "cluster_id", "job_id", "type", "route"},
	)
	metricRegistry.MustRegister(jUnitGatherer)
	metricRegistry.MustRegister(metadataGatherer)
	metricRegistry.MustRegister(addonGatherer)
	metricRegistry.MustRegister(eventGatherer)
	metricRegistry.MustRegister(routeGatherer)

	provider, err := providers.ClusterProvider()
	if err != nil {
		log.Printf("unable to get provider for metrics, failing: %v", err)
		return nil
	}

	return &Metrics{
		metricRegistry:   metricRegistry,
		jUnitGatherer:    jUnitGatherer,
		metadataGatherer: metadataGatherer,
		addonGatherer:    addonGatherer,
		eventGatherer:    eventGatherer,
		routeGatherer:    routeGatherer,
		provider:         provider,
	}
}

func init() {
	junitFileRegex = regexp.MustCompile("^junit.*\\.xml$")
	logFileRegex = regexp.MustCompile("^.*\\.(log|txt)$")
}

// WritePrometheusFile collects data and writes it out in the prometheus export file format (https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md)
// Returns the prometheus file name.
func (m *Metrics) WritePrometheusFile(reportDir string) (string, error) {
	files, err := ioutil.ReadDir(reportDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file != nil {
			// This directory name is the name of the current phase we're in, so record it and iterate through these results
			if file.IsDir() {
				phase := file.Name()
				phaseDir := filepath.Join(reportDir, phase)
				phaseFiles, err := ioutil.ReadDir(phaseDir)
				if err != nil {
					return "", err
				}

				for _, phaseFile := range phaseFiles {
					// Process the jUnit XML result files
					if junitFileRegex.MatchString(phaseFile.Name()) {
						// TODO: The addon metric prefix should reference the addon job being run to further avoid collision
						m.processJUnitXMLFile(phase, filepath.Join(phaseDir, phaseFile.Name()))
					} else if phaseFile.Name() == metadata.AddonMetadataFile {
						m.processJSONFile(m.addonGatherer, filepath.Join(phaseDir, phaseFile.Name()), phase)
					}
				}
			} else if file.Name() == metadata.CustomMetadataFile {
				m.processJSONFile(m.metadataGatherer, filepath.Join(reportDir, file.Name()), "")
			}
		}
	}

	m.processEvents(m.eventGatherer)
	m.processRoutes(m.routeGatherer)

	prometheusFileName := fmt.Sprintf(prometheusFileNamePattern, viper.GetString(config.Cluster.ID), viper.GetString(config.JobName))
	output, err := m.registryToExpositionFormat()
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filepath.Join(reportDir, prometheusFileName), output, os.FileMode(0o644))
	if err != nil {
		return "", err
	}

	return prometheusFileName, nil
}

// jUnit file processing

// processJUnitXMLFile will add results to the prometheusOutput that look like:
//
// cicd_jUnitResult {environment="prod", install_version="install-version",
// result="passed|failed|skipped", phase="currentphase",
// suite="suitename", testname="testname", upgrade_version="upgrade-version"}
func (m *Metrics) processJUnitXMLFile(phase string, junitFile string) (err error) {
	data, err := ioutil.ReadFile(junitFile)
	if err != nil {
		return err
	}

	// Use Ginkgo's JUnitTestSuite to unmarshal the JUnit XML file
	var testSuite reporters.JUnitTestSuite

	if err = xml.Unmarshal(data, &testSuite); err != nil {
		return err
	}

	for _, testcase := range testSuite.TestCases {
		var result string
		if testcase.Failure != nil {
			result = "failed"
		} else if testcase.Skipped != nil {
			result = "skipped"
		} else {
			result = "passed"
		}

		m.jUnitGatherer.WithLabelValues(viper.GetString(config.Cluster.Version),
			viper.GetString(config.Upgrade.ReleaseName),
			viper.GetString(config.CloudProvider.CloudProviderID),
			m.provider.Environment(),
			viper.GetString(config.CloudProvider.Region),
			phase,
			testSuite.Name,
			testcase.Name,
			result,
			viper.GetString(config.Cluster.ID),
			strconv.Itoa(viper.GetInt(config.JobID))).Add(testcase.Time)
	}

	return nil
}

// JSON file processing

// processJSONFile takes a JSON file and converts it into prometheus metrics of the general format:
//
// cicd_[addon_]metadata{environment="prod", install_version="install-version",
// metadata_name="full.path.to.field.separated.by.periiod",
// upgrade_version="upgrade-version"[, phase="install"]} userAssignedValue
//
// Notes: Only numerical values or strings that look like numerical values will
// be captured. This is because Prometheus can only have numerical metric
// values and capturing strings through the use of labels is of questionable
// value.
func (m *Metrics) processJSONFile(gatherer *prometheus.GaugeVec, jsonFile string, phase string) (err error) {
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return err
	}

	var jsonOutput interface{}

	if err = json.Unmarshal(data, &jsonOutput); err != nil {
		return err
	}

	m.jsonToPrometheusOutput(gatherer, phase, jsonOutput.(map[string]interface{}), []string{})

	return nil
}

// jsonToPrometheusOutput will take the JSON and write it into the gauge vector.
func (m *Metrics) jsonToPrometheusOutput(gatherer *prometheus.GaugeVec, phase string, jsonOutput map[string]interface{}, context []string) {
	for k, v := range jsonOutput {
		fullContext := append(context, k)
		switch jsonObject := v.(type) {
		case map[string]interface{}:
			m.jsonToPrometheusOutput(gatherer, phase, jsonObject, fullContext)
		default:
			metadataName := strings.Join(fullContext, ".")

			// Due to current limitations in the Spyglass metadata lens, all fields in a metadata
			// object need to be strings to display properly. Rather than look for float64 types, we'll
			// try to parse the float out of the JSON value. That way, if the number is in a string object
			// we'll still be able to use it numerically instead of shoving it into the label as we'd
			// otherwise have to do.
			stringValue := fmt.Sprintf("%v", jsonObject)

			// We're only concerned with tracking float values in Prometheus as they're the only thing we can measure
			jobID := viper.GetInt(config.JobID)
			if floatValue, err := strconv.ParseFloat(stringValue, 64); err == nil {
				if phase != "" {
					gatherer.WithLabelValues(viper.GetString(config.Cluster.Version),
						viper.GetString(config.Upgrade.ReleaseName),
						viper.GetString(config.CloudProvider.CloudProviderID),
						m.provider.Environment(),
						viper.GetString(config.CloudProvider.Region),
						metadataName,
						viper.GetString(config.Cluster.ID),
						strconv.Itoa(jobID),
						phase).Add(floatValue)
				} else {
					gatherer.WithLabelValues(viper.GetString(config.Cluster.Version),
						viper.GetString(config.Upgrade.ReleaseName),
						viper.GetString(config.CloudProvider.CloudProviderID),
						m.provider.Environment(),
						viper.GetString(config.CloudProvider.Region),
						metadataName,
						viper.GetString(config.Cluster.ID),
						strconv.Itoa(jobID)).Add(floatValue)
				}
			}
		}
	}
}

// Event processing

// processEvents will search the events list for events that have occurred over the osde2e run
// and output them in the Prometheus metrics.
func (m *Metrics) processEvents(gatherer *prometheus.CounterVec) {
	for _, event := range events.GetListOfEvents() {
		gatherer.WithLabelValues(
			viper.GetString(config.Cluster.Version),
			viper.GetString(config.Upgrade.ReleaseName),
			viper.GetString(config.CloudProvider.CloudProviderID),
			m.provider.Environment(),
			viper.GetString(config.CloudProvider.Region),
			event,
			viper.GetString(config.Cluster.ID),
			strconv.Itoa(viper.GetInt(config.JobID))).Inc()
	}
}

// Route processing
// processRoutes will search the events list for events that have occurred over the osde2e run
// and output them in the Prometheus metrics.
func (m *Metrics) processRoutes(gatherer *prometheus.GaugeVec) {
	for route, latency := range metadata.Instance.RouteLatencies {
		log.Printf("Gathering %s/latency: %v", route, latency)
		if !math.IsNaN(latency) {
			gatherer.WithLabelValues(
				viper.GetString(config.Cluster.Version),
				viper.GetString(config.Upgrade.ReleaseName),
				viper.GetString(config.CloudProvider.CloudProviderID),
				m.provider.Environment(),
				viper.GetString(config.CloudProvider.Region),
				viper.GetString(config.Cluster.ID),
				strconv.Itoa(viper.GetInt(config.JobID)),
				route, "latency").Set(latency)
		}
	}
	for route, availability := range metadata.Instance.RouteAvailabilities {
		log.Printf("Gathering %s/availability: %v", route, availability)
		if !math.IsNaN(availability) {
			gatherer.WithLabelValues(
				viper.GetString(config.Cluster.Version),
				viper.GetString(config.Upgrade.ReleaseName),
				viper.GetString(config.CloudProvider.CloudProviderID),
				m.provider.Environment(),
				viper.GetString(config.CloudProvider.Region),
				viper.GetString(config.Cluster.ID),
				strconv.Itoa(viper.GetInt(config.JobID)),
				route, "availability").Set(availability)
		}
	}
	for route, throughput := range metadata.Instance.RouteThroughputs {
		log.Printf("Gathering %s/throughput: %v", route, throughput)
		if !math.IsNaN(throughput) {
			gatherer.WithLabelValues(
				viper.GetString(config.Cluster.Version),
				viper.GetString(config.Upgrade.ReleaseName),
				viper.GetString(config.CloudProvider.CloudProviderID),
				m.provider.Environment(),
				viper.GetString(config.CloudProvider.Region),
				viper.GetString(config.Cluster.ID),
				strconv.Itoa(viper.GetInt(config.JobID)),
				route, "throughput").Set(throughput)
		}
	}
}

// Generic Prometheus export file building functions

// registryToExpositionFormat takes all of the gathered metrics and writes them out in the exposition format
func (m *Metrics) registryToExpositionFormat() ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := expfmt.NewEncoder(buf, expfmt.FmtText)
	metricFamilies, err := m.metricRegistry.Gather()
	if err != nil {
		return []byte{}, fmt.Errorf("error while gathering metrics: %v", err)
	}

	for _, metricFamily := range metricFamilies {
		if err := encoder.Encode(metricFamily); err != nil {
			return []byte{}, fmt.Errorf("error encoding metric family: %v", err)
		}
	}

	return buf.Bytes(), nil
}
