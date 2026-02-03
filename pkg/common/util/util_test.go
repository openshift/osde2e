package util

import "testing"

func TestContainsErrorMarker(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "contains ERROR uppercase",
			line:     "ERROR: connection failed",
			expected: true,
		},
		{
			name:     "contains Error mixed case",
			line:     "Error reading file",
			expected: true,
		},
		{
			name:     "contains error lowercase",
			line:     "error: invalid input",
			expected: true,
		},
		{
			name:     "contains error in middle of word",
			line:     "generator failed",
			expected: false, // "generator" doesn't contain "error", "Error", or "ERROR"
		},
		{
			name:     "contains error as substring",
			line:     "this is an error in the system",
			expected: true,
		},
		{
			name:     "no error marker",
			line:     "successful operation",
			expected: false,
		},
		{
			name:     "empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "Error with colon",
			line:     "Error: connection timeout",
			expected: true,
		},
		{
			name:     "error with colon",
			line:     "2026/01/19 09:39:15 Unable to find image. error: failed to find version",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsErrorMarker(tt.line)
			if result != tt.expected {
				t.Errorf("ContainsErrorMarker(%q) = %v, expected %v", tt.line, result, tt.expected)
			}
		})
	}
}

func TestContainsFailureMarker(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "contains [FAILED]",
			line:     "[FAILED] test description",
			expected: true,
		},
		{
			name:     "contains bullet [FAILED]",
			line:     "• [FAILED] test description",
			expected: true,
		},
		{
			name:     "no failure marker",
			line:     "test passed successfully",
			expected: false,
		},
		{
			name:     "contains PASSED not FAILED",
			line:     "[PASSED] test description",
			expected: false,
		},
		{
			name:     "empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "failed lowercase",
			line:     "test failed",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsFailureMarker(tt.line)
			if result != tt.expected {
				t.Errorf("ContainsFailureMarker(%q) = %v, expected %v", tt.line, result, tt.expected)
			}
		})
	}
}
