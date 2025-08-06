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

	systemPrompt, userPrompt, config, err := store.RenderPrompt("provisioning-default", variables)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Contains(t, userPrompt, "aws")
	assert.Contains(t, userPrompt, "us-east-1")
	assert.Contains(t, userPrompt, "4.12.0")
	assert.Contains(t, userPrompt, "test-cluster")
	assert.Contains(t, userPrompt, "installer.log")

	assert.NotEmpty(t, systemPrompt)
	assert.Contains(t, systemPrompt, "OpenShift cluster administrator")

	assert.NotNil(t, config.SystemInstruction)
	assert.NotNil(t, config.Temperature)
	assert.NotNil(t, config.TopP)
	assert.NotNil(t, config.MaxTokens)
}

func TestTemplateDefaults(t *testing.T) {
	template := &PromptTemplate{}

	assert.Equal(t, 1000, template.getMaxTokens())
	assert.Equal(t, float32(0.1), template.getTemperature())
	assert.Equal(t, float32(0.9), template.getTopP())

	template.MaxTokens = 2000
	template.Temperature = 0.5
	template.TopP = 0.8

	assert.Equal(t, 2000, template.getMaxTokens())
	assert.Equal(t, float32(0.5), template.getTemperature())
	assert.Equal(t, float32(0.8), template.getTopP())
}
