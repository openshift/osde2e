package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/openshift/osde2e/internal/reporter"
)

// Slack Integration Tests
//
// These tests send real messages to Slack to verify the workflow integration works correctly.
//
// SETUP:
//
// 1. Add the E2E Test Notifications workflow to your test channel:
//    - Open: https://slack.com/shortcuts/Ft09RL7M2AMV/60f07b46919da20d103806a8f5bba094
//    - Click "Add to Slack"
//    - Select your test channel
//    - Copy the webhook URL (starts with https://hooks.slack.com/workflows/...)
//
// 2. Get your Slack channel ID (NOT the name):
//    - Right-click your test channel in Slack
//    - Select "View channel details"
//    - Scroll to bottom and copy the channel ID (starts with C, e.g., C06HQR8HN0L)
//
// 3. Set environment variables:
//    export LOG_ANALYSIS_SLACK_WEBHOOK="https://hooks.slack.com/workflows/..."
//    export LOG_ANALYSIS_SLACK_CHANNEL="C06HQR8HN0L"
//
// RUNNING THE TESTS:
//
//    # Run all integration tests
//    dotenv run go test -v ./pkg/e2e -run ^TestSlackReporter_Integration
//
//    # Run specific test
//    dotenv run go test -v ./pkg/e2e -run ^TestSlackReporter_Integration$
//
//    # Tests automatically skip if env vars are not set
//
// WHAT TO EXPECT IN SLACK:
//
//    Test 1 (Full): 3 threaded messages
//      - Initial: Failure summary with cluster info and test suite info
//      - Reply 1: AI analysis with root cause and recommendations
//      - Reply 2: Extracted test failure logs from real Prow data
//
//    Test 2 (Minimal): 3 threaded messages
//      - Initial: Failure summary with minimal cluster info
//      - Reply 1: Plain text analysis (no JSON formatting)
//      - Reply 2: Fallback message (no logs configured)
//
//    Test 3 (Error): 3 threaded messages
//      - Initial: Failure summary with cluster info
//      - Reply 1: Analysis with error section appended
//      - Reply 2: Fallback message (no logs in config)
//
// TROUBLESHOOTING:
//
//    - "400 Bad Request": Check that channel ID is correct (starts with C)
//    - "invalid_workflow_input": Channel name used instead of ID
//    - Messages not threaded: Wrong webhook type (must be workflow webhook)
//    - "invalid_blocks" in Slack: Empty field (now fixed with fallback messages)

// TestSlackReporter_Integration tests the Slack reporter with a real webhook.
//
// This test verifies that:
// 1. The workflow payload structure is correct
// 2. The webhook accepts the payload
// 3. Messages appear in the configured Slack channel
//
// Required environment variables:
//
//	LOG_ANALYSIS_SLACK_WEBHOOK - Workflow webhook URL
//	LOG_ANALYSIS_SLACK_CHANNEL - Channel ID (e.g., C06HQR8HN0L)
//
// To run:
//
//	export LOG_ANALYSIS_SLACK_WEBHOOK="https://hooks.slack.com/workflows/..."
//	export LOG_ANALYSIS_SLACK_CHANNEL="C06HQR8HN0L"
//	go test -v -run TestSlackReporter_Integration github.com/openshift/osde2e/pkg/e2e
func TestSlackReporter_Integration(t *testing.T) {
	webhookURL := os.Getenv("LOG_ANALYSIS_SLACK_WEBHOOK")
	channelID := os.Getenv("LOG_ANALYSIS_SLACK_CHANNEL")

	if webhookURL == "" || channelID == "" {
		t.Skip("Skipping integration test: LOG_ANALYSIS_SLACK_WEBHOOK or LOG_ANALYSIS_SLACK_CHANNEL not set")
	}

	// Create test cluster info
	clusterInfo := &reporter.ClusterInfo{
		ID:         "test-cluster-123",
		Name:       "integration-test-cluster",
		Version:    "4.20",
		Provider:   "aws",
		Expiration: "2026-02-01T00:00:00Z",
	}

	// Create analysis result with JSON content
	result := &reporter.AnalysisResult{
		Content: `Based on the test output, here is my analysis:

` + "```json" + `
{
  "root_cause": "Integration test: This is a test failure notification from the osde2e Slack reporter integration test",
  "recommendations": [
    "This is a test message to verify Slack Workflow integration",
    "Check that this message appears in the configured channel",
    "Verify that analysis and logs appear as threaded replies"
  ]
}
` + "```" + `
`,
	}

	// Create reporter config
	config := &reporter.ReporterConfig{
		Type:    "slack",
		Enabled: true,
		Settings: map[string]interface{}{
			"webhook_url":  webhookURL,
			"channel":      channelID,
			"cluster_info": clusterInfo,
			"image":        "quay.io/openshift/osde2e-tests:integration-test",
			"env":          "test",
			"report_dir":   "../../internal/reporter/testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws",
		},
	}

	// Send notification
	slackReporter := reporter.NewSlackReporter()
	ctx := context.Background()

	// Debug: Show what we're sending
	t.Log("=== SENDING PAYLOAD ===")
	t.Logf("Webhook URL: %s", webhookURL[:50]+"...")
	t.Logf("Channel ID: %s", channelID)

	err := slackReporter.Report(ctx, result, config)
	if err != nil {
		t.Logf("=== ERROR DETAILS ===")
		t.Logf("Error: %v", err)
		t.Fatalf("Failed to send Slack notification: %v", err)
	}

	t.Log("✅ Integration test successful!")
	t.Log("Check your Slack channel for the test message with threaded replies")
	t.Logf("Channel: %s", channelID)
	t.Log("Expected:")
	t.Log("  1. Initial message with cluster info and test suite info")
	t.Log("  2. First reply with AI analysis (root cause and recommendations)")
	t.Log("  3. Second reply with test failure logs (if testdata exists)")
}

