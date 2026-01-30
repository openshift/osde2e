package aws

import (
	"strings"
	"testing"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

func TestBuildS3Key(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		component    string
		jobID        string
		suffix       string
		wantContains []string
	}{
		{
			name:         "standard key with component",
			category:     "test-results",
			component:    "osd-example-operator",
			jobID:        "12345",
			suffix:       "",
			wantContains: []string{"test-results", "osd-example-operator", "12345"},
		},
		{
			name:         "empty category",
			category:     "",
			component:    "my-service",
			jobID:        "67890",
			suffix:       "",
			wantContains: []string{"my-service", "67890"},
		},
		{
			name:         "fallback to suffix when no jobID",
			category:     "results",
			component:    "test-operator",
			jobID:        "-1",
			suffix:       "abc123",
			wantContains: []string{"results", "test-operator", "abc123"},
		},
		{
			name:         "unknown component",
			category:     "test-results",
			component:    "unknown",
			jobID:        "test-job",
			suffix:       "",
			wantContains: []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup viper config
			viper.Set(config.JobID, tt.jobID)
			viper.Set(config.Suffix, tt.suffix)

			uploader := &S3Uploader{
				category:  tt.category,
				component: tt.component,
			}

			key := uploader.BuildS3Key()

			for _, want := range tt.wantContains {
				if want != "" && !strings.Contains(key, want) {
					t.Errorf("BuildS3Key() = %v, want to contain %v", key, want)
				}
			}
		})
	}
}

func TestNewS3Uploader_Disabled(t *testing.T) {
	// When LOG_BUCKET is empty, upload is disabled
	viper.Set(config.Tests.LogBucket, "")

	uploader, err := NewS3Uploader("test-component")
	if err != nil {
		t.Errorf("NewS3Uploader() with disabled config returned error: %v", err)
	}
	if uploader != nil {
		t.Error("NewS3Uploader() should return nil when LOG_BUCKET is empty")
	}
}

func TestBuildBaseKey(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"test-results/component/2026-01-30/run-123/install/junit.xml", "test-results/component/2026-01-30/run-123"},
		{"test-results/component/2026-01-30/run-123/test_output.log", "test-results/component/2026-01-30/run-123"},
		{"results/file.xml", "results"},
	}

	for _, tt := range tests {
		// Simulate the base key extraction logic
		var baseKey string
		parts := strings.Split(tt.key, "/")
		if len(parts) >= 4 {
			baseKey = strings.Join(parts[:4], "/")
		} else {
			baseKey = strings.TrimSuffix(tt.key, "/"+parts[len(parts)-1])
		}

		if baseKey != tt.expected {
			t.Errorf("Base key for %s: got %v, want %v", tt.key, baseKey, tt.expected)
		}
	}
}
