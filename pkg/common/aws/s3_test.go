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
			wantContains: []string{"test-results", "osd-example-operator", "12345"},
		},
		{
			name:         "empty category",
			component:    "my-service",
			jobID:        "67890",
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
			wantContains: []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestContentTypeForFile(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"executor.log", "text/plain; charset=utf-8"},
		{"config.yaml", "text/plain; charset=utf-8"},
		{"config.yml", "text/plain; charset=utf-8"},
		{"data.csv", "text/csv; charset=utf-8"},
		{"notes.txt", "text/plain; charset=utf-8"},
		{"junit.xml", "text/xml; charset=utf-8"},
		{"report.json", "application/json"},
		{"report.html", "text/html; charset=utf-8"},
		{"screenshot.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"animation.gif", "image/gif"},
		{"EXECUTOR.LOG", "text/plain; charset=utf-8"},
		{"binary.bin", "application/octet-stream"},
		{"noext", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := contentTypeForFile(tt.filename)
			if got != tt.want {
				t.Errorf("contentTypeForFile(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestShouldUploadFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"junit.xml", true},
		{"test_output.log", true},
		{"summary.yaml", true},
		{"config.yml", true},
		{"report.json", true},
		{"screenshot.png", true},
		{"graph.jpg", true},
		{"photo.jpeg", true},
		{"animation.gif", true},
		{"data.csv", true},
		{"notes.txt", true},
		{"report.html", true},
		{"test_output", true},
		{"TEST_OUTPUT", true},
		{"summary", true},
		{"osde2e-full", true},
		{"cluster-state", true},
		{"binary.exe", false},
		{"archive.tar.gz", false},
		{"data.bin", false},
		{"library.so", false},
		{"randomfile", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := shouldUploadFile(tt.filename); got != tt.want {
				t.Errorf("shouldUploadFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
