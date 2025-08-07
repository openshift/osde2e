package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptStore(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Greater(t, len(store.templates), 0, "Should have loaded some templates")
}

func TestGetTemplate(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	template, err := store.GetTemplate("provisioning-default")
	require.NoError(t, err)
	assert.Equal(t, "provisioning-default", template.ID)

	_, err = store.GetTemplate("non-existent")
	assert.Error(t, err)
}

func TestRenderPrompt(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	variables := map[string]any{
		"Provider":    "aws",
		"Region":      "us-east-1",
		"Version":     "4.12.0",
		"ClusterName": "test-cluster",
		"FailureTime": "2024-01-01T12:00:00Z",
		"Duration":    "30m",
		"Artifacts": []any{
			map[string]any{"Name": "installer.log", "Content": "ERROR: Installation failed"},
		},
	}

	userPrompt, config, err := store.RenderPrompt("provisioning-default", variables)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Contains(t, userPrompt, "aws")
	assert.Contains(t, userPrompt, "us-east-1")
	assert.Contains(t, userPrompt, "4.12.0")
	assert.Contains(t, userPrompt, "test-cluster")
	assert.Contains(t, userPrompt, "installer.log")

	assert.NotNil(t, config.SystemInstruction)
	assert.Contains(t, *config.SystemInstruction, "OpenShift cluster administrator")

	assert.NotNil(t, config.SystemInstruction)
	assert.NotNil(t, config.Temperature)
	assert.NotNil(t, config.TopP)
	assert.NotNil(t, config.MaxTokens)
}
