package analysisengine

import (
	"github.com/openshift/osde2e/internal/llm"
	"google.golang.org/genai"
)

// ClusterInfo holds cluster-specific metadata shared by all analysis engines.
type ClusterInfo struct {
	ID            string `json:"id,omitempty" yaml:"id,omitempty"`
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	Provider      string `json:"provider,omitempty" yaml:"provider,omitempty"`
	Region        string `json:"region,omitempty" yaml:"region,omitempty"`
	CloudProvider string `json:"cloudProvider,omitempty" yaml:"cloudProvider,omitempty"`
	Version       string `json:"version,omitempty" yaml:"version,omitempty"`
	Type          string `json:"type,omitempty" yaml:"type,omitempty"` // e.g. "rosa", "osd", "aro"
	Hypershift    bool   `json:"hypershift,omitempty" yaml:"hypershift,omitempty"`
	Environment   string `json:"environment,omitempty" yaml:"environment,omitempty"` // e.g. "stage", "production", "integration"
}

// BaseConfig holds common configuration shared by all analysis engines.
type BaseConfig struct {
	ArtifactsDir string              // Directory containing artifacts or results
	APIKey       string              // LLM API key
	LLMConfig    *llm.AnalysisConfig // Optional LLM configuration overrides
	ClusterInfo  *ClusterInfo        // Cluster metadata for analysis context
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
