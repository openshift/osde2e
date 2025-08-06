// Package prompts provides a structured prompt template management system
// for LLM-based failure analysis in osde2e.
package prompts

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"google.golang.org/genai"
	"gopkg.in/yaml.v3"

	"github.com/openshift/osde2e/internal/llm"
)

// Embed the default prompt templates
//
//go:embed templates/*.yaml
var defaultTemplates embed.FS

// FailureCategory represents different types of failures that can be analyzed
type FailureCategory string

const (
	ProvisioningFailure   FailureCategory = "provisioning"
	InfrastructureFailure FailureCategory = "infrastructure"
	TestFailure           FailureCategory = "test"
	CleanupFailure        FailureCategory = "cleanup"
	UpgradeFailure        FailureCategory = "upgrade"
)

// getBaseResponseSchema returns the standard response schema used across all failure analysis templates
func getBaseResponseSchema() *genai.Schema {
	minItems := int64(2)
	maxItems := int64(3)
	minimum := 0.0
	maximum := 1.0

	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"root_cause": {
				Type:        genai.TypeString,
				Description: "Specific description of what failed",
			},
			"confidence": {
				Type:        genai.TypeNumber,
				Description: "Confidence score between 0.0 and 1.0",
				Minimum:     &minimum,
				Maximum:     &maximum,
			},
			"recommendations": {
				Type:        genai.TypeArray,
				Description: "List of 2-3 specific, actionable recommendations",
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
				MinItems: &minItems,
				MaxItems: &maxItems,
			},
			"failure_category": {
				Type:        genai.TypeString,
				Description: "Category of the failure",
			},
		},
		Required: []string{"root_cause", "confidence", "recommendations", "failure_category"},
	}
}

// getResponseSchemaForTemplate returns a complete response schema for a specific template
func getResponseSchemaForTemplate(template *PromptTemplate) *genai.Schema {
	schema := getBaseResponseSchema()

	// Add the template-specific enum values if provided
	if len(template.Categories) > 0 {
		if failureCategory, ok := schema.Properties["failure_category"]; ok {
			failureCategory.Enum = template.Categories
		}
	}

	return schema
}

// PromptTemplate represents a structured prompt template for LLM analysis
type PromptTemplate struct {
	ID          string          `yaml:"id"`
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Category    FailureCategory `yaml:"category"`
	IsDefault   bool            `yaml:"is_default"`

	// Prompt content
	SystemPrompt string        `yaml:"system_prompt"`
	UserPrompt   string        `yaml:"user_prompt"`
	Variables    []TemplateVar `yaml:"variables"`

	// LLM configuration
	MaxTokens      int           `yaml:"max_tokens"`
	Temperature    float32       `yaml:"temperature"`
	TopP           float32       `yaml:"top_p"`
	ResponseSchema *genai.Schema `yaml:"response_schema,omitempty"`
	Categories     []string      `yaml:"categories,omitempty"`

	// compiled templates (not serialized)
	systemTemplate *template.Template `yaml:"-"`
	userTemplate   *template.Template `yaml:"-"`
}

// TemplateVar defines a variable that can be substituted in the prompt
type TemplateVar struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     any    `yaml:"default"`
	Validation  string `yaml:"validation"`
}

// PromptStore manages prompt templates
type PromptStore struct {
	templates map[string]*PromptTemplate
}

// NewPromptStore creates a new prompt store with default templates
func NewPromptStore() (*PromptStore, error) {
	store := &PromptStore{
		templates: make(map[string]*PromptTemplate),
	}

	// Load default templates
	templatesFS, err := fs.Sub(defaultTemplates, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to access default templates: %w", err)
	}

	if err := store.LoadTemplates(templatesFS); err != nil {
		return nil, fmt.Errorf("failed to load default templates: %w", err)
	}

	return store, nil
}

// LoadTemplates loads all prompt templates from the given filesystem
func (ps *PromptStore) LoadTemplates(filesystem fs.FS) error {
	return fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, err := fs.ReadFile(filesystem, path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", path, err)
		}

		var template PromptTemplate
		if err := yaml.Unmarshal(data, &template); err != nil {
			return fmt.Errorf("failed to parse template file %s: %w", path, err)
		}

		// Set template ID from filename if not specified
		if template.ID == "" {
			template.ID = strings.TrimSuffix(filepath.Base(path), ".yaml")
		}

		// Auto-generate response schema if not provided
		if template.ResponseSchema == nil && template.Category != "" {
			template.ResponseSchema = getResponseSchemaForTemplate(&template)
		}

		if err := template.Validate(); err != nil {
			return fmt.Errorf("invalid template %s: %w", template.ID, err)
		}

		ps.templates[template.ID] = &template
		return nil
	})
}

