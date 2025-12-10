package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlackReporter_readTestOutput(t *testing.T) {
	tests := []struct {
		name              string
		setupFunc         func(t *testing.T) string
		shouldContain     []string
		shouldNotContain  []string
		expectedEmpty     bool
		expectedTruncated bool
	}{
		{
			name: "returns empty string when directory does not exist",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/directory"
			},
			expectedEmpty: true,
		},
		{
			name: "returns empty string when no test output files exist",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			expectedEmpty: true,
		},
		{
			name: "returns full content for small file (100 lines)",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := generateLines(100)
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			shouldContain:     []string{"line 1", "line 50", "line 100"},
			shouldNotContain:  []string{"lines omitted"},
			expectedTruncated: false,
		},
		{
			name: "returns full content for file with exactly 250 lines",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := generateLines(250)
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			// With 250 lines, threshold is 250, so it returns full content
			shouldContain:     []string{"line 1", "line 100", "line 250"},
			shouldNotContain:  []string{"lines omitted"},
			expectedTruncated: false,
		},
		{
			name: "truncates large file with smart extraction",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := generateLines(500)
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			// 500 lines: first 20 + last 80 with smart extraction
			shouldContain:     []string{"line 1", "line 20", "line 421", "line 500", "lines omitted"},
			shouldNotContain:  []string{"line 200"},
			expectedTruncated: true,
		},
		{
			name: "prefers test_output.txt over test_output.log",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				txtContent := "content from txt file\n"
				logContent := "content from log file\n"
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(txtContent), 0o644); err != nil {
					t.Fatalf("failed to create txt file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.log"), []byte(logContent), 0o644); err != nil {
					t.Fatalf("failed to create log file: %v", err)
				}
				return tmpDir
			},
			shouldContain:    []string{"content from txt file"},
			shouldNotContain: []string{"content from log file"},
		},
		{
			name: "falls back to test_output.log when test_output.txt does not exist",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				logContent := "content from log file\n"
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.log"), []byte(logContent), 0o644); err != nil {
					t.Fatalf("failed to create log file: %v", err)
				}
				return tmpDir
			},
			shouldContain: []string{"content from log file"},
		},
		{
			name: "handles file with 300 lines correctly",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := generateLines(300)
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			// 300 lines: first 20 + last 80 = 100 shown with smart extraction
			shouldContain:     []string{"line 1", "line 20", "line 221", "line 300", "lines omitted"},
			expectedTruncated: true,
		},
		{
			name: "handles empty file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(""), 0o644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			expectedEmpty: true,
		},
		{
			name: "handles file with single line",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := "single line\n"
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			shouldContain: []string{"single line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			reportDir := tt.setupFunc(t)
			result := reporter.readTestOutput(reportDir)

			if tt.expectedEmpty {
				if result != "" {
					t.Errorf("expected empty string, got: %q", result)
				}
				return
			}

			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, but it didn't", expected)
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(result, notExpected) {
					t.Errorf("expected result to NOT contain %q, but it did", notExpected)
				}
			}

			if tt.expectedTruncated && !strings.Contains(result, "lines omitted") {
				t.Errorf("expected result to be truncated with omission notice")
			}
		})
	}
}

func TestSlackReporter_Name(t *testing.T) {
	reporter := &SlackReporter{}
	if got := reporter.Name(); got != "slack" {
		t.Errorf("expected name to be 'slack', got %q", got)
	}
}

