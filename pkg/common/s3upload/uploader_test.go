package s3upload

import (
	"testing"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

func TestBuildS3Key(t *testing.T) {
	tests := []struct {
		name         string
		prefix       string
		operatorName string
		jobID        string
		suffix       string
		wantContains []string
	}{
		{
			name:         "standard key with all values",
			prefix:       "test-results",
			operatorName: "osd-example-operator",
			jobID:        "12345",
			suffix:       "",
			wantContains: []string{"test-results", "osd-example-operator", "12345"},
		},
		{
			name:         "empty prefix",
			prefix:       "",
			operatorName: "my-operator",
			jobID:        "67890",
			suffix:       "",
			wantContains: []string{"my-operator", "67890"},
		},
		{
			name:         "fallback to suffix when no jobID",
			prefix:       "results",
			operatorName: "test-op",
			jobID:        "-1",
			suffix:       "abc123",
			wantContains: []string{"results", "test-op", "abc123"},
		},
		{
			name:         "missing operator name",
			prefix:       "test-results",
			operatorName: "",
			jobID:        "test-job",
			suffix:       "",
			wantContains: []string{"unknown-operator"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup viper config
			viper.Set(config.S3Upload.Enabled, true)
			viper.Set(config.S3Upload.Bucket, "test-bucket")
			viper.Set(config.S3Upload.Prefix, tt.prefix)
			viper.Set(config.S3Upload.OperatorName, tt.operatorName)
			viper.Set(config.JobID, tt.jobID)
			viper.Set(config.Suffix, tt.suffix)

			uploader := &Uploader{
				prefix: tt.prefix,
			}

			key := uploader.BuildS3Key()

			for _, want := range tt.wantContains {
				if want != "" && !contains(key, want) {
					t.Errorf("BuildS3Key() = %v, want to contain %v", key, want)
				}
			}
		})
	}
}

func TestNewUploader_Disabled(t *testing.T) {
	viper.Set(config.S3Upload.Enabled, false)

	uploader, err := NewUploader()
	if err != nil {
		t.Errorf("NewUploader() with disabled config returned error: %v", err)
	}
	if uploader != nil {
		t.Error("NewUploader() should return nil when disabled")
	}
}

func TestNewUploader_MissingBucket(t *testing.T) {
	viper.Set(config.S3Upload.Enabled, true)
	viper.Set(config.S3Upload.Bucket, "")

	_, err := NewUploader()
	if err == nil {
		t.Error("NewUploader() should return error when bucket is empty")
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(config.S3Upload.Enabled, tt.enabled)
			if got := IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractOperatorNameFromImage(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		wantName string
	}{
		{
			name:     "full image with e2e suffix",
			image:    "quay.io/org/osd-example-operator-e2e:v1.2.3",
			wantName: "osd-example-operator",
		},
		{
			name:     "full image with test suffix",
			image:    "quay.io/org/my-operator-test:latest",
			wantName: "my-operator",
		},
		{
			name:     "full image with tests suffix",
			image:    "quay.io/org/another-operator-tests:sha-abc123",
			wantName: "another-operator",
		},
		{
			name:     "full image with harness suffix",
			image:    "quay.io/org/test-harness:v2",
			wantName: "test",
		},
		{
			name:     "image without test suffix",
			image:    "quay.io/org/simple-operator:tag",
			wantName: "simple-operator",
		},
		{
			name:     "image without tag",
			image:    "quay.io/org/no-tag-operator",
			wantName: "no-tag-operator",
		},
		{
			name:     "image without registry",
			image:    "my-operator-e2e:latest",
			wantName: "my-operator",
		},
		{
			name:     "simple image name",
			image:    "operator-test",
			wantName: "operator",
		},
		{
			name:     "empty image",
			image:    "",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOperatorNameFromImage(tt.image)
			if got != tt.wantName {
				t.Errorf("extractOperatorNameFromImage(%q) = %q, want %q", tt.image, got, tt.wantName)
			}
		})
	}
}

func TestDeriveOperatorName(t *testing.T) {
	tests := []struct {
		name         string
		explicitName string
		testImage    string
		wantName     string
	}{
		{
			name:         "explicit operator name takes priority",
			explicitName: "explicit-operator",
			testImage:    "quay.io/org/other-operator-e2e:tag",
			wantName:     "explicit-operator",
		},
		{
			name:         "derive from test image when no explicit name",
			explicitName: "",
			testImage:    "quay.io/org/osd-example-operator-e2e:latest",
			wantName:     "osd-example-operator",
		},
		{
			name:         "fallback to unknown-operator",
			explicitName: "",
			testImage:    "",
			wantName:     "unknown-operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup viper config
			viper.Set(config.S3Upload.OperatorName, tt.explicitName)
			if tt.testImage != "" {
				viper.Set(config.Tests.AdHocTestImages, []string{tt.testImage})
			} else {
				viper.Set(config.Tests.AdHocTestImages, []string{})
			}

			got := deriveOperatorName()
			if got != tt.wantName {
				t.Errorf("deriveOperatorName() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
