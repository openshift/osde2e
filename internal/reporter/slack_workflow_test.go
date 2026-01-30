package reporter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSlackReporter_buildWorkflowPayload(t *testing.T) {
	reporter := NewSlackReporter()

	result := &AnalysisResult{
		Content: `Here is the analysis

` + "```json" + `
{
  "root_cause": "Test failed due to timeout",
  "recommendations": ["Increase timeout", "Check network"]
}
` + "```" + `
`,
	}

	clusterInfo := &ClusterInfo{
		ID:         "test-123",
		Name:       "test-cluster",
		Version:    "4.20",
		Provider:   "aws",
		Expiration: "2026-01-28T10:00:00Z",
	}

	config := &ReporterConfig{
		Settings: map[string]interface{}{
			"webhook_url":  "https://hooks.slack.com/test",
			"channel":      "C06HQR8HN0L",
			"cluster_info": clusterInfo,
			"image":        "quay.io/test:abc123",
			"env":          "stage",
		},
	}

	payload := reporter.buildWorkflowPayload(result, config)

	// Verify required fields
	if payload.Channel != "C06HQR8HN0L" {
		t.Errorf("expected channel C06HQR8HN0L, got %s", payload.Channel)
	}

	if payload.Summary == "" {
		t.Error("summary field is required but empty")
	}

	if payload.Analysis == "" {
		t.Error("analysis field is required but empty")
	}

	// Verify summary contains test suite info (what failed)
	if !contains(payload.Summary, "quay.io/test") {
		t.Error("summary should contain image name")
	}
	if !contains(payload.Summary, "abc123") {
		t.Error("summary should contain commit")
	}
	if !contains(payload.Summary, "stage") {
		t.Error("summary should contain environment")
	}

	// Verify cluster_details contains cluster info (for debugging)
	if payload.ClusterDetails == "" {
		t.Error("cluster_details should not be empty when cluster info is provided")
	}
	if !contains(payload.ClusterDetails, "test-123") {
		t.Error("cluster_details should contain cluster ID")
	}
	if !contains(payload.ClusterDetails, "4.20") {
		t.Error("cluster_details should contain version")
	}

	// Verify analysis contains formatted content
	if !contains(payload.Analysis, "====== 🔍 Possible Cause ======") {
		t.Error("analysis should contain formatted root cause")
	}
	if !contains(payload.Analysis, "====== 💡 Recommendations ======") {
		t.Error("analysis should contain formatted recommendations")
	}

	// Verify optional fields
	if payload.Image != "quay.io/test:abc123" {
		t.Errorf("expected image quay.io/test:abc123, got %s", payload.Image)
	}

	if payload.Commit != "abc123" {
		t.Errorf("expected commit abc123, got %s", payload.Commit)
	}

	if payload.Env != "stage" {
		t.Errorf("expected env stage, got %s", payload.Env)
	}
}

func TestSlackReporter_Report_WorkflowFormat(t *testing.T) {
	// Create a test server to capture the webhook payload
	var capturedPayload WorkflowPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		if err := json.Unmarshal(body, &capturedPayload); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reporter := NewSlackReporter()
	result := &AnalysisResult{
		Content: "Analysis content here",
	}

	clusterInfo := &ClusterInfo{
		ID:      "test-456",
		Name:    "prod-cluster",
		Version: "4.21",
	}

	config := &ReporterConfig{
		Enabled: true,
		Settings: map[string]interface{}{
			"webhook_url":  server.URL,
			"channel":      "C123456",
			"cluster_info": clusterInfo,
			"image":        "quay.io/openshift/test:v1.0",
			"env":          "production",
		},
	}

	// Call Report
	if err := reporter.Report(context.Background(), result, config); err != nil {
		t.Fatalf("Report() failed: %v", err)
	}

	// Verify the captured payload
	if capturedPayload.Channel != "C123456" {
		t.Errorf("expected channel C123456, got %s", capturedPayload.Channel)
	}

	if capturedPayload.Summary == "" {
		t.Error("summary should not be empty")
	}

	if capturedPayload.Analysis == "" {
		t.Error("analysis should not be empty")
	}

	// Verify cluster info is in cluster_details field (not summary)
	if capturedPayload.ClusterDetails == "" {
		t.Error("cluster_details should not be empty when cluster info is provided")
	}
	if !contains(capturedPayload.ClusterDetails, "test-456") {
		t.Error("cluster_details should contain cluster ID")
	}

	// Verify summary contains test suite info
	if !contains(capturedPayload.Summary, "quay.io/openshift/test") {
		t.Error("summary should contain test image")
	}

	if capturedPayload.Image != "quay.io/openshift/test:v1.0" {
		t.Errorf("expected image quay.io/openshift/test:v1.0, got %s", capturedPayload.Image)
	}

	if capturedPayload.Commit != "v1.0" {
		t.Errorf("expected commit v1.0, got %s", capturedPayload.Commit)
	}
}

