package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlackReporter_readTestOutput(t *testing.T) {
	reporter := NewSlackReporter()

	t.Run("extracts failure blocks from real Prow data", func(t *testing.T) {
		reportDir := "testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws"
		result := reporter.readTestOutput(reportDir)

		if result == "" {
			t.Fatal("expected non-empty result from real test data")
		}

		// Should contain failure count
		if !strings.Contains(result, "Found") && !strings.Contains(result, "test failure") {
			t.Error("result should indicate test failures found")
		}

		// Should contain [FAILED] markers or indicate no failures
		if !strings.Contains(result, "[FAILED]") && !strings.Contains(result, "No [FAILED] markers found") {
			t.Error("result should contain failure markers or indicate none found")
		}

		t.Logf("Extracted test output (%d chars):\n%s", len(result), result[:min(500, len(result))])
	})

	t.Run("returns empty for non-existent directory", func(t *testing.T) {
		result := reporter.readTestOutput("/nonexistent/directory")
		if result != "" {
			t.Errorf("expected empty string for non-existent directory, got: %s", result)
		}
	})

	t.Run("handles small test output", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := "line 1\nline 2\nline 3\n"
		if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		result := reporter.readTestOutput(tmpDir)
		if result != content {
			t.Errorf("expected full content for small file, got: %s", result)
		}
	})

	t.Run("extracts failure blocks from synthetic data", func(t *testing.T) {
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
				content.WriteString("line " + string(rune('0'+i%10)) + "\n")
			}
		}

		if err := os.WriteFile(filepath.Join(tmpDir, "test_output.txt"), []byte(content.String()), 0o644); err != nil {
			t.Fatal(err)
		}

		result := reporter.readTestOutput(tmpDir)

		// Should extract both failures
		if !strings.Contains(result, "[FAILED] authentication failed") {
			t.Error("should contain first failure")
		}
		if !strings.Contains(result, "• [FAILED] connection timeout") {
			t.Error("should contain second failure")
		}
		if !strings.Contains(result, "Found 2 test failure(s)") {
			t.Error("should indicate 2 failures found")
		}
	})
}

func TestSlackReporter_extractFailureBlocks(t *testing.T) {
	reporter := NewSlackReporter()

	t.Run("extracts single failure", func(t *testing.T) {
		lines := []string{
			"line 1",
			"line 2",
			"[FAILED] test failed",
			"error details",
			"line 5",
		}

		blocks := reporter.extractFailureBlocks(lines, 0, len(lines))

		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}

		if !strings.Contains(blocks[0], "[FAILED] test failed") {
			t.Error("block should contain failure marker")
		}
		if !strings.Contains(blocks[0], "error details") {
			t.Error("block should contain context after failure")
		}
	})

	t.Run("extracts multiple failures", func(t *testing.T) {
		lines := []string{
			"start",
			"[FAILED] test 1",
			"error 1",
			"padding 1", "padding 2", "padding 3", "padding 4", "padding 5",
			"padding 6", "padding 7", "padding 8", "padding 9", "padding 10",
			"padding 11", "padding 12", "padding 13", "padding 14", "padding 15",
			"padding 16", "padding 17", "padding 18", "padding 19", "padding 20",
			"padding 21", "padding 22", "padding 23", "padding 24", "padding 25",
			"padding 26", "padding 27", "padding 28", "padding 29", "padding 30",
			"padding 31", "padding 32", "padding 33", "padding 34", "padding 35",
			"[FAILED] test 2",
			"error 2",
			"end",
		}

		blocks := reporter.extractFailureBlocks(lines, 0, len(lines))

		if len(blocks) != 2 {
			t.Fatalf("expected 2 blocks, got %d", len(blocks))
		}

		if !strings.Contains(blocks[0], "[FAILED] test 1") {
			t.Error("first block should contain first failure")
		}
		if !strings.Contains(blocks[1], "[FAILED] test 2") {
			t.Error("second block should contain second failure")
		}
	})

	t.Run("limits to max failures", func(t *testing.T) {
		lines := make([]string, 0)
		for i := 0; i < 10; i++ {
			lines = append(lines, "line before")
			lines = append(lines, "[FAILED] test "+string(rune('0'+i)))
			lines = append(lines, "line after")
		}

		blocks := reporter.extractFailureBlocks(lines, 0, len(lines))

		if len(blocks) > maxFailureBlocks {
			t.Errorf("expected max %d blocks, got %d", maxFailureBlocks, len(blocks))
		}
	})

	t.Run("handles no failures", func(t *testing.T) {
		lines := []string{"line 1", "line 2", "line 3"}
		blocks := reporter.extractFailureBlocks(lines, 0, len(lines))

		if len(blocks) != 0 {
			t.Errorf("expected 0 blocks for no failures, got %d", len(blocks))
		}
	})
}

func TestSlackReporter_buildClusterInfoSection(t *testing.T) {
	reporter := NewSlackReporter()

	t.Run("builds complete cluster info", func(t *testing.T) {
		clusterInfo := &ClusterInfo{
			ID:         "cluster-abc",
			Name:       "production-cluster",
			Version:    "4.23",
			Provider:   "aws",
			Expiration: "2026-03-01T00:00:00Z",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"cluster_info": clusterInfo,
			},
		}

		result := reporter.buildClusterInfoSection(config)

		expectedFields := []string{
			"====== ☸️ Cluster Information ======",
			"cluster-abc",
			"production-cluster",
			"4.23",
			"aws",
			"2026-03-01T00:00:00Z",
		}

		for _, field := range expectedFields {
			if !strings.Contains(result, field) {
				t.Errorf("cluster info should contain %q", field)
			}
		}
	})

	t.Run("handles minimal cluster info", func(t *testing.T) {
		clusterInfo := &ClusterInfo{
			ID: "cluster-xyz",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"cluster_info": clusterInfo,
			},
		}

		result := reporter.buildClusterInfoSection(config)

		if !strings.Contains(result, "cluster-xyz") {
			t.Error("should contain cluster ID")
		}
	})

	t.Run("returns empty for nil cluster info", func(t *testing.T) {
		config := &ReporterConfig{
			Settings: map[string]interface{}{},
		}

		result := reporter.buildClusterInfoSection(config)

		if result != "" {
			t.Errorf("expected empty string for nil cluster info, got: %s", result)
		}
	})
}

func TestSlackReporter_buildTestSuiteSection(t *testing.T) {
	reporter := NewSlackReporter()

	t.Run("builds test suite info", func(t *testing.T) {
		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"image": "quay.io/openshift/test:v2.0",
				"env":   "staging",
			},
		}

		result := reporter.buildTestSuiteSection(config)

		if !strings.Contains(result, "quay.io/openshift/test") {
			t.Error("should contain image name")
		}
		if !strings.Contains(result, "v2.0") {
			t.Error("should contain commit/tag")
		}
		if !strings.Contains(result, "staging") {
			t.Error("should contain environment")
		}
	})

	t.Run("returns empty for missing image", func(t *testing.T) {
		config := &ReporterConfig{
			Settings: map[string]interface{}{},
		}

		result := reporter.buildTestSuiteSection(config)

		if result != "" {
			t.Errorf("expected empty string for missing image, got: %s", result)
		}
	})

	t.Run("returns empty for invalid image format", func(t *testing.T) {
		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"image": "invalid-no-tag",
			},
		}

		result := reporter.buildTestSuiteSection(config)

		if result != "" {
			t.Errorf("expected empty string for invalid image format, got: %s", result)
		}
	})
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
