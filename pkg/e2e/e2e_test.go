package e2e

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo/v2/types"
	"github.com/openshift/osde2e/pkg/common/aws"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/slack"
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
	viper.Set(config.Provider, "ocm")
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

	cfg := slack.BuildNotificationConfig("", "", nil, "")

	if cfg != nil {
		t.Error("Expected nil config when slack notifications disabled")
	}
}

func TestBuildNotificationConfig_MissingCredentials(t *testing.T) {
	setupTestConfig(t)

	cfg := slack.BuildNotificationConfig("", "", nil, "")

	if cfg != nil {
		t.Error("Expected nil config when webhook/channel missing")
	}
}

func TestBuildNotificationConfig_MissingWebhook(t *testing.T) {
	setupTestConfig(t)

	cfg := slack.BuildNotificationConfig("", "#test", nil, "")

	if cfg != nil {
		t.Error("Expected nil config when webhook missing")
	}
}

func TestBuildNotificationConfig_MissingChannel(t *testing.T) {
	setupTestConfig(t)

	cfg := slack.BuildNotificationConfig("https://hooks.slack.com/test", "", nil, "")

	if cfg != nil {
		t.Error("Expected nil config when channel missing")
	}
}

func TestBuildNotificationConfig_Enabled(t *testing.T) {
	setupTestConfig(t)

	cfg := slack.BuildNotificationConfig("https://hooks.slack.com/test", "#test-channel", nil, "")

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
	err := orch.PostProcessCluster(context.TODO())
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

	err := orch.PostProcessCluster(context.TODO())
	if err != nil {
		t.Errorf("Expected no error when must-gather is skipped, got: %v", err)
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
	err := orch.Cleanup(context.TODO())
	if err != nil {
		t.Errorf("Expected no error in dry run mode, got: %v", err)
	}
}

func TestExtractDetailsFromImage(t *testing.T) {
	tests := []struct {
		name          string
		image         string
		wantComponent string
		wantTag       string
	}{
		{
			name:          "standard image with e2e suffix",
			image:         "quay.io/org/osd-example-operator-e2e:v1.2.3",
			wantComponent: "osd-example-operator",
			wantTag:       "v1.2.3",
		},
		{
			name:          "image with test suffix",
			image:         "quay.io/org/my-service-test:latest",
			wantComponent: "my-service",
			wantTag:       "latest",
		},
		{
			name:          "image with tests suffix",
			image:         "quay.io/org/my-service-tests:abc123",
			wantComponent: "my-service",
			wantTag:       "abc123",
		},
		{
			name:          "image with harness suffix",
			image:         "quay.io/org/platform-harness:v1",
			wantComponent: "platform",
			wantTag:       "v1",
		},
		{
			name:          "image without test suffix",
			image:         "quay.io/org/simple:v1",
			wantComponent: "simple",
			wantTag:       "v1",
		},
		{
			name:          "image without tag",
			image:         "quay.io/org/my-operator-e2e",
			wantComponent: "my-operator",
			wantTag:       "",
		},
		{
			name:          "image without registry or org",
			image:         "my-image-test:latest",
			wantComponent: "my-image",
			wantTag:       "latest",
		},
		{
			name:          "empty string",
			image:         "",
			wantComponent: "",
			wantTag:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotComponent, gotTag := extractDetailsFromImage(tt.image)
			if gotComponent != tt.wantComponent {
				t.Errorf("extractDetailsFromImage(%q) component = %q, want %q", tt.image, gotComponent, tt.wantComponent)
			}
			if gotTag != tt.wantTag {
				t.Errorf("extractDetailsFromImage(%q) tag = %q, want %q", tt.image, gotTag, tt.wantTag)
			}
		})
	}
}

func TestS3ResultsToArtifactLinks(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Suffix, "abc123")

	results := []aws.S3UploadResult{
		{Key: "test-results/comp/2026-01-30/123/test_output.log", PresignedURL: "https://s3/log", Size: 5000},
		{Key: "test-results/comp/2026-01-30/123/install/junit_abc123.xml", PresignedURL: "https://s3/junit", Size: 200},
		{Key: "test-results/comp/2026-01-30/123/install/operator-a-e2e/results.xml", PresignedURL: "https://s3/other", Size: 100},
	}

	links := s3ResultsToArtifactLinks(results)
	if len(links) != 2 {
		t.Fatalf("Expected 2 links (test_output.log + junit), got %d", len(links))
	}
	if links[0].Name != "test_output.log" {
		t.Errorf("links[0].Name = %q, want %q", links[0].Name, "test_output.log")
	}
	if links[1].Name != "junit_abc123.xml" {
		t.Errorf("links[1].Name = %q, want %q", links[1].Name, "junit_abc123.xml")
	}
}

