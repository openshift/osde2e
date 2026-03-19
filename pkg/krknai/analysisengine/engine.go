package analysisengine

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/osde2e/internal/analysisengine"
	"github.com/openshift/osde2e/internal/llm"
	"github.com/openshift/osde2e/internal/llm/tools"
	"github.com/openshift/osde2e/internal/prompts"
	krknAggregator "github.com/openshift/osde2e/pkg/krknai/aggregator"
	"gopkg.in/yaml.v3"
)

//go:embed prompts/*
var krknPrompts embed.FS

const (
	analysisDirName = "llm-analysis"
	summaryFileName = "summary.yaml"

	krknAIPromptTemplate = "krknai"
)

// Config holds configuration for the krkn-ai analysis engine.
type Config struct {
	analysisengine.BaseConfig
	TopScenariosCount int // Number of top scenarios to include (default: 10)
}

// Engine analyzes krkn-ai chaos test results using LLM.
type Engine struct {
	config      *Config
	aggregator  *krknAggregator.KrknAIAggregator
	promptStore *prompts.PromptStore
	llmClient   llm.LLMClient
}

// New creates a new krkn-ai analysis engine.
func New(ctx context.Context, config *Config) (*Engine, error) {
	if config.ArtifactsDir == "" {
		return nil, fmt.Errorf("results directory is required")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required for krkn-ai analysis")
	}

	// Create krkn-ai specific aggregator
	agg := krknAggregator.NewKrknAIAggregator(ctx)
	if config.TopScenariosCount > 0 {
		agg.WithTopScenariosCount(config.TopScenariosCount)
	}

	promptStore, err := prompts.NewPromptStore(prompts.DefaultTemplates())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prompt store: %w", err)
	}

	localFS, err := fs.Sub(krknPrompts, "prompts")
	if err != nil {
		return nil, fmt.Errorf("failed to load krkn-ai prompt templates: %w", err)
	}
	if err := promptStore.RegisterTemplates(localFS); err != nil {
		return nil, fmt.Errorf("failed to register krkn-ai prompt templates: %w", err)
	}

	client, err := llm.NewGeminiClient(ctx, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	return &Engine{
		config:      config,
		aggregator:  agg,
		promptStore: promptStore,
		llmClient:   client,
	}, nil
}

// Run executes the krkn-ai analysis workflow.
func (e *Engine) Run(ctx context.Context) (*analysisengine.Result, error) {
	// Collect krkn-ai results
	data, err := e.aggregator.Collect(ctx, e.config.ArtifactsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to collect krkn-ai results: %w", err)
	}

	// Create tool registry with log artifacts for read_file tool
	toolRegistry := tools.NewRegistry(data.LogArtifacts)

	// Prepare template variables from collected data
	vars := map[string]any{
		"Summary":           data.Summary,
		"TopScenarios":      data.TopScenarios,
		"FailedScenarios":   data.FailedScenarios,
		"HealthCheckReport": data.HealthCheckReport,
		"LogArtifacts":      data.LogArtifacts,
		"ConfigSummary":     data.ConfigSummary,
	}
	if e.config.ClusterInfo != nil {
		vars["ClusterInfo"] = e.config.ClusterInfo
	}

	// Render prompt using prompt store
	userPrompt, llmConfig, err := e.promptStore.RenderPrompt(krknAIPromptTemplate, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	// Apply LLM config overrides
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

	// Build analysis result
	analysisResult := &analysisengine.Result{
		Status:  "completed",
		Content: result.Content,
		Prompt:  userPrompt,
		Metadata: map[string]any{
			"analysis_type":        "krknai",
			"total_scenarios":      data.Summary.TotalScenarioCount,
			"successful_scenarios": data.Summary.SuccessfulScenarioCount,
			"failed_scenarios":     data.Summary.FailedScenarioCount,
			"generations":          data.Summary.Generations,
			"max_fitness_score":    data.Summary.MaxFitnessScore,
			"artifacts_examined": func() (count int) {
				for _, tc := range result.ToolCalls {
					if tc.Name == "read_file" {
						count++
					}
				}
				return count
			}(),
			"tool_calls": len(result.ToolCalls),
		},
	}

	// Write summary to results directory
	if err := e.writeSummary(analysisResult, data); err != nil {
		return nil, fmt.Errorf("failed to write analysis summary: %w", err)
	}

	return analysisResult, nil
}

// writeSummary writes the analysis result to a YAML summary file.
func (e *Engine) writeSummary(result *analysisengine.Result, data *krknAggregator.KrknAIData) error {
	analysisDir := filepath.Join(e.config.ArtifactsDir, analysisDirName)
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	summary := map[string]any{
		"timestamp":     time.Now().Format(time.RFC3339),
		"analysis_type": "krknai",
		"cluster_info":  e.config.ClusterInfo,
		"run_summary": map[string]any{
			"total_scenarios":      data.Summary.TotalScenarioCount,
			"successful_scenarios": data.Summary.SuccessfulScenarioCount,
			"failed_scenarios":     data.Summary.FailedScenarioCount,
			"generations":          data.Summary.Generations,
			"max_fitness_score":    data.Summary.MaxFitnessScore,
			"avg_fitness_score":    data.Summary.AvgFitnessScore,
			"scenario_types":       data.Summary.ScenarioTypes,
		},
		"top_scenarios":    data.TopScenarios,
		"failed_scenarios": data.FailedScenarios,
		"status":           result.Status,
		"prompt":           result.Prompt,
		"response":         result.Content,
		"metadata":         result.Metadata,
		"error":            result.Error,
	}

	yamlData, err := yaml.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary to YAML: %w", err)
	}

	summaryPath := filepath.Join(analysisDir, summaryFileName)
	if err := os.WriteFile(summaryPath, yamlData, 0o644); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	return nil
}
