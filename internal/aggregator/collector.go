// Package aggregator provides functionality to collect artifacts and metadata
// from osde2e test runs for LLM analysis.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/joshdk/go-junit"
)

// artifactCollector collects test failure artifacts
type artifactCollector struct {
	logger logr.Logger
}

// AggregatedData contains all collected artifacts and metadata
type AggregatedData struct {
	Metadata       map[string]any    `json:"metadata"`
	TestResults    TestResultSummary `json:"testResults"`
	FailedTests    []FailedTest      `json:"failedTests"`
	Logs           []LogEntry        `json:"logs"`
	CollectionTime time.Time         `json:"collectionTime"`
}

func (a *AggregatedData) SetMetadata(metadata map[string]any) {
	a.Metadata = metadata
}

// TestResultSummary provides high-level test execution statistics
type TestResultSummary struct {
	TotalTests   int           `json:"totalTests"`
	PassedTests  int           `json:"passedTests"`
	FailedTests  int           `json:"failedTests"`
	SkippedTests int           `json:"skippedTests"`
	ErrorTests   int           `json:"errorTests"`
	Duration     time.Duration `json:"duration"`
	SuiteCount   int           `json:"suiteCount"`
}

// FailedTest contains details about a specific test failure
type FailedTest struct {
	Name       string        `json:"name"`
	ClassName  string        `json:"className,omitempty"`
	SuiteName  string        `json:"suiteName,omitempty"`
	Duration   time.Duration `json:"duration"`
	ErrorMsg   string        `json:"errorMessage,omitempty"`
	StackTrace string        `json:"stackTrace,omitempty"`
	SystemOut  string        `json:"systemOut,omitempty"`
	SystemErr  string        `json:"systemErr,omitempty"`
}

// LogEntry represents a collected log file
type LogEntry struct {
	Source string `json:"source"` // File path or source identifier
}

// newArtifactCollector creates a new artifact collector
func newArtifactCollector(logger logr.Logger) *artifactCollector {
	return &artifactCollector{
		logger: logger,
	}
}

// collectFromReportDir collects artifacts from the specified report directory
func (a *artifactCollector) collectFromReportDir(reportDir string) (*AggregatedData, error) {
	a.logger.Info("collecting artifacts", "reportDir", reportDir)

	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("report directory does not exist: %s", reportDir)
	}

	data := &AggregatedData{
		CollectionTime: time.Now(),
	}

	if err := a.collectLogs(reportDir, data); err != nil {
		a.logger.Error(err, "failed to collect logs")
	}

	if err := a.collectTestResults(data); err != nil {
		a.logger.Error(err, "failed to collect test results")
	}

	a.logger.Info("completed artifact collection",
		"failedTests", len(data.FailedTests),
		"logEntries", len(data.Logs))

	return data, nil
}

// collectTestResults processes JUnit XML files to extract test failure information
func (a *artifactCollector) collectTestResults(data *AggregatedData) error {
	junitFiles, err := a.findJUnitFiles(data)
	if err != nil {
		return fmt.Errorf("finding junit files: %w", err)
	}

	if len(junitFiles) == 0 {
		a.logger.Info("no junit files found")
		return nil
	}

	var allSuites []junit.Suite
	for _, file := range junitFiles {
		suites, err := junit.IngestFile(file)
		if err != nil {
			a.logger.Error(err, "failed to parse junit file", "file", file)
			continue
		}
		allSuites = append(allSuites, suites...)
	}

	// Calculate summary statistics
	summary := TestResultSummary{SuiteCount: len(allSuites)}
	var failedTests []FailedTest

	for _, suite := range allSuites {
		for _, test := range suite.Tests {
			summary.TotalTests++
			summary.Duration += test.Duration

			switch test.Status {
			case junit.StatusPassed:
				summary.PassedTests++
			case junit.StatusFailed:
				summary.FailedTests++
				failedTests = append(failedTests, a.convertJUnitTest(test, suite.Name))
			case junit.StatusSkipped:
				summary.SkippedTests++
			case junit.StatusError:
				summary.ErrorTests++
				failedTests = append(failedTests, a.convertJUnitTest(test, suite.Name))
			}
		}
	}

	data.TestResults = summary
	data.FailedTests = failedTests

	return nil
}

// convertJUnitTest converts a junit.Test to our FailedTest structure
func (a *artifactCollector) convertJUnitTest(test junit.Test, suiteName string) FailedTest {
	failed := FailedTest{
		Name:      test.Name,
		ClassName: test.Classname,
		SuiteName: suiteName,
		Duration:  test.Duration,
		SystemOut: test.SystemOut,
		SystemErr: test.SystemErr,
	}

	if test.Error != nil {
		failed.ErrorMsg = test.Error.Error()
		failed.StackTrace = test.Error.Error()
	}

	return failed
}

// collectLogs recursively iterates through the report directory and collects all file names
func (a *artifactCollector) collectLogs(reportDir string, data *AggregatedData) error {
	return filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log the error but continue walking
			a.logger.Info("error accessing path", "path", path, "error", err)
			return nil
		}

		// Skip directories that start with "."
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Skip directories, only collect files
		if info.IsDir() {
			return nil
		}

		// Add the file path to the logs
		data.Logs = append(data.Logs, LogEntry{
			Source: path,
		})

		return nil
	})
}

// findJUnitFiles searches for JUnit XML files from collected log entries
func (a *artifactCollector) findJUnitFiles(data *AggregatedData) ([]string, error) {
	var junitFiles []string

	// Filter the already collected log entries for JUnit XML files
	for _, logEntry := range data.Logs {
		fileName := strings.ToLower(filepath.Base(logEntry.Source))

		// Look for XML files that might be JUnit reports
		if strings.HasSuffix(fileName, ".xml") &&
			strings.Contains(fileName, "junit") {
			junitFiles = append(junitFiles, logEntry.Source)
		}
	}

	return junitFiles, nil
}
