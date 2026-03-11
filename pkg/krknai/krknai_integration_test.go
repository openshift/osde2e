package krknai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnalyzeLogs_NoReportDir tests error handling when REPORT_DIR is not set
func TestAnalyzeLogs_NoReportDir(t *testing.T) {
	oldReportDir := viper.GetString(config.ReportDir)
	defer viper.Set(config.ReportDir, oldReportDir)

	viper.Set(config.ReportDir, "")

	k := &KrknAI{
		result: &orchestrator.Result{},
	}

	err := k.AnalyzeLogs(context.Background(), fmt.Errorf("test error"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no report directory available")
}

// TestAnalyzeLogs_MissingAPIKey tests that analysis fails gracefully without API key
func TestAnalyzeLogs_MissingAPIKey(t *testing.T) {
	tempDir := t.TempDir()
	reportDir := filepath.Join(tempDir, "report")
	reportsDir := filepath.Join(reportDir, "reports")
	require.NoError(t, os.MkdirAll(reportsDir, 0o755))

	// Create minimal mock data
	createMinimalKrknAIResults(t, reportDir, reportsDir)

	// Setup config without API key
	oldConfig := captureViperConfig()
	defer restoreViperConfig(oldConfig)

	viper.Set(config.ReportDir, reportDir)
	viper.Set(config.Cluster.ID, "test-cluster-123")
	viper.Set(config.LogAnalysis.APIKey, "") // No API key
	viper.Set(config.Slack.Enable, false)

	k := &KrknAI{
		result: &orchestrator.Result{},
	}

	err := k.AnalyzeLogs(context.Background(), fmt.Errorf("test error"))

	// Should fail because no API key
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GEMINI_API_KEY is required")
}

// TestSlackNotificationConfig tests that Slack config is built correctly
func TestSlackNotificationConfig(t *testing.T) {
	clusterInfo := &slack.ClusterInfo{
		ID:            "test-cluster-123",
		Name:          "test-cluster",
		Provider:      "rosa",
		Region:        "us-east-1",
		CloudProvider: "aws",
		Version:       "4.17.3",
	}

	// Test with valid webhook and channel
	notificationConfig := slack.BuildNotificationConfig(
		"https://hooks.slack.com/test",
		"C12345",
		clusterInfo,
		"/tmp/report",
	)

	require.NotNil(t, notificationConfig)
	assert.True(t, notificationConfig.Enabled)
	assert.Len(t, notificationConfig.Reporters, 1)
	assert.Equal(t, "slack", notificationConfig.Reporters[0].Type)

	// Test with missing webhook - should return nil
	notificationConfig = slack.BuildNotificationConfig(
		"",
		"C12345",
		clusterInfo,
		"/tmp/report",
	)
	assert.Nil(t, notificationConfig)

	// Test with missing channel - should return nil
	notificationConfig = slack.BuildNotificationConfig(
		"https://hooks.slack.com/test",
		"",
		clusterInfo,
		"/tmp/report",
	)
	assert.Nil(t, notificationConfig)
}

// TestClusterInfoBuilding tests that cluster info is extracted correctly from viper
func TestClusterInfoBuilding(t *testing.T) {
	oldConfig := captureViperConfig()
	defer restoreViperConfig(oldConfig)

	viper.Set(config.Cluster.ID, "test-cluster-abc123")
	viper.Set(config.Cluster.Name, "my-test-cluster")
	viper.Set(config.Provider, "rosa")
	viper.Set(config.CloudProvider.Region, "us-west-2")
	viper.Set(config.CloudProvider.CloudProviderID, "aws")
	viper.Set(config.Cluster.Version, "4.17.3")

	clusterInfo := &slack.ClusterInfo{
		ID:            viper.GetString(config.Cluster.ID),
		Name:          viper.GetString(config.Cluster.Name),
		Provider:      viper.GetString(config.Provider),
		Region:        viper.GetString(config.CloudProvider.Region),
		CloudProvider: viper.GetString(config.CloudProvider.CloudProviderID),
		Version:       viper.GetString(config.Cluster.Version),
	}

	assert.Equal(t, "test-cluster-abc123", clusterInfo.ID)
	assert.Equal(t, "my-test-cluster", clusterInfo.Name)
	assert.Equal(t, "rosa", clusterInfo.Provider)
	assert.Equal(t, "us-west-2", clusterInfo.Region)
	assert.Equal(t, "aws", clusterInfo.CloudProvider)
	assert.Equal(t, "4.17.3", clusterInfo.Version)
}

// Helper functions

// copyTestFile copies a file from testdata to the destination
func copyTestFile(t *testing.T, filename, dest string) {
	t.Helper()

	src := filepath.Join("testdata", filename)
	data, err := os.ReadFile(src)
	require.NoError(t, err, "failed to read testdata file: %s", src)

	err = os.WriteFile(dest, data, 0o644)
	require.NoError(t, err, "failed to write test file: %s", dest)
}

// createMinimalKrknAIResults copies minimal test fixtures from testdata
func createMinimalKrknAIResults(t *testing.T, resultsDir, reportsDir string) {
	t.Helper()

	// Copy minimal fixtures from testdata directory
	copyTestFile(t, "minimal-all.csv", filepath.Join(reportsDir, "all.csv"))
	copyTestFile(t, "minimal-krkn-ai.yaml", filepath.Join(resultsDir, "krkn-ai.yaml"))
}

// setupFullTestData sets up complete test data from testdata fixtures
func setupFullTestData(t *testing.T) string {
	t.Helper()

	reportDir := t.TempDir()
	reportsDir := filepath.Join(reportDir, "reports")
	require.NoError(t, os.MkdirAll(reportsDir, 0o755))

	// Copy full test fixtures from testdata
	copyTestFile(t, "all.csv", filepath.Join(reportsDir, "all.csv"))
	copyTestFile(t, "health_check_report.csv", filepath.Join(reportsDir, "health_check_report.csv"))
	copyTestFile(t, "krkn-ai.yaml", filepath.Join(reportDir, "krkn-ai.yaml"))

	return reportDir
}

// TestAnalyzeLogs_WithRealLLM tests the full analysis workflow with a real LLM API call.
// This test requires LOG_ANALYSIS_API_KEY environment variable.
// To run: go test ./pkg/krknai -run TestAnalyzeLogs_WithRealLLM -v
func TestAnalyzeLogs_WithRealLLM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("LOG_ANALYSIS_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: LOG_ANALYSIS_API_KEY not set")
	}

	// Setup test data using testdata fixtures
	reportDir := setupFullTestData(t)

	// Configure viper
	oldConfig := captureViperConfig()
	defer restoreViperConfig(oldConfig)

	viper.Set(config.ReportDir, reportDir)
	viper.Set(config.Cluster.ID, "test-cluster-integration")
	viper.Set(config.Cluster.Name, "integration-test-cluster")
	viper.Set(config.Provider, "rosa")
	viper.Set(config.CloudProvider.Region, "us-east-1")
	viper.Set(config.CloudProvider.CloudProviderID, "aws")
	viper.Set(config.Cluster.Version, "4.17.3")
	viper.Set(config.LogAnalysis.APIKey, apiKey)
	viper.Set(config.Slack.Enable, false)

	// Test the full AnalyzeLogs workflow
	k := &KrknAI{
		result: &orchestrator.Result{},
	}

	ctx := context.Background()
	t.Log("Running analysis with real LLM (this may take 30-60 seconds)...")
	err := k.AnalyzeLogs(ctx, fmt.Errorf("test error"))

	// Should succeed with real API key and valid data
	require.NoError(t, err, "AnalyzeLogs should succeed with valid data and API key")

	// Verify summary file was created
	summaryPath := filepath.Join(reportDir, "llm-analysis", "summary.yaml")
	_, err = os.Stat(summaryPath)
	require.NoError(t, err, "summary.yaml should be created by AnalyzeLogs")

	t.Logf("✓ Analysis completed successfully!")
	t.Logf("✓ Summary written to: %s", summaryPath)
}