// TestSlackReporter_Integration_MinimalPayload tests with minimal required fields.
func TestSlackReporter_Integration_MinimalPayload(t *testing.T) {
	webhookURL := os.Getenv("LOG_ANALYSIS_SLACK_WEBHOOK")
	channelID := os.Getenv("LOG_ANALYSIS_SLACK_CHANNEL")

	if webhookURL == "" || channelID == "" {
		t.Skip("Skipping integration test: LOG_ANALYSIS_SLACK_WEBHOOK or LOG_ANALYSIS_SLACK_CHANNEL not set")
	}

	// Minimal cluster info
	clusterInfo := &reporter.ClusterInfo{
		ID: "minimal-test-123",
	}

	// Plain text analysis (no JSON)
	result := &reporter.AnalysisResult{
		Content: "This is a minimal test with plain text analysis content. No JSON formatting.",
	}

	// Minimal config
	config := &reporter.ReporterConfig{
		Type:    "slack",
		Enabled: true,
		Settings: map[string]interface{}{
			"webhook_url":  webhookURL,
			"channel":      channelID,
			"cluster_info": clusterInfo,
		},
	}

	// Send notification
	slackReporter := reporter.NewSlackReporter()
	ctx := context.Background()

	err := slackReporter.Report(ctx, result, config)
	if err != nil {
		t.Fatalf("Failed to send minimal Slack notification: %v", err)
	}

	t.Log("✅ Minimal payload test successful!")
	t.Log("Check your Slack channel for the minimal test message")
}

// TestSlackReporter_Integration_WithError tests error handling and display.
func TestSlackReporter_Integration_WithError(t *testing.T) {
	webhookURL := os.Getenv("LOG_ANALYSIS_SLACK_WEBHOOK")
	channelID := os.Getenv("LOG_ANALYSIS_SLACK_CHANNEL")

	if webhookURL == "" || channelID == "" {
		t.Skip("Skipping integration test: LOG_ANALYSIS_SLACK_WEBHOOK or LOG_ANALYSIS_SLACK_CHANNEL not set")
	}

	clusterInfo := &reporter.ClusterInfo{
		ID:       "error-test-456",
		Name:     "error-handling-test",
		Version:  "4.21",
		Provider: "gcp",
	}

	// Analysis with error
	result := &reporter.AnalysisResult{
		Content: `Test analysis content with an error condition.

` + "```json" + `
{
  "root_cause": "Simulated error in test execution",
  "recommendations": ["Check error message below"]
}
` + "```" + `
`,
		Error: "Integration test: This is a simulated error message to verify error display in Slack",
	}

	config := &reporter.ReporterConfig{
		Type:    "slack",
		Enabled: true,
		Settings: map[string]interface{}{
			"webhook_url":  webhookURL,
			"channel":      channelID,
			"cluster_info": clusterInfo,
			"image":        "quay.io/openshift/osde2e-tests:error-test",
			"env":          "test",
		},
	}

	slackReporter := reporter.NewSlackReporter()
	ctx := context.Background()

	err := slackReporter.Report(ctx, result, config)
	if err != nil {
		t.Fatalf("Failed to send error Slack notification: %v", err)
	}

	t.Log("✅ Error handling test successful!")
	t.Log("Check your Slack channel - the analysis should include an 'Error' section")
}
