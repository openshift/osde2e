package llm

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/openshift/osde2e/internal/llm/tools"
)

const (
	DefaultModel  = "gemini-3.1-pro-preview"
	FallbackModel = "gemini-2.5-pro"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	return NewGeminiClientWithModel(ctx, apiKey, DefaultModel)
}

func NewGeminiClientWithModel(ctx context.Context, apiKey, model string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  model,
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
	var toolCalls []*genai.FunctionCall

	for range maxIterations {
		textContent, functionCalls, err := g.callLLM(ctx, contents, genConfig)
		if err != nil {
			return nil, err
		}

		toolCalls = append(toolCalls, functionCalls...)

		if len(functionCalls) == 0 {
			return &AnalysisResult{
				Content:   textContent,
				ToolCalls: toolCalls,
			}, nil
		}

		contents, err = g.processFunctionCalls(ctx, contents, functionCalls, toolRegistry)
		if err != nil {
			return nil, err
		}
	}

	// Loop exhausted: make one final call with tool calling disabled to force a text response
	contents = append(contents, genai.NewContentFromText(
		"You have used all available tool calls. Produce your final analysis based on the information gathered so far.",
		genai.RoleUser,
	))
	finalConfig := g.configWithToolsDisabled(genConfig)
	finalText, _, err := g.callLLM(ctx, contents, finalConfig)
	if err != nil {
		return &AnalysisResult{ToolCalls: toolCalls}, err
	}

	return &AnalysisResult{
		Content:   finalText,
		ToolCalls: toolCalls,
	}, nil
}

func (g *GeminiClient) callLLM(ctx context.Context, contents []*genai.Content, genConfig *genai.GenerateContentConfig) (string, []*genai.FunctionCall, error) {
	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, genConfig)
	if err != nil {
		return "", nil, fmt.Errorf("gemini API error: %w", err)
	}

	candidate, err := g.extractCandidate(resp)
	if err != nil {
		return "", nil, err
	}

	text, calls := g.processCandidateParts(candidate)
	return text, calls, nil
}

func (g *GeminiClient) configWithToolsDisabled(genConfig *genai.GenerateContentConfig) *genai.GenerateContentConfig {
	if genConfig == nil {
		return nil
	}
	cfg := *genConfig
	cfg.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingConfigModeNone,
		},
	}
	return &cfg
}

func (g *GeminiClient) extractCandidate(resp *genai.GenerateContentResponse) (*genai.Candidate, error) {
	if len(resp.Candidates) == 0 {
		return nil, ErrNoResponseCandidates
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, ErrNoContentInResponse
	}

	return candidate, nil
}

func (g *GeminiClient) processCandidateParts(candidate *genai.Candidate) (string, []*genai.FunctionCall) {
	var text strings.Builder
	var functionCalls []*genai.FunctionCall

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			text.WriteString(part.Text)
		}
		if part.FunctionCall != nil {
			functionCalls = append(functionCalls, part.FunctionCall)
		}
	}

	return text.String(), functionCalls
}

func (g *GeminiClient) processFunctionCalls(ctx context.Context, contents []*genai.Content, functionCalls []*genai.FunctionCall, toolRegistry *tools.Registry) ([]*genai.Content, error) {
	for _, functionCall := range functionCalls {
		// Add the function call to conversation history
		contents = append(contents, genai.NewContentFromParts([]*genai.Part{{FunctionCall: functionCall}}, genai.RoleModel))

		// Execute the tool and get the result
		toolResult, err := toolRegistry.HandleToolCall(ctx, functionCall)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrToolCallFailed, err)
		}

		// Add the tool result to conversation history
		contents = append(contents, toolResult)
	}

	return contents, nil
}