type viperConfig struct {
	reportDir         string
	clusterID         string
	clusterName       string
	provider          string
	region            string
	cloudProvider     string
	version           string
	apiKey            string
	enableSlackNotify bool
	slackWebhook      string
	slackChannel      string
}

func captureViperConfig() viperConfig {
	return viperConfig{
		reportDir:         viper.GetString(config.ReportDir),
		clusterID:         viper.GetString(config.Cluster.ID),
		clusterName:       viper.GetString(config.Cluster.Name),
		provider:          viper.GetString(config.Provider),
		region:            viper.GetString(config.CloudProvider.Region),
		cloudProvider:     viper.GetString(config.CloudProvider.CloudProviderID),
		version:           viper.GetString(config.Cluster.Version),
		apiKey:            viper.GetString(config.LogAnalysis.APIKey),
		enableSlackNotify: viper.GetBool(config.Slack.Enable),
		slackWebhook:      viper.GetString(config.Slack.Webhook),
		slackChannel:      viper.GetString(config.Slack.Channel),
	}
}

func restoreViperConfig(cfg viperConfig) {
	viper.Set(config.ReportDir, cfg.reportDir)
	viper.Set(config.Cluster.ID, cfg.clusterID)
	viper.Set(config.Cluster.Name, cfg.clusterName)
	viper.Set(config.Provider, cfg.provider)
	viper.Set(config.CloudProvider.Region, cfg.region)
	viper.Set(config.CloudProvider.CloudProviderID, cfg.cloudProvider)
	viper.Set(config.Cluster.Version, cfg.version)
	viper.Set(config.LogAnalysis.APIKey, cfg.apiKey)
	viper.Set(config.Slack.Enable, cfg.enableSlackNotify)
	viper.Set(config.Slack.Webhook, cfg.slackWebhook)
	viper.Set(config.Slack.Channel, cfg.slackChannel)
}
