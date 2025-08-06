package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
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
	assert.Equal(t, float32(0.1), rendered.GetTemperature())
	assert.Equal(t, float32(0.9), rendered.GetTopP())
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

func TestResponseSchema(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	// Test template with auto-generated response schema
	template, err := store.GetTemplate("test-default")
	require.NoError(t, err)
	assert.NotNil(t, template.ResponseSchema, "test-default template should have auto-generated response schema")

	variables := map[string]any{
		"TestSuite":      "e2e",
		"TestName":       "test-pod-creation",
		"ClusterVersion": "4.12.0",
		"Provider":       "aws",
		"Region":         "us-east-1",
		"Artifacts": []any{
			map[string]any{"Name": "test.log", "Content": "ERROR: Pod creation failed"},
		},
	}

	rendered, err := store.RenderPrompt("test-default", variables)
	require.NoError(t, err)

	// Test that response schema is accessible
	schema := rendered.GetResponseSchema()
	assert.NotNil(t, schema)

	// Test conversion to AnalysisConfig
	config := rendered.ToAnalysisConfig()
	require.NotNil(t, config)
	assert.NotNil(t, config.SystemInstruction)
	assert.NotNil(t, config.Temperature)
	assert.NotNil(t, config.TopP)
	assert.NotNil(t, config.MaxTokens)
	assert.NotNil(t, config.ResponseSchema)

	assert.Equal(t, rendered.SystemPrompt, *config.SystemInstruction)
	assert.Equal(t, rendered.GetTemperature(), *config.Temperature)
	assert.Equal(t, rendered.GetTopP(), *config.TopP)
	assert.Equal(t, rendered.GetMaxTokens(), *config.MaxTokens)
	assert.Equal(t, schema, config.ResponseSchema)
}

func TestGetBaseResponseSchema(t *testing.T) {
	schema := getBaseResponseSchema()
	require.NotNil(t, schema)

	// Check basic structure
	assert.Equal(t, genai.TypeObject, schema.Type)

	// Check required fields exist
	assert.Contains(t, schema.Properties, "root_cause")
	assert.Contains(t, schema.Properties, "confidence")
	assert.Contains(t, schema.Properties, "recommendations")
	assert.Contains(t, schema.Properties, "failure_category")

	// Check required array
	assert.ElementsMatch(t, []string{"root_cause", "confidence", "recommendations", "failure_category"}, schema.Required)
}

func TestGetResponseSchemaForTemplate(t *testing.T) {
	template := &PromptTemplate{
		Categories: []string{"category1", "category2", "category3"},
	}

	schema := getResponseSchemaForTemplate(template)
	require.NotNil(t, schema)

	failureCategory, ok := schema.Properties["failure_category"]
	require.True(t, ok)

	assert.ElementsMatch(t, []string{"category1", "category2", "category3"}, failureCategory.Enum)
}

func TestAutoGeneratedResponseSchema(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	// Test that all default templates have auto-generated schemas
	categories := []FailureCategory{
		TestFailure,
		ProvisioningFailure,
		InfrastructureFailure,
		CleanupFailure,
		UpgradeFailure,
	}

	for _, category := range categories {
		t.Run(string(category), func(t *testing.T) {
			template, err := store.GetDefaultTemplate(category)
			require.NoError(t, err)
			assert.NotNil(t, template.ResponseSchema, "Template should have auto-generated response schema")

			// Verify the schema has failure category enums from the template
			schema := template.ResponseSchema
			require.NotNil(t, schema)

			failureCategoryProp, ok := schema.Properties["failure_category"]
			require.True(t, ok)

			// Just verify that enums exist and are not empty
			assert.NotEmpty(t, failureCategoryProp.Enum, "Template should have failure category enums")
			assert.Contains(t, failureCategoryProp.Enum, "other", "All templates should include 'other' as a fallback category")
		})
	}
}

func TestTemplateFailureCategories(t *testing.T) {
	store, err := NewPromptStore()
	require.NoError(t, err)

	// Test that templates load their categories from YAML
	template, err := store.GetTemplate("test-default")
	require.NoError(t, err)

	assert.NotEmpty(t, template.Categories, "Template should have categories loaded from YAML")
	assert.Contains(t, template.Categories, "other", "Template should include 'other' category")
}