func TestSlackReporter_buildSummaryField(t *testing.T) {
	reporter := NewSlackReporter()

	clusterInfo := &ClusterInfo{
		ID:         "cluster-789",
		Name:       "my-test-cluster",
		Version:    "4.22",
		Provider:   "gcp",
		Expiration: "2026-02-01T12:00:00Z",
	}

	config := &ReporterConfig{
		Settings: map[string]interface{}{
			"cluster_info": clusterInfo,
			"image":        "quay.io/app:commit-xyz",
			"env":          "dev",
		},
	}

	summary := reporter.buildSummaryField(config)

	// Check for header
	if !contains(summary, ":failed:") {
		t.Error("summary should contain failure emoji")
	}
	if !contains(summary, "Pipeline Failed") {
		t.Error("summary should contain failure message")
	}

	// Summary should NOT contain cluster info (it's in cluster_details now)
	// Summary should ONLY contain test suite info (what failed)

	// Check for test suite info
	if !contains(summary, "quay.io/app") {
		t.Error("summary should contain test image")
	}
	if !contains(summary, "commit-xyz") {
		t.Error("summary should contain commit")
	}
	if !contains(summary, "dev") {
		t.Error("summary should contain environment")
	}
}

func TestSlackReporter_buildAnalysisField(t *testing.T) {
	reporter := NewSlackReporter()

	tests := []struct {
		name               string
		result             *AnalysisResult
		expectedContains   []string
		unexpectedContains []string
	}{
		{
			name: "formatted JSON analysis",
			result: &AnalysisResult{
				Content: "```json\n{\"root_cause\": \"Network issue\", \"recommendations\": [\"Fix network\"]}\n```",
			},
			expectedContains: []string{"====== 🔍 Possible Cause ======", "Network issue", "====== 💡 Recommendations ======", "Fix network"},
		},
		{
			name: "plain text analysis",
			result: &AnalysisResult{
				Content: "This is plain text analysis",
			},
			expectedContains: []string{"This is plain text analysis"},
		},
		{
			name: "analysis with error",
			result: &AnalysisResult{
				Content: "Analysis content",
				Error:   "Something went wrong",
			},
			expectedContains: []string{"Analysis content", "====== ⚠️ Error ======", "Something went wrong"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := reporter.buildAnalysisField(tt.result)

			for _, expected := range tt.expectedContains {
				if !contains(analysis, expected) {
					t.Errorf("analysis should contain %q", expected)
				}
			}

			for _, unexpected := range tt.unexpectedContains {
				if contains(analysis, unexpected) {
					t.Errorf("analysis should not contain %q", unexpected)
				}
			}
		})
	}
}

func TestSlackReporter_enforceFieldLimit(t *testing.T) {
	reporter := NewSlackReporter()

	tests := []struct {
		name      string
		content   string
		maxLength int
		wantLen   int
	}{
		{
			name:      "content within limit",
			content:   "short content",
			maxLength: 100,
			wantLen:   13,
		},
		{
			name:      "content exceeds limit",
			content:   string(make([]byte, 1000)),
			maxLength: 500,
			wantLen:   500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reporter.enforceFieldLimit(tt.content, tt.maxLength)

			if len(result) > tt.maxLength {
				t.Errorf("result length %d exceeds max length %d", len(result), tt.maxLength)
			}

			if tt.name == "content exceeds limit" {
				if !contains(result, "truncated") {
					t.Error("truncated content should contain notice")
				}
			}
		})
	}
}

