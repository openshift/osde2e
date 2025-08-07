package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"

	"github.com/openshift/osde2e/internal/tools"
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

		if config.Tools != nil {
			genConfig.Tools = config.Tools
		}
	}

	// Iteratively handle the conversation until no more function calls
	maxIterations := 5 // Prevent infinite loops
	for iteration := 0; iteration < maxIterations; iteration++ {
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

		var hasFunctionCalls bool
		var finalContent string

		// Process all parts in this response
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				finalContent += part.Text
			}
			if part.FunctionCall != nil {
				hasFunctionCalls = true
				// Add the function call to conversation history
				contents = append(contents, genai.NewContentFromParts([]*genai.Part{{FunctionCall: part.FunctionCall}}, genai.RoleModel))

				// Handle the tool call and get the result
				toolResult, err := tools.HandleToolCall(part.FunctionCall)
				if err != nil {
					return nil, fmt.Errorf("failed to handle tool call: %w", err)
				}

				// Add the tool result to conversation history
				contents = append(contents, toolResult)
			}
		}

		// If no function calls in this iteration, we're done
		if !hasFunctionCalls {
			return &AnalysisResult{
				Content: finalContent,
			}, nil
		}

		// If we reach max iterations, return what we have
		if iteration == maxIterations-1 {
			return &AnalysisResult{
				Content: finalContent,
			}, nil
		}
	}

	return nil, fmt.Errorf("max iterations reached without final response")
}
