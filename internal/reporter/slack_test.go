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
			name: "returns full content for file with exactly 150 lines",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := generateLines(150)
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			// With 150 lines + trailing newline = 151 elements after split, which is > 150, so it gets truncated
			shouldContain:     []string{"line 1", "line 50", "line 52", "line 150", "lines omitted"},
			expectedTruncated: true,
		},
		{
			name: "truncates large file with first 50 and last 100 lines",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := generateLines(500)
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			shouldContain:     []string{"line 1", "line 50", "line 402", "line 500", "(351 lines omitted)"},
			shouldNotContain:  []string{"line 51", "line 401"},
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
			name: "handles file with 151 lines correctly",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := generateLines(151)
				err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return tmpDir
			},
			// 151 lines + trailing newline = 152 elements; first 50 + last 100 = 150, omit 2
			shouldContain:     []string{"line 1", "line 50", "line 53", "line 151", "(2 lines omitted)"},
			shouldNotContain:  []string{"line 51", "line 52"},
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
		expectedCount int
		expectedExts  []string
	}{
		{
			name: "collects all log, txt, and xml files",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				files := []string{"test.log", "output.txt", "results.xml", "ignore.json", "data.csv"}
				for _, file := range files {
					if err := os.WriteFile(filepath.Join(tmpDir, file), []byte("content"), 0o644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return tmpDir
			},
			expectedCount: 3,
			expectedExts:  []string{".log", ".txt", ".xml"},
		},
		{
			name: "collects files from subdirectories",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				subDir := filepath.Join(tmpDir, "logs")
				if err := os.MkdirAll(subDir, 0o755); err != nil {
					t.Fatalf("failed to create subdir: %v", err)
				}
				files := map[string]string{
					"test.log":         "content",
					"logs/nested.log":  "content",
					"logs/results.xml": "content",
				}
				for file, content := range files {
					if err := os.WriteFile(filepath.Join(tmpDir, file), []byte(content), 0o644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return tmpDir
			},
			expectedCount: 3,
			expectedExts:  []string{".log", ".xml"},
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
		{
			name: "handles case-insensitive extensions",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				files := []string{"test.LOG", "output.TXT", "results.XML"}
				for _, file := range files {
					if err := os.WriteFile(filepath.Join(tmpDir, file), []byte("content"), 0o644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return tmpDir
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &SlackReporter{}
			reportDir := tt.setupFunc(t)
			result := reporter.collectLogFiles(reportDir)

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d files, got %d", tt.expectedCount, len(result))
			}

			if tt.expectedExts != nil {
				for _, file := range result {
					ext := strings.ToLower(filepath.Ext(file))
					found := false
					for _, expectedExt := range tt.expectedExts {
						if ext == expectedExt {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("unexpected file extension %q in result", ext)
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

// generateLines creates N lines of test content
func generateLines(n int) string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	return strings.Join(lines, "\n") + "\n"
}
