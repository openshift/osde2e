package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"

	"github.com/openshift/osde2e/internal/llm/tools"
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

func (g *GeminiClient) Analyze(ctx context.Context, userPrompt string, config *AnalysisConfig, toolRegistry *tools.Registry) (*AnalysisResult, error) {
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

		if toolRegistry != nil {
			genConfig.Tools = toolRegistry.GetTools()
		}
	}

	return g.handleConversationWithTools(ctx, contents, genConfig, toolRegistry)
}

func (g *GeminiClient) handleConversationWithTools(ctx context.Context, contents []*genai.Content, genConfig *genai.GenerateContentConfig, toolRegistry *tools.Registry) (*AnalysisResult, error) {
	const maxIterations = 5

	for i := range maxIterations {
		resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, genConfig)
		if err != nil {
			return nil, fmt.Errorf("gemini API error: %w", err)
		}

		candidate, err := g.extractCandidate(resp)
		if err != nil {
			return nil, err
		}

		textContent, functionCalls := g.processCandidateParts(candidate)

		// If no function calls, we're done
		if len(functionCalls) == 0 {
			return &AnalysisResult{Content: textContent}, nil
		}

		// Process function calls and continue conversation
		contents, err = g.processFunctionCalls(ctx, contents, functionCalls, toolRegistry)
		if err != nil {
			return nil, err
		}

		// Return partial result if we hit max iterations
		if i == maxIterations-1 {
			return &AnalysisResult{Content: textContent}, nil
		}
	}

	return nil, fmt.Errorf("max iterations reached without final response")
}

func (g *GeminiClient) extractCandidate(resp *genai.GenerateContentResponse) (*genai.Candidate, error) {
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates from gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in gemini response")
	}

	return candidate, nil
}

func (g *GeminiClient) processCandidateParts(candidate *genai.Candidate) (string, []*genai.FunctionCall) {
	var textContent string
	var functionCalls []*genai.FunctionCall

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			textContent += part.Text
		}
		if part.FunctionCall != nil {
			functionCalls = append(functionCalls, part.FunctionCall)
		}
	}

	return textContent, functionCalls
}

func (g *GeminiClient) processFunctionCalls(ctx context.Context, contents []*genai.Content, functionCalls []*genai.FunctionCall, toolRegistry *tools.Registry) ([]*genai.Content, error) {
	for _, functionCall := range functionCalls {
		// Add the function call to conversation history
		contents = append(contents, genai.NewContentFromParts([]*genai.Part{{FunctionCall: functionCall}}, genai.RoleModel))

		// Execute the tool and get the result
		toolResult, err := toolRegistry.HandleToolCall(ctx, functionCall)
		if err != nil {
			return nil, fmt.Errorf("failed to handle tool call: %w", err)
		}

		// Add the tool result to conversation history
		contents = append(contents, toolResult)
	}

	return contents, nil
}
