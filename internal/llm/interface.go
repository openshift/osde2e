package llm

import "context"

// LLMClient defines the interface for LLM providers
type LLMClient interface {
	// Analyze sends a user prompt with optional configuration
	Analyze(ctx context.Context, userPrompt string, config ...*AnalysisConfig) (*AnalysisResult, error)
	
	// HealthCheck verifies the client can communicate with the provider
	HealthCheck(ctx context.Context) error
	
	// Close cleans up resources
	Close() error
}

// AnalysisConfig contains configuration for LLM requests
type AnalysisConfig struct {
	SystemInstruction *string                `json:"systemInstruction,omitempty"`
	Temperature       *float32               `json:"temperature,omitempty"`
	TopP              *float32               `json:"topP,omitempty"`
	ResponseSchema    interface{}            `json:"responseSchema,omitempty"`
	Tools             []interface{}          `json:"tools,omitempty"`
	MaxTokens         *int                   `json:"maxTokens,omitempty"`
}

// AnalysisResult contains the response from the LLM
type AnalysisResult struct {
	Content  string
	Provider string
	Model    string
}