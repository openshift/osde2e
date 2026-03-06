package analysisengine

import (
	"github.com/openshift/osde2e/internal/llm"
	"google.golang.org/genai"
)

// BaseConfig holds common configuration shared by all analysis engines.
type BaseConfig struct {
	ArtifactsDir string              // Directory containing artifacts or results
	APIKey       string              // LLM API key
	LLMConfig    *llm.AnalysisConfig // Optional LLM configuration overrides
}

// Result represents the analysis output shared across all engines.
type Result struct {
	Status    string                `json:"status"`
	Content   string                `json:"content"`
	Metadata  map[string]any        `json:"metadata,omitempty"`
	Error     string                `json:"error,omitempty"`
	Prompt    string                `json:"prompt,omitempty"`
	ToolCalls []*genai.FunctionCall `json:"tool_calls,omitempty"`
}