func TestSlackReporter_formatAnalysisContent(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldContain []string
		shouldBeEmpty bool
	}{
		{
			name:          "returns empty string for content without JSON block",
			input:         "This is just plain text without JSON",
			shouldBeEmpty: true,
		},
		{
			name:          "returns empty string for invalid JSON",
			input:         "```json\n{invalid json content\n```",
			shouldBeEmpty: true,
		},
		{
			name:  "formats valid JSON with root_cause and recommendations",
			input: "Analysis result:\n```json\n{\n  \"root_cause\": \"Database connection timeout\",\n  \"recommendations\": [\"Check network connectivity\", \"Verify database credentials\"]\n}\n```",
			shouldContain: []string{
				"====== 🔍 Possible Cause ======",
				"Database connection timeout",
				"====== 💡 Recommendations ======",
				"1. Check network connectivity",
				"2. Verify database credentials",
			},
		},
		{
			name:  "handles JSON with only root_cause",
			input: "```json\n{\"root_cause\": \"Memory exhausted\"}\n```",
			shouldContain: []string{
				"====== 🔍 Possible Cause ======",
				"Memory exhausted",
			},
		},
		{
			name:  "handles JSON with only recommendations",
			input: "```json\n{\"recommendations\": [\"Restart service\", \"Check logs\"]}\n```",
			shouldContain: []string{
				"====== 💡 Recommendations ======",
				"1. Restart service",
				"2. Check logs",
			},
		},
		{
			name:  "handles empty root_cause gracefully",
			input: "```json\n{\"root_cause\": \"\", \"recommendations\": [\"Action item\"]}\n```",
			shouldContain: []string{
				"====== 💡 Recommendations ======",
				"1. Action item",
			},
		},
		{
			name:  "handles empty recommendations array",
			input: "```json\n{\"root_cause\": \"Error found\", \"recommendations\": []}\n```",
			shouldContain: []string{
				"====== 🔍 Possible Cause ======",
				"Error found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			result := reporter.formatAnalysisContent(tt.input)

			if tt.shouldBeEmpty {
				if result != "" {
					t.Errorf("expected empty result, got: %q", result)
				}
				return
			}

			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, but it didn't. Result: %s", expected, result)
				}
			}
		})
	}
}

func TestSlackReporter_collectLogFiles(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string
		config        *ReporterConfig
		expectedCount int
		expectedFiles []string
	}{
		{
			name: "collects default pattern files (test_output.log, test_output.txt, junit*.xml)",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				files := []string{"test_output.log", "test_output.txt", "junit_abc.xml", "other.log", "ignore.json"}
				for _, file := range files {
					if err := os.WriteFile(filepath.Join(tmpDir, file), []byte("content"), 0o644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return tmpDir
			},
			expectedCount: 3,
			expectedFiles: []string{"test_output.log", "test_output.txt", "junit_abc.xml"},
		},
		{
			name: "respects custom file patterns from config",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				files := []string{"test_output.log", "custom.log", "debug.log"}
				for _, file := range files {
					if err := os.WriteFile(filepath.Join(tmpDir, file), []byte("content"), 0o644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return tmpDir
			},
			config: &ReporterConfig{
				Settings: map[string]interface{}{
					"log_file_patterns": []string{"custom.log", "debug.log"},
				},
			},
			expectedCount: 2,
			expectedFiles: []string{"custom.log", "debug.log"},
		},
		{
			name: "respects max file count limit",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create more files than the limit
				for i := 0; i < 20; i++ {
					filename := fmt.Sprintf("junit_%d.xml", i)
					if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("content"), 0o644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return tmpDir
			},
			config: &ReporterConfig{
				Settings: map[string]interface{}{
					"max_log_files": 5,
				},
			},
			expectedCount: 5,
		},
		{
			name: "respects max size limit",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create files that exceed size limit (2MB each, limit 5MB = 2 files max)
				largeContent := strings.Repeat("x", 2*1024*1024) // 2MB
				for i := 0; i < 5; i++ {
					filename := fmt.Sprintf("junit_%d.xml", i)
					if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(largeContent), 0o644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return tmpDir
			},
			config: &ReporterConfig{
				Settings: map[string]interface{}{
					"max_log_size_mb": 5,
				},
			},
			expectedCount: 2, // Should stop at 2 files (4MB total) before hitting 5MB limit
		},
		{
			name: "collects files from subdirectories",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				subDir := filepath.Join(tmpDir, "install")
				if err := os.MkdirAll(subDir, 0o755); err != nil {
					t.Fatalf("failed to create subdir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.log"), []byte("content"), 0o644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				if err := os.WriteFile(filepath.Join(subDir, "junit_hz6c3.xml"), []byte("content"), 0o644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return tmpDir
			},
			expectedCount: 2,
			expectedFiles: []string{"test_output.log", "junit_hz6c3.xml"},
		},
		{
			name: "returns empty slice for directory with no matching files",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "test.json"), []byte("{}"), 0o644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return tmpDir
			},
			expectedCount: 0,
		},
		{
			name: "returns empty slice for empty directory",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			reportDir := tt.setupFunc(t)

			var result []string
			if tt.config != nil {
				result = reporter.collectLogFilesWithConfig(reportDir, tt.config)
			} else {
				result = reporter.collectLogFiles(reportDir)
			}

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d files, got %d. Files: %v", tt.expectedCount, len(result), result)
			}

			if tt.expectedFiles != nil {
				for _, expectedFile := range tt.expectedFiles {
					found := false
					for _, file := range result {
						if strings.HasSuffix(file, expectedFile) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected file %q not found in result: %v", expectedFile, result)
					}
				}
			}
		})
	}
}

