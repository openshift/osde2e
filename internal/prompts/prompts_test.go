package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptStore(t *testing.T) {
	store, err := NewPromptStore(DefaultTemplates())
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Greater(t, len(store.templates), 0, "Should have loaded some templates")
}

func TestGetTemplate(t *testing.T) {
	store, err := NewPromptStore(DefaultTemplates())
	require.NoError(t, err)

	template, err := store.GetTemplate("default")
	require.NoError(t, err)
	assert.NotNil(t, template)

	_, err = store.GetTemplate("non-existent")
	assert.Error(t, err)
}

func TestRenderPrompt(t *testing.T) {
	store, err := NewPromptStore(DefaultTemplates())
	require.NoError(t, err)

	variables := map[string]any{
		"ClusterID":      "test-cluster-id-123",
		"ClusterName":    "test-cluster",
		"Provider":       "aws",
		"Region":         "us-east-1",
		"Version":        "4.12.0",
		"FailureContext": "Installation failed with timeout error",
		"AnamolyLogs":    "ERROR: timeout waiting for machine-config-server",
		"Artifacts": []any{
			map[string]any{"Source": "/logs/installer.log", "LineCount": 522},
			map[string]any{"Source": "/logs/machine-config.log", "LineCount": 143},
		},
		"TestResults": map[string]any{
			"TotalTests":   5,
			"PassedTests":  3,
			"FailedTests":  2,
			"SkippedTests": 0,
			"ErrorTests":   0,
			"Duration":     "5m30s",
			"SuiteCount":   2,
		},
	}

	userPrompt, config, err := store.RenderPrompt("default", variables)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify template variables were substituted correctly
	assert.Contains(t, userPrompt, "aws")
	assert.Contains(t, userPrompt, "us-east-1")
	assert.Contains(t, userPrompt, "4.12.0")
	assert.Contains(t, userPrompt, "test-cluster")
	assert.Contains(t, userPrompt, "test-cluster-id-123")
	assert.Contains(t, userPrompt, "Installation failed with timeout error")
	assert.Contains(t, userPrompt, "/logs/installer.log (522 lines)")
	assert.Contains(t, userPrompt, "/logs/machine-config.log (143 lines)")

	// Verify test results section
	assert.Contains(t, userPrompt, "Total Tests: 5")
	assert.Contains(t, userPrompt, "Failed: 2")

	// Verify configuration
	assert.NotNil(t, config.SystemInstruction)
	assert.Contains(t, *config.SystemInstruction, "OpenShift administrator")
	assert.NotNil(t, config.Temperature)
	assert.NotNil(t, config.TopP)
	assert.NotNil(t, config.MaxTokens)
}
