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
			shouldContain:     []string{"line 1", "line 100", "line 250"},
			shouldNotContain:  []string{"lines omitted"},
			expectedTruncated: false,
		},
		{
			name: "shows last 80 lines when no failures found in large file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := generateLines(500)
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			shouldContain:     []string{"No [FAILED] markers found", "line 421", "line 500"},
			shouldNotContain:  []string{"line 1", "line 20", "line 200"},
			expectedTruncated: false,
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
			shouldContain:     []string{"No [FAILED] markers found", "line 221", "line 300"},
			expectedTruncated: false,
		},
		{
			name: "extracts failure blocks from large file with [FAILED] markers",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				var content strings.Builder
				for i := 1; i <= 500; i++ {
					if i == 100 {
						content.WriteString("Running test: authentication\n")
						content.WriteString("[FAILED] authentication failed\n")
						content.WriteString("Expected: true\n")
						content.WriteString("Got: false\n")
					} else if i == 300 {
						content.WriteString("Running test: database connection\n")
						content.WriteString("• [FAILED] connection timeout\n")
						content.WriteString("Timeout after 30s\n")
					} else {
						content.WriteString(fmt.Sprintf("line %d\n", i))
					}
				}
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content.String()), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				t.Logf("Created test file with %d lines", len(strings.Split(content.String(), "\n")))
				return tmpDir
			},
			shouldContain:    []string{"Found 2 test failure(s)", "[FAILED] authentication failed", "• [FAILED] connection timeout", "---"},
			shouldNotContain: []string{"line 50", "line 450", "No [FAILED] markers found"},
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
				"# Possible Cause",
				"Database connection timeout",
				"# Recommendations",
				"1. Check network connectivity",
				"2. Verify database credentials",
			},
		},
		{
			name:  "handles JSON with only root_cause",
			input: "```json\n{\"root_cause\": \"Memory exhausted\"}\n```",
			shouldContain: []string{
				"# Possible Cause",
				"Memory exhausted",
			},
		},
		{
			name:  "handles JSON with only recommendations",
			input: "```json\n{\"recommendations\": [\"Restart service\", \"Check logs\"]}\n```",
			shouldContain: []string{
				"# Recommendations",
				"1. Restart service",
				"2. Check logs",
			},
		},
		{
			name:  "handles empty root_cause gracefully",
			input: "```json\n{\"root_cause\": \"\", \"recommendations\": [\"Action item\"]}\n```",
			shouldContain: []string{
				"# Recommendations",
				"1. Action item",
			},
		},
		{
			name:  "handles empty recommendations array",
			input: "```json\n{\"root_cause\": \"Error found\", \"recommendations\": []}\n```",
			shouldContain: []string{
				"# Possible Cause",
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
	t.Logf("First line: %q", resultLines[0])

	// With the new approach, we extract only failure blocks, so we expect "Found N test failure(s)"
	if !strings.Contains(result, "Found") || !strings.Contains(result, "test failure(s)") {
		t.Error("expected 'Found N test failure(s)' in result")
	}

	// Verify we get the important failure details
	// With the new approach, we extract only failure blocks with [FAILED] markers
	expectedFailures := []string{
		"FAIL",                                // Test failure marker
		"osd-metrics-exporter",                // One of the failing tests
		"managed-cluster-validating-webhooks", // Another failing test
	}

	for _, expected := range expectedFailures {
		if !strings.Contains(result, expected) {
			t.Errorf("expected result to contain failure detail %q", expected)
		}
	}

	// Verify we're getting reasonable content length with only failure blocks
	// We extract up to 3 failure blocks with context, so should have meaningful content
	lines := strings.Split(result, "\n")
	if len(lines) < 10 {
		t.Errorf("expected at least 10 lines in failure block output, got %d", len(lines))
	}

	// The middle repetitive content should be omitted since we only extract failure blocks
	// Lines 100-300 had repeated "Unable to find image using selector" messages
	repetitiveCount := strings.Count(result, "Unable to find image using selector")
	// We should see very few instances since we're only extracting failure blocks
	if repetitiveCount > 10 {
		t.Errorf("expected failure block extraction to reduce repetitive content, but found %d instances of repeated message", repetitiveCount)
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

func TestSlackReporter_buildClusterInfoSection(t *testing.T) {
	tests := []struct {
		name          string
		clusterInfo   *ClusterInfo
		shouldContain []string
		shouldBeEmpty bool
	}{
		{
			name: "builds full cluster info section",
			clusterInfo: &ClusterInfo{
				ID:         "cluster-123",
				Name:       "test-cluster",
				Version:    "4.20.0",
				Provider:   "aws",
				Expiration: "2026-01-30",
			},
			shouldContain: []string{
				"# Cluster Info",
				"Cluster ID: cluster-123",
				"Name: test-cluster",
				"Version: 4.20.0",
				"Provider: aws",
				"Expiration: 2026-01-30",
			},
		},
		{
			name: "handles cluster info with only ID",
			clusterInfo: &ClusterInfo{
				ID: "cluster-456",
			},
			shouldContain: []string{
				"# Cluster Info",
				"Cluster ID: cluster-456",
			},
		},
		{
			name:          "returns empty string for nil cluster info",
			clusterInfo:   nil,
			shouldBeEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			config := &ReporterConfig{
				Settings: map[string]interface{}{
					"cluster_info": tt.clusterInfo,
				},
			}

			result := reporter.buildClusterInfoSection(config)

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

func TestSlackReporter_buildTestSuiteSection(t *testing.T) {
	tests := []struct {
		name          string
		image         string
		env           string
		shouldContain []string
		shouldBeEmpty bool
	}{
		{
			name:  "builds test suite section with image and env",
			image: "quay.io/osde2e:abc123",
			env:   "stage",
			shouldContain: []string{
				"Test suite: quay.io/osde2e",
				"Commit: abc123",
				"Environment: stage",
			},
		},
		{
			name:  "builds test suite section without env",
			image: "quay.io/osde2e:def456",
			shouldContain: []string{
				"Test suite: quay.io/osde2e",
				"Commit: def456",
			},
		},
		{
			name:          "returns empty for invalid image format",
			image:         "invalid-image-no-colon",
			shouldBeEmpty: true,
		},
		{
			name:          "returns empty for empty image",
			image:         "",
			shouldBeEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			config := &ReporterConfig{
				Settings: map[string]interface{}{
					"image": tt.image,
					"env":   tt.env,
				},
			}

			result := reporter.buildTestSuiteSection(config)

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

func TestSlackReporter_buildAnalysisSection(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		shouldContain []string
	}{
		{
			name: "formats JSON analysis with root_cause and recommendations",
			content: `Analysis result:
` + "```json" + `
{
  "root_cause": "Database connection timeout",
  "recommendations": ["Check network", "Verify credentials"]
}
` + "```",
			shouldContain: []string{
				"# Possible Cause",
				"Database connection timeout",
				"# Recommendations",
				"1. Check network",
				"2. Verify credentials",
			},
		},
		{
			name:          "returns plain content when no JSON",
			content:       "This is plain analysis without JSON",
			shouldContain: []string{"Analysis:", "This is plain analysis without JSON"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			result := &AnalysisResult{
				Content: tt.content,
			}

			output := reporter.buildAnalysisSection(result)

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, but it didn't. Output: %s", expected, output)
				}
			}
		})
	}
}

func TestSlackReporter_buildTruncationNotice(t *testing.T) {
	tests := []struct {
		name             string
		omittedContent   string
		hasBotToken      bool
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:           "notice with bot token mentions attached files",
			omittedContent: "test suite info",
			hasBotToken:    true,
			shouldContain:  []string{"test suite info omitted", "see attached files"},
		},
		{
			name:             "notice without bot token mentions length limit",
			omittedContent:   "analysis",
			hasBotToken:      false,
			shouldContain:    []string{"analysis omitted", "Slack message length limit"},
			shouldNotContain: []string{"attached files"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			result := reporter.buildTruncationNotice(tt.omittedContent, tt.hasBotToken)

			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, but it didn't. Result: %s", expected, result)
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(result, notExpected) {
					t.Errorf("expected result to NOT contain %q, but it did. Result: %s", notExpected, result)
				}
			}
		})
	}
}

func TestSlackReporter_buildTruncatedMessage(t *testing.T) {
	tests := []struct {
		name             string
		header           string
		clusterInfo      string
		analysis         string
		errorMsg         string
		testSuiteInfo    string
		maxLength        int
		hasBotToken      bool
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "everything fits within limit",
			header:        "Header\n",
			clusterInfo:   "Cluster: 123\n",
			analysis:      "Analysis content\n",
			errorMsg:      "Error: test\n",
			testSuiteInfo: "Test suite: foo\n",
			maxLength:     1000,
			hasBotToken:   false,
			shouldContain: []string{"Header", "Cluster: 123", "Analysis content", "Error: test", "Test suite: foo"},
		},
		{
			name:             "drops test suite when over limit",
			header:           "Header\n",
			clusterInfo:      "Cluster: 123\n",
			analysis:         "Analysis content\n",
			errorMsg:         "Error: test\n",
			testSuiteInfo:    strings.Repeat("x", 500),
			maxLength:        200,
			hasBotToken:      true,
			shouldContain:    []string{"Header", "Cluster: 123", "Analysis content", "Error: test", "test suite info omitted", "see attached files"},
			shouldNotContain: []string{"xxxxx"},
		},
		{
			name:             "drops error message when over limit",
			header:           "Header\n",
			clusterInfo:      "Cluster: 123\n",
			analysis:         strings.Repeat("a", 50),
			errorMsg:         strings.Repeat("e", 200),
			testSuiteInfo:    "Test suite\n",
			maxLength:        200,
			hasBotToken:      false,
			shouldContain:    []string{"Header", "Cluster: 123", "error message and test suite info omitted"},
			shouldNotContain: []string{"eeeee"},
		},
		{
			name:          "truncates analysis when over limit",
			header:        "Header\n",
			clusterInfo:   "Cluster: 123\n",
			analysis:      strings.Repeat("a", 500),
			errorMsg:      "",
			testSuiteInfo: "",
			maxLength:     300,
			hasBotToken:   true,
			shouldContain: []string{"Header", "Cluster: 123", "aaa", "partial analysis", "omitted", "see attached files"},
		},
		{
			name:          "handles cluster info too large",
			header:        "H\n",
			clusterInfo:   strings.Repeat("c", 500),
			analysis:      "A\n",
			errorMsg:      "",
			testSuiteInfo: "",
			maxLength:     100,
			hasBotToken:   false,
			shouldContain: []string{"H", "WARNING: Cluster information too large"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			result := reporter.buildTruncatedMessage(
				tt.header,
				tt.clusterInfo,
				tt.analysis,
				tt.errorMsg,
				tt.testSuiteInfo,
				tt.maxLength,
				tt.hasBotToken,
			)

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

			if len(result) > tt.maxLength {
				t.Errorf("result length %d exceeds maxLength %d", len(result), tt.maxLength)
			}
		})
	}
}

func TestSlackReporter_extractFailureBlocks(t *testing.T) {
	tests := []struct {
		name          string
		lines         []string
		startIdx      int
		endIdx        int
		expectedCount int
		shouldContain []string
	}{
		{
			name: "extracts single failure block",
			lines: []string{
				"line 1",
				"line 2",
				"[FAILED] test failed",
				"error details",
				"line 5",
			},
			startIdx:      0,
			endIdx:        5,
			expectedCount: 1,
			shouldContain: []string{"[FAILED] test failed", "error details"},
		},
		{
			name: "extracts multiple failure blocks",
			lines: append(append(
				generateLinesArray(50),
				"[FAILED] first failure",
				"error 1",
			),
				append(
					generateLinesArray(50),
					"• [FAILED] second failure",
					"error 2",
				)...,
			),
			startIdx:      0,
			endIdx:        104,
			expectedCount: 2,
			shouldContain: []string{"[FAILED] first failure", "• [FAILED] second failure"},
		},
		{
			name: "limits to 3 failure blocks maximum",
			lines: func() []string {
				lines := generateLinesArray(10)
				lines = append(lines, "[FAILED] failure 1")
				lines = append(lines, generateLinesArray(30)...)
				lines = append(lines, "[FAILED] failure 2")
				lines = append(lines, generateLinesArray(30)...)
				lines = append(lines, "[FAILED] failure 3")
				lines = append(lines, generateLinesArray(30)...)
				lines = append(lines, "[FAILED] failure 4")
				lines = append(lines, generateLinesArray(10)...)
				return lines
			}(),
			startIdx:      0,
			endIdx:        125,
			expectedCount: 3,
			shouldContain: []string{"failure 1", "failure 2", "failure 3"},
		},
		{
			name:          "returns empty for no failures",
			lines:         generateLinesArray(100),
			startIdx:      0,
			endIdx:        100,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			result := reporter.extractFailureBlocks(tt.lines, tt.startIdx, tt.endIdx)

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d blocks, got %d", tt.expectedCount, len(result))
			}

			combinedResult := strings.Join(result, "\n")
			for _, expected := range tt.shouldContain {
				if !strings.Contains(combinedResult, expected) {
					t.Errorf("expected result to contain %q, but it didn't", expected)
				}
			}
		})
	}
}

func TestSlackReporter_formatMessageWithStdout(t *testing.T) {
	tests := []struct {
		name             string
		analysisResult   *AnalysisResult
		reportDir        string
		setupReportDir   func(t *testing.T) string
		hasBotToken      bool
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "includes full test failures when it fits",
			analysisResult: &AnalysisResult{
				Content: "Short analysis",
			},
			setupReportDir: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := "Found 1 test failure(s):\n\n[FAILED] test failed\nerror details\n"
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return tmpDir
			},
			hasBotToken:   false,
			shouldContain: []string{"# Test Failures", "Found 1 test failure(s)", "[FAILED] test failed"},
		},
		{
			name: "truncates test failures from beginning when too long",
			analysisResult: &AnalysisResult{
				Content: strings.Repeat("Long analysis content. ", 100),
			},
			setupReportDir: func(t *testing.T) string {
				tmpDir := t.TempDir()
				var content strings.Builder
				content.WriteString("Found 3 test failure(s):\n\n")
				content.WriteString("FIRST FAILURE BLOCK - This should appear in truncated output\n")
				content.WriteString(strings.Repeat("x", 1000))
				content.WriteString("\n\nLAST FAILURE BLOCK - This should be truncated\n")
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content.String()), 0o644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return tmpDir
			},
			hasBotToken: false,
			shouldContain: []string{
				"# Test Failures",
				"Found 3 test failure(s)",
				"FIRST FAILURE BLOCK",
				"...",
			},
			shouldNotContain: []string{"LAST FAILURE BLOCK"},
		},
		{
			name: "mentions attached files when bot token is present",
			analysisResult: &AnalysisResult{
				Content: strings.Repeat("Analysis ", 200),
			},
			setupReportDir: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := "Found 2 test failure(s):\n\n" + strings.Repeat("Failure details\n", 200)
				if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
				return tmpDir
			},
			hasBotToken:   true,
			shouldContain: []string{"see attached files for full output"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			reportDir := tt.setupReportDir(t)

			config := &ReporterConfig{
				Settings: map[string]interface{}{
					"report_dir": reportDir,
				},
			}
			if tt.hasBotToken {
				config.Settings["bot_token"] = "xoxb-test-token"
			}

			result := reporter.formatMessageWithStdout(tt.analysisResult, config)

			for _, expected := range tt.shouldContain {
				if !strings.Contains(result.Text, expected) {
					t.Errorf("expected Text to contain %q, but it didn't. Text length: %d", expected, len(result.Text))
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(result.Text, notExpected) {
					t.Errorf("expected Text to NOT contain %q, but it did", notExpected)
				}
			}
		})
	}
}

