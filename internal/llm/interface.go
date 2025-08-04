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
