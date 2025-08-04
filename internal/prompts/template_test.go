package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptTemplateValidate(t *testing.T) {
	tests := []struct {
		name        string
		template    PromptTemplate
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid template",
			template: PromptTemplate{
				ID:           "test-template",
				Name:         "Test Template",
				Category:     ProvisioningFailure,
				SystemPrompt: "You are a test assistant.",
				UserPrompt:   "Analyze this: {{.Input}}",
			},
			expectError: false,
		},
		{
			name: "missing ID",
			template: PromptTemplate{
				Name:         "Test Template",
				Category:     ProvisioningFailure,
				SystemPrompt: "You are a test assistant.",
				UserPrompt:   "Analyze this: {{.Input}}",
			},
			expectError: true,
			errorMsg:    "template ID is required",
		},
		{
			name: "invalid category",
			template: PromptTemplate{
				ID:           "test-template",
				Name:         "Test Template",
				Category:     FailureCategory("invalid"),
				SystemPrompt: "You are a test assistant.",
				UserPrompt:   "Analyze this: {{.Input}}",
			},
			expectError: true,
			errorMsg:    "invalid category: invalid",
		},
		{
			name: "invalid template syntax",
			template: PromptTemplate{
				ID:           "test-template",
				Name:         "Test Template",
				Category:     ProvisioningFailure,
				SystemPrompt: "You are a test assistant.",
				UserPrompt:   "Analyze this: {{.Input", // Invalid template syntax
			},
			expectError: true,
			errorMsg:    "template compilation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromptTemplateRender(t *testing.T) {
	template := PromptTemplate{
		ID:           "test-template",
		Name:         "Test Template",
		Category:     ProvisioningFailure,
		SystemPrompt: "You are analyzing {{.Provider}} failures.",
		UserPrompt:   "Analyze cluster {{.ClusterName}} in {{.Region}}.",
	}

	variables := map[string]interface{}{
		"Provider":    "aws",
		"ClusterName": "test-cluster",
		"Region":      "us-east-1",
	}

	// Test system prompt rendering
	systemPrompt, err := template.RenderSystemPrompt(variables)
	require.NoError(t, err)
	assert.Equal(t, "You are analyzing aws failures.", systemPrompt)

	// Test user prompt rendering
	userPrompt, err := template.RenderUserPrompt(variables)
	require.NoError(t, err)
	assert.Equal(t, "Analyze cluster test-cluster in us-east-1.", userPrompt)
}

func TestPromptTemplateValidateVariables(t *testing.T) {
	template := PromptTemplate{
		Variables: []TemplateVar{
			{
				Name:     "Provider",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "Region",
				Type:     "string",
				Required: false,
				Default:  "us-east-1",
			},
		},
	}

	// Test with all required variables
	variables := map[string]interface{}{
		"Provider": "aws",
		"Region":   "us-west-2",
	}
	err := template.ValidateVariables(variables)
	assert.NoError(t, err)

	// Test with missing required variable
	variables = map[string]interface{}{
		"Region": "us-west-2",
	}
	err = template.ValidateVariables(variables)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required variable Provider is missing")

	// Test with missing optional variable (should be OK)
	variables = map[string]interface{}{
		"Provider": "aws",
	}
	err = template.ValidateVariables(variables)
	assert.NoError(t, err)
}

func TestPromptTemplateGetVariableDefaults(t *testing.T) {
	template := PromptTemplate{
		Variables: []TemplateVar{
			{
				Name:     "Provider",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "Region",
				Type:     "string",
				Required: false,
				Default:  "us-east-1",
			},
			{
				Name:     "NodeCount",
				Type:     "int",
				Required: false,
				Default:  3,
			},
		},
	}

	defaults := template.GetVariableDefaults()
	expected := map[string]interface{}{
		"Region":    "us-east-1",
		"NodeCount": 3,
	}
	assert.Equal(t, expected, defaults)
}