func TestS3ResultsToArtifactLinks_EmptyPresignedURL(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Suffix, "test")

	results := []aws.S3UploadResult{
		{Key: "test-results/comp/2026-01-30/123/test_output.log", PresignedURL: "", Size: 5000},
	}

	links := s3ResultsToArtifactLinks(results)
	if len(links) != 0 {
		t.Errorf("Expected 0 links when presigned URLs are empty, got %d", len(links))
	}
}

func TestS3ResultsToArtifactLinks_Empty(t *testing.T) {
	setupTestConfig(t)

	links := s3ResultsToArtifactLinks(nil)
	if len(links) != 0 {
		t.Errorf("Expected 0 links for nil input, got %d", len(links))
	}
}

func TestGlobalTestOutputLink_EmptyPresignedURL(t *testing.T) {
	results := []aws.S3UploadResult{
		{Key: "test-results/comp/2026-01-30/123/test_output.log", PresignedURL: "", Size: 5000},
	}

	link := globalTestOutputLink(results)
	if link != nil {
		t.Errorf("Expected nil link when presigned URL is empty, got %+v", link)
	}
}

func TestGlobalTestOutputLink_PreservesSize(t *testing.T) {
	results := []aws.S3UploadResult{
		{Key: "test-results/comp/2026-01-30/123/test_output.log", PresignedURL: "https://s3/log", Size: 12345},
	}

	link := globalTestOutputLink(results)
	if link == nil {
		t.Fatal("Expected non-nil link")
	}
	if link.Size != 12345 {
		t.Errorf("link.Size = %d, want %d", link.Size, 12345)
	}
}

func TestSuiteArtifactLinks_PreservesOrder(t *testing.T) {
	results := []aws.S3UploadResult{
		{Key: "test-results/op-a/2026-01-30/123/alpha.xml", PresignedURL: "https://s3/a", Size: 10},
		{Key: "test-results/op-a/2026-01-30/123/beta.log", PresignedURL: "https://s3/b", Size: 20},
		{Key: "test-results/op-a/2026-01-30/123/gamma.json", PresignedURL: "https://s3/c", Size: 30},
	}

	links := suiteArtifactLinks(results)
	if len(links) != 3 {
		t.Fatalf("Expected 3 links, got %d", len(links))
	}
	expected := []string{"alpha.xml", "beta.log", "gamma.json"}
	for i, want := range expected {
		if links[i].Name != want {
			t.Errorf("links[%d].Name = %q, want %q", i, links[i].Name, want)
		}
	}
}

func TestDeriveGlobalComponent_LegacyAdHocTestImages(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Tests.AdHocTestImages, "quay.io/org/legacy-operator-e2e:v1")
	viper.Set(config.JobName, "")

	got := deriveGlobalComponent()
	if got != "legacy-operator" {
		t.Errorf("deriveGlobalComponent() with legacy format = %q, want %q", got, "legacy-operator")
	}
}

func TestDeriveGlobalComponent_SingleSuite(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Tests.TestSuites, []config.TestSuite{
		{Image: "quay.io/org/osd-example-operator-e2e:latest"},
	})

	got := deriveGlobalComponent()
	if got != "osd-example-operator" {
		t.Errorf("deriveGlobalComponent() = %q, want %q", got, "osd-example-operator")
	}
}

func TestDeriveGlobalComponent_MultiSuiteWithJobName(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Tests.TestSuites, []config.TestSuite{
		{Image: "quay.io/org/operator-a-e2e:v1"},
		{Image: "quay.io/org/operator-b-e2e:v2"},
	})
	viper.Set(config.JobName, "osde2e-prod-gcp-nightly")

	got := deriveGlobalComponent()
	if got != "osde2e-prod-gcp-nightly" {
		t.Errorf("deriveGlobalComponent() = %q, want %q", got, "osde2e-prod-gcp-nightly")
	}
}

func TestDeriveGlobalComponent_MultiSuiteNoJobName(t *testing.T) {
	setupTestConfig(t)
	viper.Set(config.Tests.TestSuites, []config.TestSuite{
		{Image: "quay.io/org/operator-a-e2e:v1"},
		{Image: "quay.io/org/operator-b-e2e:v2"},
	})
	viper.Set(config.JobName, "")

	got := deriveGlobalComponent()
	if got != "operator-a" {
		t.Errorf("deriveGlobalComponent() = %q, want %q (fallback to first image)", got, "operator-a")
	}
}

func TestDeriveGlobalComponent_NoSuites(t *testing.T) {
	setupTestConfig(t)

	got := deriveGlobalComponent()
	if got != "unknown" {
		t.Errorf("deriveGlobalComponent() = %q, want %q", got, "unknown")
	}
}

