package aws

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openshift/osde2e/internal/sanitizer"
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

func TestPrepareUploadBody_SanitizesSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	// Build synthetic secret-shaped values at runtime to avoid triggering secret scanners
	// while still exercising the sanitizer rules with correctly formatted patterns.
	awsKey := "AKIA" + "IOSFODNN7EXAMPLE"
	ghToken := "ghp_" + strings.Repeat("a1b2c3d4", 5)
	secretContent := "AWS_ACCESS_KEY_ID=" + awsKey + "\ntoken=" + ghToken + "\n"
	if err := os.WriteFile(logFile, []byte(secretContent), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := sanitizer.New(&sanitizer.Config{
		EnableAudit: false,
	})
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	uploader := &S3Uploader{sanitizer: s}
	body, size, err := uploader.prepareUploadBody(logFile, "test.log")
	if err != nil {
		t.Fatalf("prepareUploadBody failed: %v", err)
	}

	defer body.Close()
	content, _ := io.ReadAll(body)
	if size != int64(len(content)) {
		t.Errorf("size mismatch: reported %d, actual %d", size, len(content))
	}

	result := string(content)
	if strings.Contains(result, awsKey) {
		t.Error("AWS access key was not redacted")
	}
	if strings.Contains(result, ghToken) {
		t.Error("GitHub token was not redacted")
	}
	if !strings.Contains(result, "[AWS-ACCESS-KEY-REDACTED]") {
		t.Error("Expected AWS redaction marker not found")
	}
	if !strings.Contains(result, "[GITHUB-TOKEN-REDACTED]") {
		t.Error("Expected GitHub redaction marker not found")
	}
}

func TestPrepareUploadBody_FailOpenOnOversizedContent(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "big.log")
	awsKey := "AKIA" + "IOSFODNN7EXAMPLE"
	content := strings.Repeat(awsKey+"\n", 100)
	if err := os.WriteFile(logFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := sanitizer.New(&sanitizer.Config{
		EnableAudit:    false,
		MaxContentSize: 50, // Tiny limit to trigger failure
	})
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}
	uploader := &S3Uploader{sanitizer: s}

	body, size, prepErr := uploader.prepareUploadBody(logFile, "big.log")
	if prepErr != nil {
		t.Fatalf("prepareUploadBody should not error on sanitization failure: %v", prepErr)
	}
	if size != int64(len(content)) {
		t.Errorf("Expected raw content size %d, got %d", len(content), size)
	}

	// Fail-open: raw content must be returned unchanged
	defer body.Close()
	result, _ := io.ReadAll(body)
	if string(result) != content {
		t.Errorf("Expected raw content unchanged on fail-open, got different content")
	}
}

func TestPrepareUploadBody_NilSanitizer(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	content := "AKIA" + "IOSFODNN7EXAMPLE"
	if err := os.WriteFile(logFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	uploader := &S3Uploader{sanitizer: nil}
	body, size, err := uploader.prepareUploadBody(logFile, "test.log")
	if err != nil {
		t.Fatalf("prepareUploadBody failed: %v", err)
	}

	// Without sanitizer, raw content is returned unchanged
	defer body.Close()
	result, _ := io.ReadAll(body)
	if string(result) != content {
		t.Errorf("Expected raw content %q, got %q", content, string(result))
	}
	if size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), size)
	}
}

func TestPrepareUploadBody_LargeFileStreamed(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "large.log")
	// Write a file larger than maxSanitizableBytes (50MB) — use sparse write via truncate
	f, err := os.Create(logFile)
	if err != nil {
		t.Fatal(err)
	}
	size := maxSanitizableBytes + 1
	if err := f.Truncate(size); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	s, err := sanitizer.New(&sanitizer.Config{EnableAudit: false})
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}
	uploader := &S3Uploader{sanitizer: s}

	body, reportedSize, prepErr := uploader.prepareUploadBody(logFile, "large.log")
	if prepErr != nil {
		t.Fatalf("prepareUploadBody failed: %v", prepErr)
	}
	defer body.Close()

	if reportedSize != size {
		t.Errorf("Expected size %d, got %d", size, reportedSize)
	}
}
