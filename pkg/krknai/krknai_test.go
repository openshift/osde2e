package krknai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

func TestDetectContainerRuntime(t *testing.T) {
	runtime, err := detectContainerRuntime()
	// This test will pass if either podman or docker is installed
	// If neither is installed, it should return an error
	if err != nil {
		t.Logf("No container runtime found (expected in CI without containers): %v", err)
		return
	}

	if runtime == "" {
		t.Error("detectContainerRuntime() returned empty string without error")
	}

	t.Logf("Detected container runtime: %s", runtime)
}

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		noCreds string // redacted URL must not contain this
		has     string // redacted URL must still contain this
	}{
		{"userinfo removed", "https://user:secret@example.com/health", "secret", "example.com"},
		{"query removed", "https://example.com/health?token=abc", "token=abc", "example.com"},
		{"userinfo and query removed", "https://u:p@host/path?k=v", "k=v", "host"},
		{"plain URL unchanged host/path", "https://example.com/health", "", "example.com"},
		{"invalid URL redacted", "://bad", "", "<redacted>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactURL(tt.rawURL)
			if tt.noCreds != "" && strings.Contains(got, tt.noCreds) {
				t.Errorf("redactURL(%q) = %q, must not contain %q", tt.rawURL, got, tt.noCreds)
			}
			if tt.has != "" && !strings.Contains(got, tt.has) {
				t.Errorf("redactURL(%q) = %q, must contain %q", tt.rawURL, got, tt.has)
			}
		})
	}
}

func TestValidateHealthCheckURLsReachable(t *testing.T) {
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer okServer.Close()
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }))
	defer failServer.Close()

	tests := []struct {
		name    string
		apps    []map[string]interface{}
		wantErr bool
	}{
		{
			name: "all reachable 2xx",
			apps: []map[string]interface{}{
				{"name": "a", "url": okServer.URL},
			},
			wantErr: false,
		},
		{
			name: "non-2xx returns error",
			apps: []map[string]interface{}{
				{"name": "b", "url": failServer.URL},
			},
			wantErr: true,
		},
		{
			name: "unreachable URL returns error",
			apps: []map[string]interface{}{
				{"name": "c", "url": "http://127.0.0.1:0/"},
			},
			wantErr: true,
		},
		{
			name:    "empty list succeeds",
			apps:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHealthCheckURLsReachable(context.Background(), tt.apps)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParseHealthCheckEndpoints(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantCount int
		wantNames []string
	}{
		{
			name:      "valid https endpoint",
			input:     "console=https://console.example.com/health",
			wantCount: 1,
			wantNames: []string{"console"},
		},
		{
			name:      "valid http endpoint",
			input:     "api=http://api.example.com/ready",
			wantCount: 1,
			wantNames: []string{"api"},
		},
		{
			name:      "multiple valid endpoints",
			input:     "console=https://console.example.com/health,api=https://api.example.com/ready",
			wantCount: 2,
			wantNames: []string{"console", "api"},
		},
		{
			name:    "missing scheme rejected",
			input:   "console=console.example.com/health",
			wantErr: true,
		},
		{
			name:    "missing host rejected",
			input:   "console=https:///health",
			wantErr: true,
		},
		{
			name:    "unsupported scheme rejected",
			input:   "console=ftp://files.example.com",
			wantErr: true,
		},
		{
			name:    "empty value rejected",
			input:   "console=",
			wantErr: true,
		},
		{
			name:    "empty name rejected",
			input:   "=https://example.com/health",
			wantErr: true,
		},
		{
			name:    "whitespace-only name rejected",
			input:   "  =https://example.com/health",
			wantErr: true,
		},
		{
			name:    "missing equals rejected",
			input:   "just-a-string",
			wantErr: true,
		},
		{
			name:    "mix of valid and invalid returns error on first invalid",
			input:   "good=https://ok.com/health,bad=not-a-url,also-good=http://fine.com/ready",
			wantErr: true,
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apps, err := parseHealthCheckEndpoints(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, apps, tt.wantCount)
			for i, name := range tt.wantNames {
				assert.Equal(t, name, apps[i]["name"])
			}
		})
	}
}

func TestMarkdownToHTML(t *testing.T) {
	input := "# Krkn-AI Chaos Test Report\n\n## Executive Summary\nCluster shows **moderate** resilience.\n\n| Metric | Value |\n|--------|-------|\n| Total | 5 |\n"

	html, err := markdownToHTML(input)
	require.NoError(t, err)

	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "<h1")
	assert.Contains(t, html, "<table>")
	assert.Contains(t, html, "<strong>moderate</strong>")
	assert.NotContains(t, html, "## Executive Summary")
}

func TestKrknAIViperConfig(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
		mode     string // "discover" or "run"
	}{
		// Discover mode specific fields
		{"Namespace", config.KrknAI.Namespace, "default", "discover"},
		{"PodLabel", config.KrknAI.PodLabel, "", "discover"},
		{"NodeLabel", config.KrknAI.NodeLabel, "kubernetes.io/hostname", "discover"},
		{"SkipPodName", config.KrknAI.SkipPodName, "", "discover"},

		// Run mode specific fields (FitnessQuery, Scenarios)
		{"FitnessQuery", config.KrknAI.FitnessQuery, "", "run"},
		{"Scenarios", config.KrknAI.Scenarios, "", "run"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := viper.GetString(tt.key)
			if value != tt.expected {
				t.Errorf("viper.GetString(%q) = %q, want %q (mode: %s)", tt.key, value, tt.expected, tt.mode)
			}
		})
	}
}
