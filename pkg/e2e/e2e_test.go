package e2e

import (
	"testing"

	"github.com/onsi/ginkgo/v2/types"
	"github.com/openshift/osde2e/internal/reporter"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/runner"
)

func setupTestConfig(t *testing.T) {
	t.Helper()
	viper.Reset()
	viper.Set(config.ReportDir, t.TempDir())
	viper.Set(config.Suffix, "test")
	viper.Set(config.Tests.SuiteTimeout, 1)
	viper.Set(config.Tests.GinkgoLogLevel, "succinct")
	viper.Set(config.DryRun, true)
	viper.Set(config.Cluster.ID, "test-cluster-123")
	viper.Set(config.Cluster.Name, "test-cluster")
	viper.Set(config.Provider, "mock")
	viper.Set(config.SkipMustGather, true)
	viper.Set(config.Cluster.SkipDestroyCluster, true)
}

func TestConfigureGinkgo_DryRun(t *testing.T) {
	setupTestConfig(t)

	var suiteConfig types.SuiteConfig
	var reporterConfig types.ReporterConfig

	configureGinkgo(&suiteConfig, &reporterConfig)

	if !suiteConfig.DryRun {
		t.Error("Expected DryRun to be true")
	}
	if !reporterConfig.NoColor {
		t.Error("Expected NoColor to be true")
	}
	if !reporterConfig.Succinct {
		t.Error("Expected Succinct reporter mode")
	}
}

func TestConfigureGinkgo_VerboseModes(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantVV   bool
		wantV    bool
	}{
		{"verbose", "v", false, true},
		{"very verbose", "vv", true, false},
		{"succinct", "succinct", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)
			viper.Set(config.Tests.GinkgoLogLevel, tt.logLevel)

			var suiteConfig types.SuiteConfig
			var reporterConfig types.ReporterConfig
			configureGinkgo(&suiteConfig, &reporterConfig)

			if reporterConfig.VeryVerbose != tt.wantVV {
				t.Errorf("VeryVerbose = %v, want %v", reporterConfig.VeryVerbose, tt.wantVV)
			}
			if reporterConfig.Verbose != tt.wantV {
				t.Errorf("Verbose = %v, want %v", reporterConfig.Verbose, tt.wantV)
			}
		})
	}
}

func TestConfigureGinkgo_Filters(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Tests.GinkgoSkip, "skip-pattern")
	viper.Set(config.Tests.GinkgoFocus, "focus-pattern")
	viper.Set(config.Tests.GinkgoLabelFilter, "label-filter")
	viper.Set(config.Tests.TestsToRun, []string{"test1", "test2"})

	var suiteConfig types.SuiteConfig
	var reporterConfig types.ReporterConfig
	configureGinkgo(&suiteConfig, &reporterConfig)

	if len(suiteConfig.SkipStrings) == 0 || suiteConfig.SkipStrings[0] != "skip-pattern" {
		t.Error("SkipStrings not configured correctly")
	}
	if len(suiteConfig.FocusStrings) != 3 {
		t.Errorf("Expected 3 FocusStrings (2 tests + 1 focus), got %d", len(suiteConfig.FocusStrings))
	}
	if suiteConfig.LabelFilter != "label-filter" {
		t.Error("LabelFilter not configured correctly")
	}
}

func TestConfigureGinkgo_Timeout(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Tests.SuiteTimeout, 5)

	var suiteConfig types.SuiteConfig
	var reporterConfig types.ReporterConfig
	configureGinkgo(&suiteConfig, &reporterConfig)

	expectedTimeout := 5 * 60 * 60 * 1000000000 // 5 hours in nanoseconds
	if int64(suiteConfig.Timeout) != int64(expectedTimeout) {
		t.Errorf("Expected timeout %d, got %d", expectedTimeout, suiteConfig.Timeout)
	}
}

