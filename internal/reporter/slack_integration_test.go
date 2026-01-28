package reporter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestSlackReporter_Integration is an integration test for the Slack reporter.
// This test will actually send a message to Slack if properly configured.
//
// To run this test with dotenv:
//
//	dotenv go test -v ./internal/reporter -run TestSlackReporter_Integration
//
// Or set environment variables manually:
//
//	export LOG_ANALYSIS_SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
//	export LOG_ANALYSIS_SLACK_CHANNEL="#your-test-channel"
//	export LOG_ANALYSIS_SLACK_BOT_TOKEN="xoxb-your-bot-token"  # Optional, for file attachments
//	go test -v ./internal/reporter -run TestSlackReporter_Integration
//
// The test will skip if required environment variables are not set.
func TestSlackReporter_Integration(t *testing.T) {
	// Use the actual environment variables from viper config (pkg/common/config/config.go:955-961)
	webhookURL := os.Getenv("LOG_ANALYSIS_SLACK_WEBHOOK")
	if webhookURL == "" {
		t.Skip("Skipping integration test: LOG_ANALYSIS_SLACK_WEBHOOK not set")
	}

	channel := os.Getenv("LOG_ANALYSIS_SLACK_CHANNEL")
	if channel == "" {
		t.Skip("Skipping integration test: LOG_ANALYSIS_SLACK_CHANNEL not set")
	}

	botToken := os.Getenv("LOG_ANALYSIS_SLACK_BOT_TOKEN") // Optional

	// Use real testdata from Prow build
	reportDir := setupTestReportDirFromTestdata(t)

	// Create a realistic cluster info
	clusterInfo := &ClusterInfo{
		ID:            "2ntr2hoo8487ite28bd98pg5ph0m04gf",
		Name:          "integration-test-cluster",
		Provider:      "AWS",
		Region:        "us-east-1",
		CloudProvider: "aws",
		Version:       "4.20-nightly",
		Expiration:    "2026-01-30T12:00:00Z",
	}

	// Create a realistic analysis result
	analysisResult := &AnalysisResult{
		Content: `Analysis of test failure from periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws:

` + "```json" + `
{
  "root_cause": "Integration test - verifying Slack notification system with real Prow testdata",
  "recommendations": [
    "This is a test message from the osde2e Slack reporter integration test",
    "Check that the test_output.log shows real Prow build output",
    "Verify file attachments are included (if bot token is configured)",
    "Review the formatted analysis and cluster information"
  ]
}
` + "```",
		Error: "",
	}

	// Build the reporter config
	config := &ReporterConfig{
		Type:    "slack",
		Enabled: true,
		Settings: map[string]interface{}{
			"webhook_url":  webhookURL,
			"channel":      channel,
			"cluster_info": clusterInfo,
			"report_dir":   reportDir,
		},
	}

	// Add bot token if available
	if botToken != "" {
		config.Settings["bot_token"] = botToken
		t.Logf("Bot token configured - files will be attached")
	} else {
		t.Logf("No bot token - using webhook fallback (no file attachments)")
	}

	// Create the reporter and send the notification
	reporter := NewSlackReporter()
	ctx := context.Background()

	t.Logf("Sending test notification to Slack channel: %s", channel)
	t.Logf("Using real testdata from: testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws/")

	err := reporter.Report(ctx, analysisResult, config)
	if err != nil {
		t.Fatalf("Failed to send Slack notification: %v", err)
	}

	t.Logf("✅ Successfully sent test notification to Slack!")
	t.Logf("Check your Slack channel: %s", channel)
}

// TestSlackReporter_Integration_WithCustomFilters tests the file filtering functionality
func TestSlackReporter_Integration_WithCustomFilters(t *testing.T) {
	// Use the actual environment variables from viper config
	webhookURL := os.Getenv("LOG_ANALYSIS_SLACK_WEBHOOK")
	if webhookURL == "" {
		t.Skip("Skipping integration test: LOG_ANALYSIS_SLACK_WEBHOOK not set")
	}

	channel := os.Getenv("LOG_ANALYSIS_SLACK_CHANNEL")
	if channel == "" {
		t.Skip("Skipping integration test: LOG_ANALYSIS_SLACK_CHANNEL not set")
	}

	botToken := os.Getenv("LOG_ANALYSIS_SLACK_BOT_TOKEN")
	if botToken == "" {
		t.Skip("Skipping file filter test: LOG_ANALYSIS_SLACK_BOT_TOKEN required for file attachments")
	}

	// Use real testdata from Prow build
	reportDir := setupTestReportDirFromTestdata(t)

	clusterInfo := &ClusterInfo{
		ID:       "2ntr2hoo8487ite28bd98pg5ph0m04gf",
		Name:     "file-filter-test",
		Provider: "AWS",
		Version:  "4.20-nightly",
	}

	analysisResult := &AnalysisResult{
		Content: `Testing file filtering with real Prow testdata:

This test verifies that the Slack reporter correctly filters files based on:
- File name patterns (only test_output.log and junit*.xml)
- Maximum file count (limited to 3 files)
- Maximum total size (limited to 10MB)

Check the attachments - you should only see the filtered files.`,
	}

	// Test with custom file filters
	config := &ReporterConfig{
		Type:    "slack",
		Enabled: true,
		Settings: map[string]interface{}{
			"webhook_url":       webhookURL,
			"channel":           channel,
			"bot_token":         botToken,
			"cluster_info":      clusterInfo,
			"report_dir":        reportDir,
			"log_file_patterns": []string{"test_output.log"}, // Only test_output.log
			"max_log_files":     3,                           // Limit to 3 files
			"max_log_size_mb":   10,                          // Limit to 10MB
		},
	}

	reporter := NewSlackReporter()
	ctx := context.Background()

	t.Logf("Sending test notification with custom file filters")
	t.Logf("Patterns: %v", config.Settings["log_file_patterns"])
	t.Logf("Max files: %d", config.Settings["max_log_files"])

	err := reporter.Report(ctx, analysisResult, config)
	if err != nil {
		t.Fatalf("Failed to send Slack notification: %v", err)
	}

	t.Logf("✅ Successfully sent test notification with filtered files!")
	t.Logf("Check your Slack channel: %s", channel)
	t.Logf("You should only see test_output.log attached")
}

// setupTestReportDirFromTestdata copies real Prow testdata to a temp directory
func setupTestReportDirFromTestdata(t *testing.T) string {
	tmpDir := t.TempDir()

	// Path to real testdata
	testdataPath := "testdata/periodic-ci-openshift-osde2e-main-nightly-4.20-osd-aws"

	// Copy build-log.txt as test_output.log
	buildLogPath := filepath.Join(testdataPath, "build-log.txt")
	buildLogContent, err := os.ReadFile(buildLogPath)
	if err != nil {
		t.Fatalf("Failed to read testdata build-log.txt: %v", err)
	}

	destPath := filepath.Join(tmpDir, "test_output.log")
	if err := os.WriteFile(destPath, buildLogContent, 0o644); err != nil {
		t.Fatalf("Failed to create test_output.log: %v", err)
	}

	t.Logf("Using real Prow testdata (%d bytes)", len(buildLogContent))

	return tmpDir
}
