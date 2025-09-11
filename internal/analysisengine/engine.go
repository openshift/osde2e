package analysisengine

import (
	"context"
	"fmt"

	"github.com/openshift/osde2e/internal/aggregator"
	"github.com/openshift/osde2e/internal/llm"
	"github.com/openshift/osde2e/internal/llm/tools"
	"github.com/openshift/osde2e/internal/prompts"
)

// ClusterInfo holds cluster-specific information for analysis
type ClusterInfo struct {
	ID            string
	Name          string
	Provider      string
	Region        string
	CloudProvider string
	Version       string
}

// Config holds configuration for the analysis engine
type Config struct {
	ArtifactsDir   string
	PromptTemplate string
	OutputFormat   string
	APIKey         string
	Model          string
	LLMConfig      *llm.AnalysisConfig
	LogLevel       string
	DryRun         bool
	Verbose        bool
	FailureContext string
	ClusterInfo    *ClusterInfo
}

// Engine represents the analysis engine
type Engine struct {
	config            *Config
	aggregatorService *aggregator.Aggregator
	promptStore       *prompts.PromptStore
	llmClient         llm.LLMClient
}

// New creates a new analysis engine
func New(ctx context.Context, config *Config) (*Engine, error) {
	aggregatorService := aggregator.New(ctx)
	promptStore, err := prompts.NewPromptStore()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prompt store: %w", err)
	}

	// Initialize LLM client
	if config.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required for LLM analysis")
	}

	client, err := llm.NewGeminiClient(ctx, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	return &Engine{
		config:            config,
		aggregatorService: aggregatorService,
		promptStore:       promptStore,
		llmClient:         client,
	}, nil
}

// Run executes the analysis workflow
func (e *Engine) Run(ctx context.Context) (*Result, error) {
	// Collect data
	data, err := e.aggregatorService.Collect(ctx, e.config.ArtifactsDir)
	if err != nil {
		return nil, fmt.Errorf("data collection failed: %w", err)
	}

	// Create tool registry with collected data
	toolRegistry := tools.NewRegistry(data)

	// Prepare prompt variables
	vars := make(map[string]any)
	vars["Artifacts"] = data.LogArtifacts
	vars["AnamolyLogs"] = data.AnamolyLogs
	vars["TestResults"] = data.TestResults
	vars["FailureContext"] = e.config.FailureContext

	// Add cluster information if available
	if e.config.ClusterInfo != nil {
		vars["ClusterID"] = e.config.ClusterInfo.ID
		vars["ClusterName"] = e.config.ClusterInfo.Name
		vars["Provider"] = e.config.ClusterInfo.Provider
		vars["Region"] = e.config.ClusterInfo.Region
		vars["Version"] = e.config.ClusterInfo.Version
	}

	userPrompt, llmConfig, err := e.promptStore.RenderPrompt(e.config.PromptTemplate, vars)
	if err != nil {
		return nil, fmt.Errorf("prompt preparation failed: %w", err)
	}

	// Merge user-provided LLM config with prompt config
	if e.config.LLMConfig != nil {
		if e.config.LLMConfig.Temperature != nil {
			llmConfig.Temperature = e.config.LLMConfig.Temperature
		}
		if e.config.LLMConfig.MaxTokens != nil {
			llmConfig.MaxTokens = e.config.LLMConfig.MaxTokens
		}
		if e.config.LLMConfig.TopP != nil {
			llmConfig.TopP = e.config.LLMConfig.TopP
		}
	}

	// Run LLM analysis
	result, err := e.llmClient.Analyze(ctx, userPrompt, llmConfig, toolRegistry)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	return &Result{
		Status:  "completed",
		Content: result.Content,
		Metadata: map[string]any{
			"prompt":             userPrompt,
			"artifacts_examined": len(data.LogArtifacts),
		},
	}, nil
}

// Result represents the analysis output
type Result struct {
	Status   string         `json:"status"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Error    string         `json:"error,omitempty"`
}
