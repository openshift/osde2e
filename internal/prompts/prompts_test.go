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

	templates := store.ListTemplates()
	assert.Greater(t, len(templates), 0, "Should have loaded some templates")
}

func TestGetTemplate(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	template, err := store.GetTemplate("provisioning-default")
	require.NoError(t, err)
	assert.Equal(t, "provisioning-default", template.ID)
	assert.Equal(t, ProvisioningFailure, template.Category)

	_, err = store.GetTemplate("non-existent")
	assert.Error(t, err)
}

func TestGetTemplatesByCategory(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	templates, err := store.GetTemplatesByCategory(ProvisioningFailure)
	require.NoError(t, err)
	assert.Greater(t, len(templates), 0)

	for _, template := range templates {
		assert.Equal(t, ProvisioningFailure, template.Category)
	}
}

func TestGetDefaultTemplate(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	template, err := store.GetDefaultTemplate(ProvisioningFailure)
	require.NoError(t, err)
	assert.Equal(t, ProvisioningFailure, template.Category)
	assert.True(t, template.IsDefault)
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

	rendered, err := store.RenderPrompt("provisioning-default", variables)
	require.NoError(t, err)
	require.NotNil(t, rendered)

	assert.Contains(t, rendered.UserPrompt, "aws")
	assert.Contains(t, rendered.UserPrompt, "us-east-1")
	assert.Contains(t, rendered.UserPrompt, "4.12.0")
	assert.Contains(t, rendered.UserPrompt, "test-cluster")
	assert.Contains(t, rendered.UserPrompt, "installer.log")

	assert.NotEmpty(t, rendered.SystemPrompt)
	assert.Contains(t, rendered.SystemPrompt, "OpenShift cluster administrator")
}

func TestRenderPromptWithDefaults(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	variables := map[string]any{
		"Provider":    "aws",
		"Region":      "us-east-1",
		"Version":     "4.12.0",
		"ClusterName": "test-cluster",
		"Artifacts": []any{
			map[string]any{"Name": "installer.log", "Content": "ERROR: Installation failed"},
		},
	}

	rendered, err := store.RenderPrompt("provisioning-default", variables)
	require.NoError(t, err)

	assert.Contains(t, rendered.UserPrompt, "OVNKubernetes") // default NetworkType
	assert.Contains(t, rendered.UserPrompt, "3")             // default ComputeNodes
}

func TestRenderedPromptMethods(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	variables := map[string]any{
		"Provider":    "aws",
		"Region":      "us-east-1",
		"Version":     "4.12.0",
		"ClusterName": "test-cluster",
		"Artifacts": []any{
			map[string]any{"Name": "installer.log", "Content": "ERROR: Installation failed"},
		},
	}

	rendered, err := store.RenderPrompt("provisioning-default", variables)
	require.NoError(t, err)

	assert.Equal(t, 1000, rendered.GetMaxTokens())
	assert.Equal(t, 0.1, rendered.GetTemperature())
	assert.Equal(t, 0.9, rendered.GetTopP())
}

func TestRenderPromptValidation(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	variables := map[string]any{
		"Provider": "aws",
		// Missing other required variables
	}

	_, err = store.RenderPrompt("provisioning-default", variables)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable validation failed")
}