func TestSlackReporterConfig(t *testing.T) {
	tests := []struct {
		name       string
		webhookURL string
		enabled    bool
	}{
		{
			name:       "creates config with enabled reporter",
			webhookURL: "https://hooks.slack.com/services/TEST/WEBHOOK/URL",
			enabled:    true,
		},
		{
			name:       "creates config with disabled reporter",
			webhookURL: "https://hooks.slack.com/services/TEST/WEBHOOK/URL",
			enabled:    false,
		},
		{
			name:       "handles empty webhook URL",
			webhookURL: "",
			enabled:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SlackReporterConfig(tt.webhookURL, tt.enabled)

			if config.Type != "slack" {
				t.Errorf("expected type 'slack', got %q", config.Type)
			}

			if config.Enabled != tt.enabled {
				t.Errorf("expected enabled to be %v, got %v", tt.enabled, config.Enabled)
			}

			webhookURL, ok := config.Settings["webhook_url"].(string)
			if !ok {
				t.Fatal("webhook_url setting is not a string")
			}

			if webhookURL != tt.webhookURL {
				t.Errorf("expected webhook_url to be %q, got %q", tt.webhookURL, webhookURL)
			}
		})
	}
}

// TestSlackReporter_GenerateTruncatedFixture generates a fixture showing truncated output
// Run with: go test -v ./internal/reporter -run TestSlackReporter_GenerateTruncatedFixture
func TestSlackReporter_GenerateTruncatedFixture(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping fixture generation in short mode")
	}

	reporter := &SlackReporter{}

	// Point to the real Prow build log
	testDataDir := "testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws"
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(testDataDir, "build-log.txt")
	destFile := filepath.Join(tmpDir, "test_output.txt")

	content, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Skipf("Skipping: real test data not available: %v", err)
		return
	}

	if err := os.WriteFile(destFile, content, 0o644); err != nil {
		t.Fatalf("failed to setup test file: %v", err)
	}

	// Generate truncated output
	truncated := reporter.readTestOutput(tmpDir)

	// Write to fixture file
	fixtureFile := filepath.Join(testDataDir, "build-log-truncated.txt")
	if err := os.WriteFile(fixtureFile, []byte(truncated), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	// Report statistics
	originalLines := len(strings.Split(strings.TrimRight(string(content), "\n"), "\n"))
	truncatedLines := len(strings.Split(strings.TrimRight(truncated, "\n"), "\n"))

	t.Logf("✓ Generated fixture: %s", fixtureFile)
	t.Logf("  Original: %d lines", originalLines)
	t.Logf("  Truncated: %d lines", truncatedLines)
	t.Logf("  Reduction: %.1f%%", float64(originalLines-truncatedLines)/float64(originalLines)*100)
}

