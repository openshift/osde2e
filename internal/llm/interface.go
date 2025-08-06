package llm

import "context"

type LLMClient interface {
	Analyze(ctx context.Context, userPrompt string, config *AnalysisConfig) (*AnalysisResult, error)
}