func TestSlackReporter_ExtendedLogsFallback(t *testing.T) {
	reporter := NewSlackReporter()

	t.Run("no report_dir returns fallback message", func(t *testing.T) {
		result := &AnalysisResult{
			Content: "Test analysis",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url": "https://test.com",
				"channel":     "C123456",
				// No report_dir
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if payload.ExtendedLogs == "" {
			t.Error("ExtendedLogs should not be empty when no report_dir")
		}

		if !contains(payload.ExtendedLogs, "not available") {
			t.Errorf("Expected fallback message, got: %s", payload.ExtendedLogs)
		}
	})

	t.Run("empty report_dir returns fallback message", func(t *testing.T) {
		result := &AnalysisResult{
			Content: "Test analysis",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url": "https://test.com",
				"channel":     "C123456",
				"report_dir":  "", // Empty string
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if payload.ExtendedLogs == "" {
			t.Error("ExtendedLogs should not be empty when report_dir is empty string")
		}

		if !contains(payload.ExtendedLogs, "not available") {
			t.Errorf("Expected fallback message, got: %s", payload.ExtendedLogs)
		}
	})

	t.Run("nonexistent report_dir returns fallback message", func(t *testing.T) {
		result := &AnalysisResult{
			Content: "Test analysis",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url": "https://test.com",
				"channel":     "C123456",
				"report_dir":  "/nonexistent/path",
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if payload.ExtendedLogs == "" {
			t.Error("ExtendedLogs should not be empty when report_dir doesn't exist")
		}

		// When readTestOutput returns empty string, we get the "not found" fallback
		if !contains(payload.ExtendedLogs, "No test failure logs found") {
			t.Errorf("Expected fallback message, got: %s", payload.ExtendedLogs)
		}
	})

	t.Run("valid report_dir with logs returns actual logs", func(t *testing.T) {
		result := &AnalysisResult{
			Content: "Test analysis",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url": "https://test.com",
				"channel":     "C123456",
				"report_dir":  "testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws",
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if payload.ExtendedLogs == "" {
			t.Error("ExtendedLogs should not be empty when valid logs exist")
		}

		// Should contain actual failure logs, not fallback message
		if contains(payload.ExtendedLogs, "not available") {
			t.Errorf("Should contain real logs, not fallback. Got: %s", payload.ExtendedLogs[:100])
		}

		// Should contain failure markers from real data
		if !contains(payload.ExtendedLogs, "Found") && !contains(payload.ExtendedLogs, "test failure") {
			t.Error("Should contain failure count or marker")
		}
	})
}

func TestSlackReporter_ClusterDetailsFallback(t *testing.T) {
	reporter := NewSlackReporter()

	t.Run("no cluster_info returns fallback message", func(t *testing.T) {
		result := &AnalysisResult{
			Content: "Test analysis",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url": "https://test.com",
				"channel":     "C123456",
				// No cluster_info
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if payload.ClusterDetails == "" {
			t.Error("ClusterDetails should not be empty when no cluster_info")
		}

		if !contains(payload.ClusterDetails, "not available") {
			t.Errorf("Expected fallback message, got: %s", payload.ClusterDetails)
		}
	})

	t.Run("nil cluster_info returns fallback message", func(t *testing.T) {
		result := &AnalysisResult{
			Content: "Test analysis",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url":  "https://test.com",
				"channel":      "C123456",
				"cluster_info": nil,
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if payload.ClusterDetails == "" {
			t.Error("ClusterDetails should not be empty when cluster_info is nil")
		}

		if !contains(payload.ClusterDetails, "not available") {
			t.Errorf("Expected fallback message, got: %s", payload.ClusterDetails)
		}
	})

	t.Run("valid cluster_info returns cluster details", func(t *testing.T) {
		result := &AnalysisResult{
			Content: "Test analysis",
		}

		clusterInfo := &ClusterInfo{
			ID:       "test-cluster-123",
			Name:     "my-cluster",
			Version:  "4.20",
			Provider: "aws",
		}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url":  "https://test.com",
				"channel":      "C123456",
				"cluster_info": clusterInfo,
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if payload.ClusterDetails == "" {
			t.Error("ClusterDetails should not be empty when valid cluster_info exists")
		}

		// Should contain actual cluster info, not fallback message
		if contains(payload.ClusterDetails, "not available") {
			t.Errorf("Should contain real cluster info, not fallback. Got: %s", payload.ClusterDetails)
		}

		// Should contain cluster details
		if !contains(payload.ClusterDetails, "test-cluster-123") {
			t.Error("Should contain cluster ID")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