func TestE2EOrchestrator_Result(t *testing.T) {
	setupTestConfig(t)

	orch := &E2EOrchestrator{
		result: &orchestrator.Result{
			ExitCode:    config.Success,
			TestsPassed: true,
			ClusterID:   "test-123",
		},
	}

	result := orch.Result()
	if result.ExitCode != config.Success {
		t.Errorf("Expected exit code %d, got %d", config.Success, result.ExitCode)
	}
	if !result.TestsPassed {
		t.Error("Expected TestsPassed to be true")
	}
	if result.ClusterID != "test-123" {
		t.Errorf("Expected cluster ID 'test-123', got '%s'", result.ClusterID)
	}
}

func TestBuildNotificationConfig_Disabled(t *testing.T) {
	setupTestConfig(t)

	cfg := reporter.BuildNotificationConfig("", "", nil, "", "")

	if cfg != nil {
		t.Error("Expected nil config when slack notifications disabled")
	}
}

func TestBuildNotificationConfig_MissingCredentials(t *testing.T) {
	setupTestConfig(t)

	cfg := reporter.BuildNotificationConfig("", "", nil, "", "")

	if cfg != nil {
		t.Error("Expected nil config when webhook/channel missing")
	}
}

func TestBuildNotificationConfig_MissingWebhook(t *testing.T) {
	setupTestConfig(t)

	cfg := reporter.BuildNotificationConfig("", "#test", nil, "", "")

	if cfg != nil {
		t.Error("Expected nil config when webhook missing")
	}
}

func TestBuildNotificationConfig_MissingChannel(t *testing.T) {
	setupTestConfig(t)

	cfg := reporter.BuildNotificationConfig("https://hooks.slack.com/test", "", nil, "", "")

	if cfg != nil {
		t.Error("Expected nil config when channel missing")
	}
}

func TestBuildNotificationConfig_Enabled(t *testing.T) {
	setupTestConfig(t)

	cfg := reporter.BuildNotificationConfig("https://hooks.slack.com/test", "#test-channel", nil, "", "")

	if cfg == nil {
		t.Fatal("Expected non-nil notification config")
	}
	if !cfg.Enabled {
		t.Error("Expected config to be enabled")
	}
	if len(cfg.Reporters) != 1 {
		t.Errorf("Expected 1 reporter, got %d", len(cfg.Reporters))
	}
}

func TestWriteLogs(t *testing.T) {
	setupTestConfig(t)

	logs := map[string][]byte{
		"install": []byte("install log content"),
		"upgrade": []byte("upgrade log content"),
	}

	// Should not panic
	runner.WriteLogs(logs)
}

func TestCollectAndWriteLogs_NoProvider(t *testing.T) {
	setupTestConfig(t)

	// Should not panic with nil provider
	runner.ReportClusterInstallLogs(nil)
}

func TestCollectAndWriteLogs_EmptyClusterID(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Cluster.ID, "")

	// Should not panic with empty cluster ID
	runner.ReportClusterInstallLogs(nil)
}

func TestPostProcessCluster_DryRun(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.DryRun, true)

	orch := &E2EOrchestrator{
		suiteConfig: types.SuiteConfig{
			DryRun: true,
		},
		result: &orchestrator.Result{
			ExitCode: config.Success,
		},
	}

	// Should return immediately in dry run mode
	err := orch.PostProcessCluster(nil)
	if err != nil {
		t.Errorf("Expected no error in dry run mode, got: %v", err)
	}
}

func TestPostProcessCluster_SkipMustGather(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.DryRun, false)
	viper.Set(config.SkipMustGather, true)

	orch := &E2EOrchestrator{
		suiteConfig: types.SuiteConfig{
			DryRun: false,
		},
		result: &orchestrator.Result{
			ExitCode: config.Success,
			Errors:   []error{},
		},
	}

	// Should handle helper creation failure gracefully
	err := orch.PostProcessCluster(nil)
	if err == nil {
		t.Error("Expected error when helper creation fails")
	}
}

func TestCleanup_DryRun(t *testing.T) {
	setupTestConfig(t)

	orch := &E2EOrchestrator{
		suiteConfig: types.SuiteConfig{
			DryRun: true,
		},
		result: &orchestrator.Result{
			ExitCode: config.Success,
		},
	}

	// Should return immediately in dry run mode
	err := orch.Cleanup(nil)
	if err != nil {
		t.Errorf("Expected no error in dry run mode, got: %v", err)
	}
}
