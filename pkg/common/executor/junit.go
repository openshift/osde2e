package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/joshdk/go-junit"
)

// testResults contains the summary of test execution results
type testResults struct {
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	ErrorTests   int
	Duration     time.Duration
	Suites       []junit.Suite
}

// processJUnitResults finds and processes all JUnit XML files in the given output directory
func processJUnitResults(logger logr.Logger, outputDir string) (*testResults, error) {
	// Find all JUnit XML files in the output directory
	junitFiles, err := findJUnitFiles(outputDir)
	if err != nil {
		return nil, fmt.Errorf("finding junit files: %w", err)
	}

	if len(junitFiles) == 0 {
		logger.Info("no junit files found in output directory")
		return &testResults{}, nil
	}

	logger.Info("found junit files", "count", len(junitFiles))

	// Parse all JUnit files
	var allSuites []junit.Suite
	for _, file := range junitFiles {
		suites, err := junit.IngestFile(file)
		if err != nil {
			logger.Error(err, "failed to parse junit file", "file", file)
			continue
		}
		allSuites = append(allSuites, suites...)
	}

	// Calculate test statistics
	results := &testResults{
		Suites: allSuites,
	}

	for _, suite := range allSuites {
		for _, test := range suite.Tests {
			results.TotalTests++
			switch test.Status {
			case junit.StatusPassed:
				results.PassedTests++
			case junit.StatusFailed:
				results.FailedTests++
			case junit.StatusSkipped:
				results.SkippedTests++
			case junit.StatusError:
				results.ErrorTests++
			}
		}
	}

	logger.Info("processed test results",
		"total", results.TotalTests,
		"passed", results.PassedTests,
		"failed", results.FailedTests,
		"skipped", results.SkippedTests,
		"errors", results.ErrorTests)

	return results, nil
}

// findJUnitFiles recursively searches for XML files in the given directory
func findJUnitFiles(outputDir string) ([]string, error) {
	var junitFiles []string

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Look for XML files that might be JUnit reports
		if strings.HasSuffix(strings.ToLower(info.Name()), ".xml") {
			// Additional filtering could be added here to be more specific about JUnit files
			// For now, we'll process all XML files and let the parser handle invalid ones
			junitFiles = append(junitFiles, path)
		}

		return nil
	})

	return junitFiles, err
}