// GetTemplate retrieves a template by ID
func (ps *PromptStore) GetTemplate(id string) (*PromptTemplate, error) {
	template, exists := ps.templates[id]
	if !exists {
		return nil, fmt.Errorf("template %s not found", id)
	}
	return template, nil
}

// GetTemplatesByCategory returns all templates for a specific failure category
func (ps *PromptStore) GetTemplatesByCategory(category FailureCategory) ([]*PromptTemplate, error) {
	var templates []*PromptTemplate
	for _, template := range ps.templates {
		if template.Category == category {
			templates = append(templates, template)
		}
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found for category %s", category)
	}

	return templates, nil
}

// GetDefaultTemplate returns the default template for a category
func (ps *PromptStore) GetDefaultTemplate(category FailureCategory) (*PromptTemplate, error) {
	templates, err := ps.GetTemplatesByCategory(category)
	if err != nil {
		return nil, err
	}

	// Find template marked as default
	for _, template := range templates {
		if template.IsDefault {
			return template, nil
		}
	}

	// Return first template if no default is marked
	return templates[0], nil
}

// ListTemplates returns all available template IDs
func (ps *PromptStore) ListTemplates() []string {
	ids := make([]string, 0, len(ps.templates))
	for id := range ps.templates {
		ids = append(ids, id)
	}
	return ids
}

// RenderPrompt renders a template with the given variables
func (ps *PromptStore) RenderPrompt(templateID string, variables map[string]any) (*RenderedPrompt, error) {
	template, err := ps.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Merge with defaults
	mergedVars := template.GetVariableDefaults()
	for k, v := range variables {
		mergedVars[k] = v
	}

	// Validate required variables
	if err := template.ValidateVariables(mergedVars); err != nil {
		return nil, fmt.Errorf("variable validation failed: %w", err)
	}

	// Render prompts
	systemPrompt, err := template.RenderSystemPrompt(mergedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to render system prompt: %w", err)
	}

	userPrompt, err := template.RenderUserPrompt(mergedVars)
	if err != nil {
		return nil, fmt.Errorf("failed to render user prompt: %w", err)
	}

	return &RenderedPrompt{
		Template:     template,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Variables:    mergedVars,
	}, nil
}

// RenderedPrompt represents a fully rendered prompt ready for LLM consumption
type RenderedPrompt struct {
	Template     *PromptTemplate
	SystemPrompt string
	UserPrompt   string
	Variables    map[string]any
}

// GetMaxTokens returns the maximum tokens setting for the template
func (rp *RenderedPrompt) GetMaxTokens() int {
	if rp.Template.MaxTokens > 0 {
		return rp.Template.MaxTokens
	}
	return 4000 // default
}

// GetTemperature returns the temperature setting for the template
func (rp *RenderedPrompt) GetTemperature() float32 {
	if rp.Template.Temperature > 0 {
		return rp.Template.Temperature
	}
	return 0.1 // default for analysis tasks
}

// GetTopP returns the top-p setting for the template
func (rp *RenderedPrompt) GetTopP() float32 {
	if rp.Template.TopP > 0 {
		return rp.Template.TopP
	}
	return 0.9 // default
}

// GetResponseSchema returns the response schema for the template
func (rp *RenderedPrompt) GetResponseSchema() *genai.Schema {
	return rp.Template.ResponseSchema
}

// ToAnalysisConfig converts the rendered prompt to an LLM AnalysisConfig
func (rp *RenderedPrompt) ToAnalysisConfig() *llm.AnalysisConfig {
	config := &llm.AnalysisConfig{
		SystemInstruction: llm.StringPtr(rp.SystemPrompt),
		Temperature:       llm.Float32Ptr(rp.GetTemperature()),
		TopP:              llm.Float32Ptr(rp.GetTopP()),
		MaxTokens:         llm.IntPtr(rp.GetMaxTokens()),
		ResponseSchema:    rp.GetResponseSchema(),
	}
	return config
}
