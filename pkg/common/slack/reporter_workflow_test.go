package slack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			"repo":         "image quay.io/test",
			"commit":       "abc123",
			"env":          "stage",
		},
	}

	payload := reporter.buildWorkflowPayload(result, config)

	// Verify required fields
	if payload.Channel != "C06HQR8HN0L" {
		t.Errorf("expected channel C06HQR8HN0L, got %s", payload.Channel)
	}

	if payload.Analysis == "" {
		t.Error("analysis field is required but empty")
	}

	// Verify cluster_details contains cluster info (for debugging)
	if payload.ClusterDetails == "" {
		t.Error("cluster_details should not be empty when cluster info is provided")
	}
	if !strings.Contains(payload.ClusterDetails, "test-123") {
		t.Error("cluster_details should contain cluster ID")
	}
	if !strings.Contains(payload.ClusterDetails, "4.20") {
		t.Error("cluster_details should contain version")
	}

	// Verify analysis contains formatted content
	if !strings.Contains(payload.Analysis, "====== 🔍 Possible Cause ======") {
		t.Error("analysis should contain formatted root cause")
	}
	if !strings.Contains(payload.Analysis, "====== 💡 Recommendations ======") {
		t.Error("analysis should contain formatted recommendations")
	}

	// Verify optional fields
	assert.Equal(t, "image quay.io/test", payload.Image)
	assert.Equal(t, "abc123", payload.Commit)
	assert.Equal(t, "stage", payload.Env, "Env")
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

	if capturedPayload.Analysis == "" {
		t.Error("analysis should not be empty")
	}

	// Verify cluster info is in cluster_details field
	if capturedPayload.ClusterDetails == "" {
		t.Error("cluster_details should not be empty when cluster info is provided")
	}
	if !strings.Contains(capturedPayload.ClusterDetails, "test-456") {
		t.Error("cluster_details should contain cluster ID")
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
				if !strings.Contains(analysis, expected) {
					t.Errorf("analysis should contain %q", expected)
				}
			}

			for _, unexpected := range tt.unexpectedContains {
				if strings.Contains(analysis, unexpected) {
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
				if !strings.Contains(result, "truncated") {
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

		if !strings.Contains(payload.ExtendedLogs, "not available") {
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

		if !strings.Contains(payload.ExtendedLogs, "not available") {
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
		if !strings.Contains(payload.ExtendedLogs, "No test failure logs found") {
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
		if strings.Contains(payload.ExtendedLogs, "not available") {
			t.Errorf("Should contain real logs, not fallback. Got: %s", payload.ExtendedLogs[:100])
		}

		// Should contain failure markers from real data
		if !strings.Contains(payload.ExtendedLogs, "Found") && !strings.Contains(payload.ExtendedLogs, "test failure") {
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

		if !strings.Contains(payload.ClusterDetails, "not available") {
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

		if !strings.Contains(payload.ClusterDetails, "not available") {
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
		if strings.Contains(payload.ClusterDetails, "not available") {
			t.Errorf("Should contain real cluster info, not fallback. Got: %s", payload.ClusterDetails)
		}

		// Should contain cluster details
		if !strings.Contains(payload.ClusterDetails, "test-cluster-123") {
			t.Error("Should contain cluster ID")
		}
	})
}

func TestSlackReporter_ArtifactLinks(t *testing.T) {
	reporter := NewSlackReporter()

	t.Run("artifact links preferred over embedded logs", func(t *testing.T) {
		result := &AnalysisResult{Content: "Test analysis"}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url": "https://test.com",
				"channel":     "C123456",
				"report_dir":  "testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws",
				"artifact_links": []ArtifactLink{
					{Name: "test_output.log", URL: "https://s3.example.com/test_output.log?sig=abc", Size: 1024},
					{Name: "junit_e2e.xml", URL: "https://s3.example.com/junit_e2e.xml?sig=def", Size: 2048},
				},
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)
		// contains junit file
		assert.Contains(t, payload.JunitXMLLink, "junit_e2e.xml")
		// contains raw presigned url
		assert.Contains(t, payload.LogLink, "https://s3.example.com/test_output.log?sig=abc")
	})

	t.Run("falls back to embedded logs when no artifact links", func(t *testing.T) {
		result := &AnalysisResult{Content: "Test analysis"}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url": "https://test.com",
				"channel":     "C123456",
				"report_dir":  "testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws",
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if strings.Contains(payload.ExtendedLogs, "Artifacts") {
			t.Error("should not contain artifacts header without artifact links")
		}
	})

	t.Run("falls back to embedded logs with empty artifact links", func(t *testing.T) {
		result := &AnalysisResult{Content: "Test analysis"}

		config := &ReporterConfig{
			Settings: map[string]interface{}{
				"webhook_url":    "https://test.com",
				"channel":        "C123456",
				"report_dir":     "testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws",
				"artifact_links": []ArtifactLink{},
			},
		}

		payload := reporter.buildWorkflowPayload(result, config)

		if strings.Contains(payload.ExtendedLogs, "Artifacts") {
			t.Error("should not contain artifacts header with empty artifact links")
		}
	})
}
