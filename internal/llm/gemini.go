package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  "gemini-2.5-pro",
	}, nil
}

func (g *GeminiClient) Analyze(ctx context.Context, userPrompt string, config *AnalysisConfig) (*AnalysisResult, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(userPrompt, genai.RoleUser),
	}

	var genConfig *genai.GenerateContentConfig
	if config != nil {
		genConfig = &genai.GenerateContentConfig{}

		if config.SystemInstruction != nil {
			genConfig.SystemInstruction = genai.NewContentFromText(*config.SystemInstruction, genai.RoleModel)
		}

		if config.Temperature != nil {
			genConfig.Temperature = config.Temperature
		}

		if config.TopP != nil {
			genConfig.TopP = config.TopP
		}

		if config.MaxTokens != nil {
			genConfig.MaxOutputTokens = int32(*config.MaxTokens)
		}

		if config.ResponseSchema != nil {
			genConfig.ResponseSchema = config.ResponseSchema
			genConfig.ResponseMIMEType = "application/json"
		}
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
		Content: content,
	}, nil
}
