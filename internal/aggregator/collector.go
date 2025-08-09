// Package aggregator provides functionality to collect artifacts and metadata
// from osde2e test runs for LLM analysis.
package aggregator

import (
	"context"
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
	Metadata       ClusterMetadata   `json:"metadata"`
	TestResults    TestResultSummary `json:"testResults"`
	FailedTests    []FailedTest      `json:"failedTests"`
	ClusterLogs    []LogEntry        `json:"clusterLogs"`
	MustGatherData []LogEntry        `json:"mustGatherData"`
	BuildLogs      []LogEntry        `json:"buildLogs"`
	CollectionTime time.Time         `json:"collectionTime"`
}

// ClusterMetadata contains essential cluster information
type ClusterMetadata struct {
	ClusterID      string            `json:"clusterId,omitempty"`
	ClusterVersion string            `json:"clusterVersion,omitempty"`
	Provider       string            `json:"provider,omitempty"`
	CloudProvider  string            `json:"cloudProvider,omitempty"`
	Region         string            `json:"region,omitempty"`
	JobName        string            `json:"jobName,omitempty"`
	JobID          string            `json:"jobId,omitempty"`
	Phase          string            `json:"phase,omitempty"`
	Environment    string            `json:"environment,omitempty"`
	Properties     map[string]string `json:"properties,omitempty"`
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
	Source  string `json:"source"`  // File path or source identifier
	Content string `json:"content"` // Log content
}

// newArtifactCollector creates a new artifact collector
func newArtifactCollector(logger logr.Logger) *artifactCollector {
	return &artifactCollector{
		logger: logger,
	}
}

// collectFromReportDir collects artifacts from the specified report directory
func (a *artifactCollector) collectFromReportDir(ctx context.Context, reportDir string) (*AggregatedData, error) {
	a.logger.Info("collecting artifacts", "reportDir", reportDir)

	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("report directory does not exist: %s", reportDir)
	}

	data := &AggregatedData{
		CollectionTime: time.Now(),
	}

	// Collect test results from JUnit files
	if err := a.collectTestResults(reportDir, data); err != nil {
		a.logger.Error(err, "failed to collect test results")
	}

	// Collect build logs
	if err := a.collectBuildLogs(reportDir, data); err != nil {
		a.logger.Error(err, "failed to collect build logs")
	}

	// Collect cluster logs
	if err := a.collectClusterLogs(reportDir, data); err != nil {
		a.logger.Error(err, "failed to collect cluster logs")
	}

	// Collect must-gather data
	if err := a.collectMustGatherData(reportDir, data); err != nil {
		a.logger.Error(err, "failed to collect must-gather data")
	}

	a.logger.Info("completed artifact collection",
		"failedTests", len(data.FailedTests),
		"logEntries", len(data.ClusterLogs)+len(data.MustGatherData)+len(data.BuildLogs))

	return data, nil
}

// collectTestResults processes JUnit XML files to extract test failure information
func (a *artifactCollector) collectTestResults(reportDir string, data *AggregatedData) error {
	junitFiles, err := a.findJUnitFiles(reportDir)
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

// collectBuildLogs collects build logs from the report directory
func (a *artifactCollector) collectBuildLogs(reportDir string, data *AggregatedData) error {
	buildLogPath := filepath.Join(reportDir, "build-log.txt")
	if content, err := a.readLogFile(buildLogPath); err == nil {
		data.BuildLogs = append(data.BuildLogs, content)
	}
	return nil
}

// collectClusterLogs collects cluster-related logs
func (a *artifactCollector) collectClusterLogs(reportDir string, data *AggregatedData) error {
	clusterLogsDir := filepath.Join(reportDir, "cluster-logs")
	return a.collectLogsFromDir(clusterLogsDir, &data.ClusterLogs)
}

// collectMustGatherData collects must-gather diagnostic data
func (a *artifactCollector) collectMustGatherData(reportDir string, data *AggregatedData) error {
	mustGatherDir := filepath.Join(reportDir, "must-gather")
	return a.collectLogsFromDir(mustGatherDir, &data.MustGatherData)
}

// collectLogsFromDir recursively collects log files from a directory
func (a *artifactCollector) collectLogsFromDir(dir string, logs *[]LogEntry) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if info.IsDir() {
			return nil
		}

		// Only collect log files and important text files
		if a.isLogFile(path) {
			if content, err := a.readLogFile(path); err == nil {
				*logs = append(*logs, content)
			}
		}

		return nil
	})
}

// isLogFile determines if a file should be collected as a log
func (a *artifactCollector) isLogFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	name := strings.ToLower(filepath.Base(path))

	// Include common log file extensions and names
	logExtensions := []string{".log", ".txt", ".out", ".err"}
	for _, logExt := range logExtensions {
		if ext == logExt {
			return true
		}
	}

	// Include files with log-like names
	logNames := []string{"events", "describe", "status", "pods", "nodes"}
	for _, logName := range logNames {
		if strings.Contains(name, logName) {
			return true
		}
	}

	return false
}

// readLogFile reads a log file
func (a *artifactCollector) readLogFile(path string) (LogEntry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return LogEntry{}, err
	}

	return LogEntry{
		Source:  path,
		Content: string(content),
	}, nil
}

// findJUnitFiles recursively searches for JUnit XML files
func (a *artifactCollector) findJUnitFiles(reportDir string) ([]string, error) {
	var junitFiles []string

	err := filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Look for XML files that might be JUnit reports
		if strings.HasSuffix(strings.ToLower(info.Name()), ".xml") &&
			strings.Contains(strings.ToLower(info.Name()), "junit") {
			junitFiles = append(junitFiles, path)
		}

		return nil
	})

	return junitFiles, err
}