func TestBuildNotificationConfig(t *testing.T) {
	tests := []struct {
		name        string
		webhook     string
		channel     string
		clusterInfo interface{}
		reportDir   string
		botToken    string
		expectNil   bool
		checkFunc   func(t *testing.T, config *NotificationConfig)
	}{
		{
			name:      "returns nil when webhook is empty",
			webhook:   "",
			channel:   "test-channel",
			expectNil: true,
		},
		{
			name:      "returns nil when channel is empty",
			webhook:   "https://hooks.slack.com/test",
			channel:   "",
			expectNil: true,
		},
		{
			name:    "creates config with all settings",
			webhook: "https://hooks.slack.com/test",
			channel: "test-channel",
			clusterInfo: &ClusterInfo{
				ID:   "cluster-123",
				Name: "test-cluster",
			},
			reportDir: "/tmp/reports",
			botToken:  "xoxb-test",
			expectNil: false,
			checkFunc: func(t *testing.T, config *NotificationConfig) {
				if !config.Enabled {
					t.Error("expected config to be enabled")
				}
				if len(config.Reporters) != 1 {
					t.Fatalf("expected 1 reporter, got %d", len(config.Reporters))
				}
				reporter := config.Reporters[0]
				if reporter.Type != "slack" {
					t.Errorf("expected type 'slack', got %q", reporter.Type)
				}
				if !reporter.Enabled {
					t.Error("expected reporter to be enabled")
				}
				if webhook, ok := reporter.Settings["webhook_url"].(string); !ok || webhook != "https://hooks.slack.com/test" {
					t.Errorf("unexpected webhook_url: %v", reporter.Settings["webhook_url"])
				}
				if channel, ok := reporter.Settings["channel"].(string); !ok || channel != "test-channel" {
					t.Errorf("unexpected channel: %v", reporter.Settings["channel"])
				}
				if reportDir, ok := reporter.Settings["report_dir"].(string); !ok || reportDir != "/tmp/reports" {
					t.Errorf("unexpected report_dir: %v", reporter.Settings["report_dir"])
				}
				if botToken, ok := reporter.Settings["bot_token"].(string); !ok || botToken != "xoxb-test" {
					t.Errorf("unexpected bot_token: %v", reporter.Settings["bot_token"])
				}
			},
		},
		{
			name:    "creates config without bot token",
			webhook: "https://hooks.slack.com/test",
			channel: "test-channel",
			checkFunc: func(t *testing.T, config *NotificationConfig) {
				reporter := config.Reporters[0]
				if _, exists := reporter.Settings["bot_token"]; exists {
					t.Error("expected bot_token to not be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNotificationConfig(tt.webhook, tt.channel, tt.clusterInfo, tt.reportDir, tt.botToken)

			if tt.expectNil {
				if result != nil {
					t.Errorf("expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// Helper function to generate array of lines
func generateLinesArray(n int) []string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	return lines
}
