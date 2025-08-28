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
)

type Aggregator struct {
	logger logr.Logger
}

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
	Name       string        `json:"name"`
	ClassName  string        `json:"className,omitempty"`
	SuiteName  string        `json:"suiteName,omitempty"`
	Duration   time.Duration `json:"duration"`
	ErrorMsg   string        `json:"errorMessage,omitempty"`
	StackTrace string        `json:"stackTrace,omitempty"`
	SystemOut  string        `json:"systemOut,omitempty"`
	SystemErr  string        `json:"systemErr,omitempty"`
}

type LogEntry struct {
	Source string `json:"source"`
}

func New(logger logr.Logger) *Aggregator {
	return &Aggregator{
		logger: logger,
	}
}

func (a *Aggregator) Collect(ctx context.Context, reportDir string) (*AggregatedData, error) {
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

func (a *Aggregator) collectLogs(reportDir string, data *AggregatedData) error {
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

		data.Logs = append(data.Logs, LogEntry{
			Source: path,
		})

		return nil
	})
}

func (a *Aggregator) findJUnitFiles(data *AggregatedData) ([]string, error) {
	var junitFiles []string

	for _, logEntry := range data.Logs {
		fileName := strings.ToLower(filepath.Base(logEntry.Source))

		if strings.HasSuffix(fileName, ".xml") &&
			strings.Contains(fileName, "junit") {
			junitFiles = append(junitFiles, logEntry.Source)
		}
	}

	return junitFiles, nil
}
