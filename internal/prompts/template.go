package prompts

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"text/template"
)

// Validate checks if the template is valid
func (pt *PromptTemplate) Validate() error {
	if pt.ID == "" {
		return fmt.Errorf("template ID is required")
	}
	if pt.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if pt.Category == "" {
		return fmt.Errorf("template category is required")
	}
	if pt.SystemPrompt == "" {
		return fmt.Errorf("system prompt is required")
	}
	if pt.UserPrompt == "" {
		return fmt.Errorf("user prompt is required")
	}

	// Validate category
	validCategories := []FailureCategory{
		ProvisioningFailure, InfrastructureFailure, TestFailure, CleanupFailure, UpgradeFailure,
	}
	if !slices.Contains(validCategories, pt.Category) {
		return fmt.Errorf("invalid category: %s", pt.Category)
	}

	// Validate template syntax
	if err := pt.compileTemplates(); err != nil {
		return fmt.Errorf("template compilation failed: %w", err)
	}

	return nil
}

// compileTemplates compiles the Go templates for validation and later use
func (pt *PromptTemplate) compileTemplates() error {
	var err error

	// Compile system prompt template
	pt.systemTemplate, err = template.New("system").Parse(pt.SystemPrompt)
	if err != nil {
		return fmt.Errorf("failed to compile system prompt: %w", err)
	}

	// Compile user prompt template
	pt.userTemplate, err = template.New("user").Parse(pt.UserPrompt)
	if err != nil {
		return fmt.Errorf("failed to compile user prompt: %w", err)
	}

	return nil
}

// RenderSystemPrompt renders the system prompt with the given variables
func (pt *PromptTemplate) RenderSystemPrompt(variables map[string]interface{}) (string, error) {
	if pt.systemTemplate == nil {
		if err := pt.compileTemplates(); err != nil {
			return "", err
		}
	}

	var buf bytes.Buffer
	if err := pt.systemTemplate.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render system prompt: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// RenderUserPrompt renders the user prompt with the given variables
func (pt *PromptTemplate) RenderUserPrompt(variables map[string]interface{}) (string, error) {
	if pt.userTemplate == nil {
		if err := pt.compileTemplates(); err != nil {
			return "", err
		}
	}

	var buf bytes.Buffer
	if err := pt.userTemplate.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render user prompt: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// ValidateVariables validates that all required variables are provided
func (pt *PromptTemplate) ValidateVariables(variables map[string]interface{}) error {
	for _, templateVar := range pt.Variables {
		if templateVar.Required {
			if _, exists := variables[templateVar.Name]; !exists {
				return fmt.Errorf("required variable %s is missing", templateVar.Name)
			}
		}
	}
	return nil
}

// GetVariableDefaults returns a map of default values for template variables
func (pt *PromptTemplate) GetVariableDefaults() map[string]interface{} {
	defaults := make(map[string]interface{})
	for _, templateVar := range pt.Variables {
		if templateVar.Default != nil {
			defaults[templateVar.Name] = templateVar.Default
		}
	}
	return defaults
}
