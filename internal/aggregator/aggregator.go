package aggregator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/joshdk/go-junit"
	"github.com/openshift/osde2e/internal/sanitizer"
	"github.com/openshift/osde2e/pkg/common/util"
)

type Aggregator struct {
	logger    logr.Logger
	sanitizer *sanitizer.Sanitizer // Optional data sanitizer
}

// NewWithSanitizer creates an aggregator with data sanitization capability
func NewWithSanitizer(ctx context.Context, sanitizerConfig *sanitizer.Config) (*Aggregator, error) {
	dataSanitizer, err := sanitizer.New(sanitizerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create sanitizer: %w", err)
	}

	return &Aggregator{
		logger:    logr.FromContextOrDiscard(ctx),
		sanitizer: dataSanitizer,
	}, nil
}

type AggregatedData struct {
	TestResults  TestResultSummary `json:"testResults"`
	FailedTests  []FailedTest      `json:"failedTests"`
	LogArtifacts []LogEntry        `json:"logArtifacts"`
	AnamolyLogs  string            `json:"anamolyLogs"`
}

type TestResultSummary struct {
	TotalTests   int           `json:"totalTests"`
	PassedTests  int           `json:"passedTests"`
	FailedTests  int           `json:"failedTests"`
	SkippedTests int           `json:"skippedTests"`
	ErrorTests   int           `json:"errorTests"`
	Duration     time.Duration `json:"duration"`
	SuiteCount   int           `json:"suiteCount"`
}

type FailedTest struct {
	Name      string `json:"name"`
	ClassName string `json:"className,omitempty"`
	SuiteName string `json:"suiteName,omitempty"`
}

type LogEntry struct {
	Source    string `json:"source"`
	LineCount int    `json:"lineCount"`
}

func New(ctx context.Context) *Aggregator {
	return &Aggregator{
		logger: logr.FromContextOrDiscard(ctx),
	}
}

func (a *Aggregator) Collect(ctx context.Context, reportDir string) (*AggregatedData, error) {
	a.logger.Info("collecting artifacts", "reportDir", reportDir)

	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("report directory does not exist: %s", reportDir)
	}

	data := &AggregatedData{}

	var collectionErrors []string

	if err := a.collectLogArtifacts(reportDir, data); err != nil {
		errMsg := fmt.Sprintf("failed to collect log artifacts: %v", err)
		a.logger.Error(err, "failed to collect log artifacts")
		collectionErrors = append(collectionErrors, errMsg)
	}

	if err := a.collectLogAnomalies(reportDir, data); err != nil {
		errMsg := fmt.Sprintf("failed to collect log anomalies: %v", err)
		a.logger.Error(err, "failed to collect log anomaly")
		collectionErrors = append(collectionErrors, errMsg)
	}

	if err := a.collectTestResults(data); err != nil {
		errMsg := fmt.Sprintf("failed to collect test results: %v", err)
		a.logger.Error(err, "failed to collect test results")
		collectionErrors = append(collectionErrors, errMsg)
	}

	a.logger.Info("completed artifact collection",
		"failedTests", len(data.FailedTests),
		"logEntries", len(data.LogArtifacts),
		"errors", len(collectionErrors))

	return data, nil
}

func (a *Aggregator) collectLogAnomalies(reportDir string, data *AggregatedData) error {
	// TODO: Get file name in a generic way
	logFilePath := filepath.Join(reportDir, "test_output.log")
	errors, err := extractErrorsFromLogFile(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to collect log anomaly: %w", err)
	}

	// If data sanitization is enabled, sanitize anomaly logs
	if a.sanitizer != nil && errors != "" {
		result, err := a.sanitizer.SanitizeText(errors, logFilePath)
		if err != nil {
			a.logger.Error(err, "failed to sanitize anomaly logs", "source", logFilePath)
			// Use original content on sanitization failure, but log warning
		} else {
			errors = result.Content
			if result.MatchesFound > 0 {
				a.logger.Info("sanitized anomaly logs",
					"source", logFilePath,
					"matches_found", result.MatchesFound,
					"rules_applied", result.RulesApplied)
			}
		}
	}

	data.AnamolyLogs = errors
	return nil
}

func (a *Aggregator) collectTestResults(data *AggregatedData) error {
	junitFiles, err := a.findJUnitFiles(data)
	if err != nil {
		return fmt.Errorf("finding junit files: %w", err)
	}

	if len(junitFiles) == 0 {
		a.logger.Info("no junit files found")
		return nil
	}

	type junitResult struct {
		suites []junit.Suite
		err    error
		file   string
	}

	resultCh := make(chan junitResult, len(junitFiles))
	var wg sync.WaitGroup

	for _, file := range junitFiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			suites, err := junit.IngestFile(f)
			resultCh <- junitResult{suites: suites, err: err, file: f}
		}(file)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var allSuites []junit.Suite
	for result := range resultCh {
		if result.err != nil {
			a.logger.Error(result.err, "failed to parse junit file", "file", result.file)
			continue
		}
		allSuites = append(allSuites, result.suites...)
	}

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

	sort.Slice(failedTests, func(i, j int) bool {
		return failedTests[i].Name < failedTests[j].Name
	})

	data.TestResults = summary
	data.FailedTests = failedTests

	return nil
}

func (a *Aggregator) convertJUnitTest(test junit.Test, suiteName string) FailedTest {
	return FailedTest{
		Name:      test.Name,
		ClassName: test.Classname,
		SuiteName: suiteName,
	}
}

func (a *Aggregator) collectLogArtifacts(reportDir string, data *AggregatedData) error {
	return filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			a.logger.Info("error accessing path", "path", path, "error", err)
			return nil
		}

		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		lineCount := 0
		if content, err := os.ReadFile(path); err == nil {
			lineCount = strings.Count(string(content), "\n")
			if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
				lineCount++
			}
		} else {
			a.logger.Info("unable to read file for line count", "path", path, "error", err)
		}

		data.LogArtifacts = append(data.LogArtifacts, LogEntry{
			Source:    path,
			LineCount: lineCount,
		})

		return nil
	})
}

func (a *Aggregator) findJUnitFiles(data *AggregatedData) ([]string, error) {
	var junitFiles []string

	for _, logEntry := range data.LogArtifacts {
		fileName := strings.ToLower(filepath.Base(logEntry.Source))

		if strings.HasSuffix(fileName, ".xml") &&
			strings.Contains(fileName, "junit") {
			junitFiles = append(junitFiles, logEntry.Source)
		}
	}

	return junitFiles, nil
}

func extractErrorsFromLogFile(logFile string) (string, error) {
	content, err := os.ReadFile(logFile)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")

	// use string builder to collect errors
	var errors strings.Builder
	for _, line := range lines {
		if util.ContainsErrorMarker(line) {
			errors.WriteString(line)
			errors.WriteString("\n") // Add newline separator
		}
	}
	return errors.String(), nil
}