func TestGlobalTestOutputLink(t *testing.T) {
	results := []aws.S3UploadResult{
		{Key: "test-results/comp/2026-01-30/123/install/junit_abc.xml", PresignedURL: "https://s3/junit", Size: 100},
		{Key: "test-results/comp/2026-01-30/123/test_output.log", PresignedURL: "https://s3/log", Size: 5000},
	}

	link := globalTestOutputLink(results)
	if link == nil {
		t.Fatal("Expected non-nil link for test_output.log")
	}
	if link.Name != "test_output.log" {
		t.Errorf("link.Name = %q, want %q", link.Name, "test_output.log")
	}
	if link.URL != "https://s3/log" {
		t.Errorf("link.URL = %q, want %q", link.URL, "https://s3/log")
	}
}

func TestGlobalTestOutputLink_NotFound(t *testing.T) {
	results := []aws.S3UploadResult{
		{Key: "test-results/comp/2026-01-30/123/install/junit_abc.xml", PresignedURL: "https://s3/junit", Size: 100},
	}

	link := globalTestOutputLink(results)
	if link != nil {
		t.Errorf("Expected nil link when test_output.log not present, got %+v", link)
	}
}

func TestSuiteArtifactLinks(t *testing.T) {
	results := []aws.S3UploadResult{
		{Key: "test-results/operator-a/2026-01-30/123/junit.xml", PresignedURL: "https://s3/junit", Size: 100},
		{Key: "test-results/operator-a/2026-01-30/123/executor.log", PresignedURL: "https://s3/exec", Size: 2000},
		{Key: "test-results/operator-a/2026-01-30/123/no-url.log", PresignedURL: "", Size: 500},
	}

	links := suiteArtifactLinks(results)
	if len(links) != 2 {
		t.Fatalf("Expected 2 links (skipping empty presigned URL), got %d", len(links))
	}
	if links[0].Name != "junit.xml" {
		t.Errorf("links[0].Name = %q, want %q", links[0].Name, "junit.xml")
	}
	if links[1].Name != "executor.log" {
		t.Errorf("links[1].Name = %q, want %q", links[1].Name, "executor.log")
	}
}

func TestSuiteArtifactLinks_Empty(t *testing.T) {
	links := suiteArtifactLinks(nil)
	if len(links) != 0 {
		t.Errorf("Expected 0 links for nil input, got %d", len(links))
	}
}

func TestArtifactLinksForSuite_PerSuite(t *testing.T) {
	orch := &E2EOrchestrator{
		perSuiteS3Results: map[string][]aws.S3UploadResult{
			"quay.io/org/op-e2e:abc": {
				{Key: "test-results/op-abc/junit.xml", PresignedURL: "https://s3/junit", Size: 100},
				{Key: "test-results/op-abc/exec.log", PresignedURL: "https://s3/exec", Size: 200},
			},
		},
		result: &orchestrator.Result{ExitCode: config.Success},
	}
	globalLink := &slack.ArtifactLink{Name: "test_output.log", URL: "https://s3/log", Size: 5000}
	fallback := []slack.ArtifactLink{{Name: "fallback.log", URL: "https://s3/fb"}}

	links := orch.artifactLinksForSuite("quay.io/org/op-e2e:abc", globalLink, fallback)
	if len(links) != 3 {
		t.Fatalf("Expected 3 links (1 global + 2 per-suite), got %d", len(links))
	}
	if links[0].Name != "test_output.log" {
		t.Errorf("links[0] = %q, want global test_output.log", links[0].Name)
	}
	if links[1].Name != "junit.xml" {
		t.Errorf("links[1] = %q, want junit.xml", links[1].Name)
	}
}

func TestArtifactLinksForSuite_Fallback(t *testing.T) {
	orch := &E2EOrchestrator{
		perSuiteS3Results: map[string][]aws.S3UploadResult{},
		result:            &orchestrator.Result{ExitCode: config.Success},
	}
	fallback := []slack.ArtifactLink{{Name: "test_output.log", URL: "https://s3/log"}}

	links := orch.artifactLinksForSuite("quay.io/org/unknown:xyz", nil, fallback)
	if len(links) != 1 || links[0].Name != "test_output.log" {
		t.Errorf("Expected fallback links, got %v", links)
	}
}

func TestArtifactLinksForSuite_NoGlobalLink(t *testing.T) {
	orch := &E2EOrchestrator{
		perSuiteS3Results: map[string][]aws.S3UploadResult{
			"img:v1": {
				{Key: "k/junit.xml", PresignedURL: "https://s3/j", Size: 50},
			},
		},
		result: &orchestrator.Result{ExitCode: config.Success},
	}

	links := orch.artifactLinksForSuite("img:v1", nil, nil)
	if len(links) != 1 || links[0].Name != "junit.xml" {
		t.Errorf("Expected 1 per-suite link without global, got %v", links)
	}
}
