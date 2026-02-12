package analysisengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/osde2e/internal/aggregator"
	"github.com/openshift/osde2e/internal/llm"
	"github.com/openshift/osde2e/internal/llm/tools"
	"github.com/openshift/osde2e/internal/prompts"
	"github.com/openshift/osde2e/internal/reporter"
	"gopkg.in/yaml.v3"
)

const (
	// KrknAIPromptTemplate is the prompt template for krkn-ai analysis
	KrknAIPromptTemplate = "krknai"
)

// KrknAIConfig holds configuration for the krkn-ai analysis engine.
type KrknAIConfig struct {
	ResultsDir         string                       // Directory containing krkn-ai results
	APIKey             string                       // Gemini API key
	LLMConfig          *llm.AnalysisConfig          // Optional LLM configuration overrides
	NotificationConfig *reporter.NotificationConfig // Optional notification configuration
	TopScenariosCount  int                          // Number of top scenarios to include (default: 10)
}

// KrknAIEngine analyzes krkn-ai chaos test results using LLM.
type KrknAIEngine struct {
	config           *KrknAIConfig
	aggregator       *aggregator.KrknAIAggregator
	promptStore      *prompts.PromptStore
	llmClient        llm.LLMClient
	reporterRegistry *reporter.ReporterRegistry
}

// NewKrknAI creates a new krkn-ai analysis engine.
func NewKrknAI(ctx context.Context, config *KrknAIConfig) (*KrknAIEngine, error) {
	if config.ResultsDir == "" {
		return nil, fmt.Errorf("results directory is required")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required for krkn-ai analysis")
	}

	// Create krkn-ai specific aggregator
	krknAggregator := aggregator.NewKrknAIAggregator(ctx)
	if config.TopScenariosCount > 0 {
		krknAggregator.WithTopScenariosCount(config.TopScenariosCount)
	}

	promptStore, err := prompts.NewPromptStore()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize prompt store: %w", err)
	}

	client, err := llm.NewGeminiClient(ctx, config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	// Initialize reporter registry
	reporterRegistry := reporter.NewReporterRegistry()
	reporterRegistry.Register(reporter.NewSlackReporter())

	return &KrknAIEngine{
		config:           config,
		aggregator:       krknAggregator,
		promptStore:      promptStore,
		llmClient:        client,
		reporterRegistry: reporterRegistry,
	}, nil
}

// Run executes the krkn-ai analysis workflow.
func (e *KrknAIEngine) Run(ctx context.Context) (*Result, error) {
	// Collect krkn-ai results
	data, err := e.aggregator.Collect(ctx, e.config.ResultsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to collect krkn-ai results: %w", err)
	}

	// Create tool registry with log artifacts for read_file tool
	toolRegistry := tools.NewRegistry(data.LogArtifacts)

	// Prepare template variables
	vars := map[string]any{
		"Summary":           data.Summary,
		"TopScenarios":      data.TopScenarios,
		"FailedScenarios":   data.FailedScenarios,
		"HealthCheckReport": data.HealthCheckReport,
		"LogArtifacts":      data.LogArtifacts,
		"ConfigSummary":     data.ConfigSummary,
	}

	// Render prompt
	userPrompt, llmConfig, err := e.promptStore.RenderPrompt(KrknAIPromptTemplate, vars)
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
	analysisResult := &Result{
		Status:    "completed",
		Content:   result.Content,
		Prompt:    userPrompt,
		ToolCalls: result.ToolCalls,
		Metadata: map[string]any{
			"analysis_type":        "krknai",
			"total_scenarios":      data.Summary.TotalScenarios,
			"successful_scenarios": data.Summary.SuccessfulScenarios,
			"failed_scenarios":     data.Summary.FailedScenarios,
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

	// Send notifications if configured
	if e.config.NotificationConfig != nil && e.config.NotificationConfig.Enabled {
		e.sendNotifications(ctx, analysisResult)
	}

	return analysisResult, nil
}

// writeSummary writes the analysis result to a YAML summary file.
func (e *KrknAIEngine) writeSummary(result *Result, data *aggregator.KrknAIData) error {
	analysisDir := filepath.Join(e.config.ResultsDir, AnalysisDirName)
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	summary := map[string]any{
		"timestamp":     time.Now().Format(time.RFC3339),
		"analysis_type": "krknai",
		"run_summary": map[string]any{
			"total_scenarios":      data.Summary.TotalScenarios,
			"successful_scenarios": data.Summary.SuccessfulScenarios,
			"failed_scenarios":     data.Summary.FailedScenarios,
			"generations":          data.Summary.Generations,
			"max_fitness_score":    data.Summary.MaxFitnessScore,
			"avg_fitness_score":    data.Summary.AvgFitnessScore,
			"scenario_types":       data.Summary.ScenarioTypes,
		},
		"top_scenarios":    data.TopScenarios,
		"failed_scenarios": data.FailedScenarios,
		"status":           result.Status,
		"prompt":           result.Prompt,
		"tool_calls":       result.ToolCalls,
		"response":         result.Content,
		"metadata":         result.Metadata,
		"error":            result.Error,
	}

	yamlData, err := yaml.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary to YAML: %w", err)
	}

	summaryPath := filepath.Join(analysisDir, SummaryFileName)
	if err := os.WriteFile(summaryPath, yamlData, 0o644); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	return nil
}

// sendNotifications sends analysis results to configured reporters.
func (e *KrknAIEngine) sendNotifications(ctx context.Context, result *Result) {
	reporterResult := &reporter.AnalysisResult{
		Status:   result.Status,
		Content:  result.Content,
		Metadata: result.Metadata,
		Error:    result.Error,
		Prompt:   result.Prompt,
	}

	for _, reporterConfig := range e.config.NotificationConfig.Reporters {
		if err := e.reporterRegistry.SendNotification(ctx, reporterResult, &reporterConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to send notification via %s: %v\n", reporterConfig.Type, err)
		}
	}
}
