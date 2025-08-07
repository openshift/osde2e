package prompts

import (
	"bytes"
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

//go:embed templates/*.yaml
var defaultTemplates embed.FS

// Package-wide defaults for all prompts
const (
	defaultMaxTokens   = 4000
	defaultTemperature = float32(0.1)
	defaultTopP        = float32(0.9)
)

type PromptTemplate struct {
	ID           string `yaml:"id"`
	SystemPrompt string `yaml:"system_prompt"`
	UserPrompt   string `yaml:"user_prompt"`
}

type PromptStore struct {
	templates map[string]*PromptTemplate
}

func NewPromptStore() (*PromptStore, error) {
	store := &PromptStore{
		templates: make(map[string]*PromptTemplate),
	}

	templatesFS, err := fs.Sub(defaultTemplates, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to access templates: %w", err)
	}

	return store, store.loadTemplates(templatesFS)
}

func (ps *PromptStore) loadTemplates(filesystem fs.FS) error {
	return fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}

		data, err := fs.ReadFile(filesystem, path)
		if err != nil {
			return err
		}

		var template PromptTemplate
		if err := yaml.Unmarshal(data, &template); err != nil {
			return err
		}

		if template.ID == "" {
			template.ID = strings.TrimSuffix(filepath.Base(path), ".yaml")
		}

		ps.templates[template.ID] = &template
		return nil
	})
}

func (ps *PromptStore) GetTemplate(id string) (*PromptTemplate, error) {
	template, exists := ps.templates[id]
	if !exists {
		return nil, fmt.Errorf("template %s not found", id)
	}
	return template, nil
}

func (ps *PromptStore) RenderPrompt(templateID string, variables map[string]any) (userPrompt string, config *llm.AnalysisConfig, err error) {
	template, err := ps.GetTemplate(templateID)
	if err != nil {
		return "", nil, err
	}

	systemPrompt, err := template.render(template.SystemPrompt, variables)
	if err != nil {
		return "", nil, fmt.Errorf("failed to render system prompt: %w", err)
	}

	userPrompt, err = template.render(template.UserPrompt, variables)
	if err != nil {
		return "", nil, fmt.Errorf("failed to render user prompt: %w", err)
	}

	config = &llm.AnalysisConfig{
		SystemInstruction: genai.Ptr(systemPrompt),
		Temperature:       genai.Ptr[float32](defaultTemperature),
		TopP:              genai.Ptr[float32](defaultTopP),
		MaxTokens:         genai.Ptr(defaultMaxTokens),
		ResponseSchema:    getAnalysisResponseSchema(),
	}

	return userPrompt, config, nil
}

func (pt *PromptTemplate) render(promptText string, variables map[string]any) (string, error) {
	tmpl, err := template.New("prompt").Parse(promptText)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// getAnalysisResponseSchema returns the standard response schema for failure analysis
func getAnalysisResponseSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"root_cause": {
				Type:        genai.TypeString,
				Description: "Specific description of what failed",
			},
			"confidence_score": {
				Type:        genai.TypeNumber,
				Description: "Confidence score between 0.0 and 1.0",
			},
			"recommendations": {
				Type:        genai.TypeArray,
				Description: "List of 2-3 specific, actionable recommendations",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"title": {
							Type:        genai.TypeString,
							Description: "Brief title for the recommendation",
						},
						"description": {
							Type:        genai.TypeString,
							Description: "Detailed description of the recommendation",
						},
					},
					Required: []string{"title", "description"},
				},
			},
		},
		Required: []string{"root_cause", "confidence_score", "recommendations"},
	}
}