// TestSlackReporter_readTestOutput_Debug helps debug truncation logic
func TestSlackReporter_readTestOutput_Debug(t *testing.T) {
	reporter := &SlackReporter{}

	// Test 250 lines to understand the truncation
	tmpDir := t.TempDir()
	content := generateLines(250)
	if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result := reporter.readTestOutput(tmpDir)
	resultLines := strings.Split(result, "\n")

	t.Logf("Total lines in result: %d", len(resultLines))
	t.Logf("First 5 lines: %v", resultLines[0:5])
	t.Logf("Contains 'line 20': %v", strings.Contains(result, "line 20"))
	t.Logf("Contains 'line 21': %v", strings.Contains(result, "line 21"))
	t.Logf("Contains 'line 70': %v", strings.Contains(result, "line 70"))
	t.Logf("Contains 'line 71': %v", strings.Contains(result, "line 71"))
	t.Logf("Contains omission: %v", strings.Contains(result, "lines omitted"))
}

// TestSlackReporter_readTestOutput_RealProwData tests with actual Prow job artifacts
func TestSlackReporter_readTestOutput_RealProwData(t *testing.T) {
	reporter := &SlackReporter{}

	// Point to the real Prow build log we downloaded
	testDataDir := "testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws"

	// Create a copy as test_output.txt for the test
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(testDataDir, "build-log.txt")
	destFile := filepath.Join(tmpDir, "test_output.txt")

	content, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Skipf("Skipping test: real test data not available: %v", err)
		return
	}

	if err := os.WriteFile(destFile, content, 0o644); err != nil {
		t.Fatalf("failed to setup test file: %v", err)
	}

	result := reporter.readTestOutput(tmpDir)

	// Verify the result is not empty
	if result == "" {
		t.Fatal("expected non-empty result from real Prow data")
	}

	resultLines := strings.Split(result, "\n")
	t.Logf("Real Prow data - Total lines in result: %d", len(resultLines))
	t.Logf("First 3 lines: %v", resultLines[0:3])
	t.Logf("Contains 'Loading config': %v", strings.Contains(result, "Loading config"))
	t.Logf("Contains 'Will load config': %v", strings.Contains(result, "Will load config"))
	t.Logf("First line: %q", resultLines[0])

	// The build log is 646 lines, so it should be truncated
	if !strings.Contains(result, "lines omitted") {
		t.Error("expected truncation notice in result")
	}

	// Verify we get the initial context (first ~20 lines)
	expectedInitial := []string{
		"Will load config", // From initial lines
		"aws",              // Config name
		"stage",            // Environment
		"e2e-suite",        // Test suite
	}

	for _, expected := range expectedInitial {
		if !strings.Contains(result, expected) {
			t.Errorf("expected result to contain initial context %q", expected)
		}
	}

	// Verify we get the important failure details (from the end)
	expectedFailures := []string{
		"FAIL",                                     // Test failure marker
		"osd-metrics-exporter",                     // One of the failing tests
		"managed-cluster-validating-webhooks",      // Another failing test
		"Summarizing 2 Failures",                   // Summary section
		"Tests failed: tests failed",               // Final result
		"Cluster 2ntr2hoo8487ite28bd98pg5ph0m04gf", // Cluster ID at the end
	}

	for _, expected := range expectedFailures {
		if !strings.Contains(result, expected) {
			t.Errorf("expected result to contain failure detail %q", expected)
		}
	}

	// Verify we're getting meaningful content length
	// With 20 first + 180 last lines, plus omission notice, should be substantial
	lines := strings.Split(result, "\n")
	if len(lines) < 150 {
		t.Errorf("expected at least 150 lines in truncated output, got %d", len(lines))
	}

	// The middle repetitive content should be mostly omitted
	// Lines 100-300 had repeated "Unable to find image using selector" messages
	repetitiveCount := strings.Count(result, "Unable to find image using selector")
	// We should see some from the first 20 lines, but not all ~100+ instances
	if repetitiveCount > 30 {
		t.Errorf("expected truncation to reduce repetitive content, but found %d instances of repeated message", repetitiveCount)
	}
}

// generateLines creates N lines of test content
func generateLines(n int) string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	return strings.Join(lines, "\n") + "\n"
}
