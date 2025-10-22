package sanitizer

import (
	"strings"
	"testing"
)

func TestDataSanitizer_BasicFunctionality(t *testing.T) {
	// Create sanitizer with default configuration
	config := &Config{
		EnableAudit:    false,       // Disable audit logging for tests
		MaxContentSize: 1024 * 1024, // 1MB
		StrictMode:     false,
	}

	sanitizer, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	testCases := []struct {
		name             string
		input            string
		shouldContain    []string // Content that should be present after sanitization
		shouldNotContain []string // Content that should not be present after sanitization
		expectMatches    bool     // Whether matches are expected to be found
	}{
		{
			name:             "AWS Access Key",
			input:            "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			shouldContain:    []string{"[AWS-ACCESS-KEY-REDACTED]"},
			shouldNotContain: []string{"AKIAIOSFODNN7EXAMPLE"},
			expectMatches:    true,
		},
		{
			name:             "GitHub Token",
			input:            "export GITHUB_TOKEN=ghp_1234567890abcdef1234567890abcdef12",
			shouldContain:    []string{"[GITHUB-TOKEN-REDACTED]"},
			shouldNotContain: []string{"ghp_1234567890abcdef1234567890abcdef12"},
			expectMatches:    true,
		},
		{
			name:             "JWT Token",
			input:            "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			shouldContain:    []string{"[JWT-TOKEN-REDACTED]"},
			shouldNotContain: []string{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
			expectMatches:    true,
		},
		{
			name:             "Multiple Tokens",
			input:            "Token: ghp_abcdef123456789012345678901234567890, Key: AKIAIOSFODNN7EXAMPLE",
			shouldContain:    []string{"[GITHUB-TOKEN-REDACTED]", "[AWS-ACCESS-KEY-REDACTED]"},
			shouldNotContain: []string{"ghp_abcdef123456789012345678901234567890", "AKIAIOSFODNN7EXAMPLE"},
			expectMatches:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := sanitizer.SanitizeText(tc.input, "test-source")
			if err != nil {
				t.Fatalf("SanitizeText failed: %v", err)
			}

			// Check if matches were found as expected
			if tc.expectMatches && result.MatchesFound == 0 {
				t.Errorf("Expected to find matches, but found none")
			}
			if !tc.expectMatches && result.MatchesFound > 0 {
				t.Errorf("Expected no matches, but found %d", result.MatchesFound)
			}

			// Check content that should be present
			for _, expected := range tc.shouldContain {
				if !strings.Contains(result.Content, expected) {
					t.Errorf("Expected sanitized content to contain '%s', but it didn't. Content: %s", expected, result.Content)
				}
			}

			// Check content that should not be present
			for _, notExpected := range tc.shouldNotContain {
				if strings.Contains(result.Content, notExpected) {
					t.Errorf("Expected sanitized content to NOT contain '%s', but it did. Content: %s", notExpected, result.Content)
				}
			}

			// Validate metadata
			if result.Source != "test-source" {
				t.Errorf("Expected source to be 'test-source', got '%s'", result.Source)
			}

			if result.Timestamp.IsZero() {
				t.Errorf("Expected timestamp to be set")
			}
		})
	}
}

func TestDataSanitizer_Configuration(t *testing.T) {
	// Test default configuration
	sanitizer, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create sanitizer with default config: %v", err)
	}

	result, err := sanitizer.SanitizeText("AKIAIOSFODNN7EXAMPLE", "test")
	if err != nil {
		t.Fatalf("SanitizeText failed: %v", err)
	}

	if result.MatchesFound == 0 {
		t.Errorf("Expected matches with default rules")
	}
}

func TestDataSanitizer_ErrorHandling(t *testing.T) {
	config := &Config{
		EnableAudit:    false,
		MaxContentSize: 100, // Very small limit
		StrictMode:     true,
	}

	sanitizer, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	// Test file size limit
	largeContent := strings.Repeat("a", 200) // Exceeds 100 byte limit
	_, err = sanitizer.SanitizeText(largeContent, "test")
	if err == nil {
		t.Errorf("Expected error for content exceeding size limit, but got none")
	}
}

func TestDataSanitizer_AuditCleanup(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableAudit:        true,
		AuditLogDir:        tmpDir,
		AuditRetentionDays: 1, // 1 day retention for testing
		StrictMode:         false,
	}

	sanitizer, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	// Test cleanup functionality
	err = sanitizer.CleanupAuditLogs()
	if err != nil {
		t.Errorf("CleanupAuditLogs failed: %v", err)
	}
}

func TestDataSanitizer_NewTokenTypes(t *testing.T) {
	config := &Config{
		EnableAudit: false,
		StrictMode:  false,
	}

	sanitizer, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create sanitizer: %v", err)
	}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Docker Auth Token",
			input:    "docker_auth=dGVzdDp0ZXN0MTIzNDU2Nzg5MA==",
			expected: "docker_auth=[DOCKER-AUTH-REDACTED]",
		},
		{
			name:     "Generic Access Token",
			input:    "access_token=abc123def456ghi789jkl012mno345pqr678",
			expected: "access_token=[TOKEN-REDACTED]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := sanitizer.SanitizeText(tc.input, "test-source")
			if err != nil {
				t.Fatalf("SanitizeText failed: %v", err)
			}

			if result.MatchesFound == 0 {
				t.Errorf("Expected matches to be found for %s", tc.name)
			}

			if !strings.Contains(result.Content, tc.expected) {
				t.Errorf("Expected content to contain %s, got %s", tc.expected, result.Content)
			}
		})
	}
}

func BenchmarkDataSanitizer(b *testing.B) {
	config := &Config{
		EnableAudit: false, // Disable audit logging for clean performance testing
		StrictMode:  false,
	}

	sanitizer, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create sanitizer: %v", err)
	}

	// Test content containing multiple types of sensitive data
	testContent := `
		Log entry with multiple sensitive data:
		AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
		GITHUB_TOKEN=ghp_1234567890abcdef1234567890abcdef12
		Email: admin@company.com
		JWT: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
		Database: postgresql://user:password123@localhost:5432/db
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sanitizer.SanitizeText(testContent, "benchmark")
		if err != nil {
			b.Fatalf("SanitizeText failed: %v", err)
		}
	}
}
