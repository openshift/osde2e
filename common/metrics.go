package common

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo/reporters"
	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/events"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/state"
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
)

var junitFileRegex, logFileRegex *regexp.Regexp

// Metrics is the metrics object which can parse jUnit and JSON metadata and produce Prometheus metrics.
type Metrics struct {
	metricRegistry   *prometheus.Registry
	jUnitGatherer    *prometheus.GaugeVec
	metadataGatherer *prometheus.GaugeVec
	addonGatherer    *prometheus.GaugeVec
	eventGatherer    *prometheus.CounterVec
}

// NewMetrics creates a new metrics object using the given config object.
func NewMetrics() *Metrics {
	// Set up Prometheus metrics registry and gatherers
	metricRegistry := prometheus.NewRegistry()
	jUnitGatherer := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: jUnitMetricName,
		},
		[]string{"install_version", "upgrade_version", "environment", "phase", "suite", "testname", "result"},
	)
	metadataGatherer := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metadataMetricName,
		},
		[]string{"install_version", "upgrade_version", "environment", "metadata_name"},
	)
	addonGatherer := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: addonMetricName,
		},
		[]string{"install_version", "upgrade_version", "environment", "metadata_name", "phase"},
	)
	eventGatherer := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: eventMetricName,
		},
		[]string{"install_version", "upgrade_version", "environment", "event"},
	)
	metricRegistry.MustRegister(jUnitGatherer)
	metricRegistry.MustRegister(metadataGatherer)
	metricRegistry.MustRegister(addonGatherer)
	metricRegistry.MustRegister(eventGatherer)

	return &Metrics{
		metricRegistry:   metricRegistry,
		jUnitGatherer:    jUnitGatherer,
		metadataGatherer: metadataGatherer,
		addonGatherer:    addonGatherer,
		eventGatherer:    eventGatherer,
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

	prometheusFileName := fmt.Sprintf(prometheusFileNamePattern, state.Instance.Cluster.ID, config.Instance.JobName)
	output, err := m.registryToExpositionFormat()

	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filepath.Join(reportDir, prometheusFileName), output, os.FileMode(0644))
	if err != nil {
		return "", err
	}

	return prometheusFileName, nil
}

// jUnit file processing

// processJUnitXMLFile will add results to the prometheusOutput that look like:
//
// cicd_jUnitResult {environment="prod", install_version="install-version", result="passed|failed|skipped", phase="currentphase", suite="suitename",
//                   testname="testname", upgrade_version="upgrade-version"} testLength
func (m *Metrics) processJUnitXMLFile(phase string, junitFile string) (err error) {
	cfg := config.Instance
	state := state.Instance

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
		if testcase.FailureMessage != nil {
			result = "failed"
		} else if testcase.Skipped != nil {
			result = "skipped"
		} else {
			result = "passed"
		}

		m.jUnitGatherer.WithLabelValues(state.Cluster.Version, state.Upgrade.ReleaseName, cfg.OCM.Env, phase, testSuite.Name, testcase.Name, result).Add(testcase.Time)
	}

	return nil
}

// JSON file processing

// processJSONFile takes a JSON file and converts it into prometheus metrics of the general format:
//
// cicd_[addon_]metadata{environment="prod", install_version="install-version",
//                       metadata_name="full.path.to.field.separated.by.periiod",
//                       upgrade_version="upgrade-version"[, phase="install"]} userAssignedValue
//
// Notes: Only numerical values or strings that look like numerical values will be captured. This is because
//        Prometheus can only have numerical metric values and capturing strings through the use of labels is
//        of questionable value.
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
	cfg := config.Instance
	state := state.Instance

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
			if floatValue, err := strconv.ParseFloat(stringValue, 64); err == nil {
				if phase != "" {
					gatherer.WithLabelValues(state.Cluster.Version, state.Upgrade.ReleaseName, cfg.OCM.Env, metadataName, phase).Add(floatValue)
				} else {
					gatherer.WithLabelValues(state.Cluster.Version, state.Upgrade.ReleaseName, cfg.OCM.Env, metadataName).Add(floatValue)
				}
			}
		}
	}
}

// Event processing

// processEvents will search the events list for events that have occurred over the osde2e run
// and output them in the Prometheus metrics.
func (m *Metrics) processEvents(gatherer *prometheus.CounterVec) {
	cfg := config.Instance
	state := state.Instance

	for _, event := range events.GetListOfEvents() {
		gatherer.WithLabelValues(state.Cluster.Version, state.Upgrade.ReleaseName, cfg.OCM.Env, event).Inc()
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
