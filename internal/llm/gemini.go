package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// GeminiClient implements LLMClient for Google's Gemini API
type GeminiClient struct {
	client *genai.Client
	model  string
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  "gemini-pro",
	}, nil
}

// Analyze sends a prompt to Gemini with optional configuration
func (g *GeminiClient) Analyze(ctx context.Context, userPrompt string, config ...*AnalysisConfig) (*AnalysisResult, error) {
	// Create content from user prompt
	contents := []*genai.Content{
		genai.NewContentFromText(userPrompt, genai.RoleUser),
	}

	// Build generation config
	var genConfig *genai.GenerateContentConfig
	if len(config) > 0 && config[0] != nil {
		cfg := config[0]

		genConfig = &genai.GenerateContentConfig{}

		// Set system instruction if provided
		if cfg.SystemInstruction != nil {
			genConfig.SystemInstruction = genai.NewContentFromText(*cfg.SystemInstruction, genai.RoleModel)
		}

		// Set generation parameters directly on the config
		if cfg.Temperature != nil {
			genConfig.Temperature = cfg.Temperature
		}

		if cfg.TopP != nil {
			genConfig.TopP = cfg.TopP
		}

		if cfg.MaxTokens != nil {
			genConfig.MaxOutputTokens = int32(*cfg.MaxTokens)
		}

		// TODO: Handle ResponseSchema and Tools when we need them
		// These require more complex type mapping from interface{} to genai types
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, genConfig)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates from gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in gemini response")
	}

	// Extract text from all parts
	var content string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			content += part.Text
		}
	}

	if content == "" {
		return nil, fmt.Errorf("no text content in gemini response")
	}

	return &AnalysisResult{
		Content:  content,
		Provider: "gemini",
		Model:    g.model,
	}, nil
}

// HealthCheck verifies connectivity to Gemini
func (g *GeminiClient) HealthCheck(ctx context.Context) error {
	_, err := g.Analyze(ctx, "Hello")
	return err
}

// Close cleans up the Gemini client
func (g *GeminiClient) Close() error {
	// The genai client doesn't have a Close method in this version
	return nil
}
