package llm

import (
	"context"

	"github.com/openshift/osde2e/internal/llm/tools"
)

type LLMClient interface {
	Analyze(ctx context.Context, userPrompt string, config *AnalysisConfig, toolRegistry *tools.Registry) (*AnalysisResult, error)
}
