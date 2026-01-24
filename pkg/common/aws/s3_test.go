package aws

import (
	"testing"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

func TestBuildS3Key(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		testImage    string
		jobID        string
		suffix       string
		wantContains []string
	}{
		{
			name:         "standard key with test image",
			category:     "test-results",
			testImage:    "quay.io/org/osd-example-operator-e2e:v1.0",
			jobID:        "12345",
			suffix:       "",
			wantContains: []string{"test-results", "osd-example-operator", "12345"},
		},
		{
			name:         "empty category",
			category:     "",
			testImage:    "quay.io/org/my-service-test:latest",
			jobID:        "67890",
			suffix:       "",
			wantContains: []string{"my-service", "67890"},
		},
		{
			name:         "fallback to suffix when no jobID",
			category:     "results",
			testImage:    "quay.io/org/test-operator-e2e:tag",
			jobID:        "-1",
			suffix:       "abc123",
			wantContains: []string{"results", "test-operator", "abc123"},
		},
		{
			name:         "fallback to unknown when no test image",
			category:     "test-results",
			testImage:    "",
			jobID:        "test-job",
			suffix:       "",
			wantContains: []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup viper config
			viper.Set(config.Tests.LogBucket, "test-bucket")
			viper.Set(config.JobID, tt.jobID)
			viper.Set(config.Suffix, tt.suffix)
			if tt.testImage != "" {
				viper.Set(config.Tests.AdHocTestImages, []string{tt.testImage})
			} else {
				viper.Set(config.Tests.AdHocTestImages, []string{})
			}

			uploader := &S3Uploader{
				category: tt.category, // test with different categories
			}

			key := uploader.BuildS3Key()

			for _, want := range tt.wantContains {
				if want != "" && !containsSubstr(key, want) {
					t.Errorf("BuildS3Key() = %v, want to contain %v", key, want)
				}
			}
		})
	}
}

func TestNewS3Uploader_Disabled(t *testing.T) {
	// When LOG_BUCKET is empty, upload is disabled
	viper.Set(config.Tests.LogBucket, "")

	uploader, err := NewS3Uploader()
	if err != nil {
		t.Errorf("NewS3Uploader() with disabled config returned error: %v", err)
	}
	if uploader != nil {
		t.Error("NewS3Uploader() should return nil when LOG_BUCKET is empty")
	}
}

func TestIsS3UploadEnabled(t *testing.T) {
	tests := []struct {
		name   string
		bucket string
		want   bool
	}{
		{"enabled when bucket set", "test-bucket", true},
		{"disabled when bucket empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(config.Tests.LogBucket, tt.bucket)
			if got := IsS3UploadEnabled(); got != tt.want {
				t.Errorf("IsS3UploadEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractNameFromImage(t *testing.T) {
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
			image:    "quay.io/org/my-service-test:latest",
			wantName: "my-service",
		},
		{
			name:     "full image with tests suffix",
			image:    "quay.io/org/another-service-tests:sha-abc123",
			wantName: "another-service",
		},
		{
			name:     "full image with harness suffix",
			image:    "quay.io/org/test-harness:v2",
			wantName: "test",
		},
		{
			name:     "image without test suffix",
			image:    "quay.io/org/simple-app:tag",
			wantName: "simple-app",
		},
		{
			name:     "image without tag",
			image:    "quay.io/org/no-tag-service",
			wantName: "no-tag-service",
		},
		{
			name:     "image without registry",
			image:    "my-app-e2e:latest",
			wantName: "my-app",
		},
		{
			name:     "simple image name",
			image:    "service-test",
			wantName: "service",
		},
		{
			name:     "empty image",
			image:    "",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNameFromImage(tt.image)
			if got != tt.wantName {
				t.Errorf("extractNameFromImage(%q) = %q, want %q", tt.image, got, tt.wantName)
			}
		})
	}
}

func TestDeriveComponent(t *testing.T) {
	tests := []struct {
		name      string
		testImage string
		wantName  string
	}{
		{
			name:      "derive from test image",
			testImage: "quay.io/org/osd-example-operator-e2e:latest",
			wantName:  "osd-example-operator",
		},
		{
			name:      "fallback to unknown when no image",
			testImage: "",
			wantName:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup viper config
			if tt.testImage != "" {
				viper.Set(config.Tests.AdHocTestImages, []string{tt.testImage})
			} else {
				viper.Set(config.Tests.AdHocTestImages, []string{})
			}

			got := deriveComponent()
			if got != tt.wantName {
				t.Errorf("deriveComponent() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
