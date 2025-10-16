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
	"google.golang.org/genai"
	"gopkg.in/yaml.v3"
)

const (
	AnalysisDirName = "llm-analysis"
	SummaryFileName = "summary.yaml"
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
	ArtifactsDir     string
	PromptTemplate   string
	APIKey           string
	LLMConfig        *llm.AnalysisConfig
	FailureContext   string
	ClusterInfo      *ClusterInfo
	EnableMustGather bool // Enable must-gather tool integration
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
	data, err := e.aggregatorService.Collect(ctx, e.config.ArtifactsDir)
	if err != nil {
		return nil, fmt.Errorf("data collection failed: %w", err)
	}

	toolRegistry := tools.NewRegistry(data, &tools.RegistryConfig{
		EnableMustGather: e.config.EnableMustGather,
	})
	// Set up cleanup to run when function exits
	defer func() {
		if err := toolRegistry.Cleanup(); err != nil {
			// Log cleanup error but don't fail the analysis
			fmt.Printf("Warning: Failed to cleanup tools: %v\n", err)
		}
	}()

	vars := make(map[string]any)
	vars["Artifacts"] = data.LogArtifacts
	vars["AnamolyLogs"] = data.AnamolyLogs
	vars["TestResults"] = data.TestResults
	vars["FailedTests"] = data.FailedTests
	vars["FailureContext"] = e.config.FailureContext
	vars["EnableMustGather"] = e.config.EnableMustGather

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

	result, err := e.llmClient.Analyze(ctx, userPrompt, llmConfig, toolRegistry)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	analysisResult := &Result{
		Status:    "completed",
		Content:   result.Content,
		Prompt:    userPrompt,
		ToolCalls: result.ToolCalls,
		Metadata: map[string]any{
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

	if err := analysisResult.WriteSummary(e.config.ArtifactsDir, e.config.ClusterInfo, e.config.FailureContext); err != nil {
		return nil, fmt.Errorf("failed to write analysis files: %w", err)
	}

	return analysisResult, nil
}

// Result represents the analysis output
type Result struct {
	Status    string                `json:"status"`
	Content   string                `json:"content"`
	Metadata  map[string]any        `json:"metadata,omitempty"`
	Error     string                `json:"error,omitempty"`
	Prompt    string                `json:"prompt,omitempty"`
	ToolCalls []*genai.FunctionCall `json:"tool_calls,omitempty"`
}

// WriteSummary writes the analysis result to a YAML summary file
func (res *Result) WriteSummary(reportDir string, clusterInfo *ClusterInfo, failureContext string) error {
	analysisDir := filepath.Join(reportDir, AnalysisDirName)
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	artifactsCount := 0
	if count, ok := res.Metadata["artifacts_examined"].(int); ok {
		artifactsCount = count
	}

	summary := map[string]any{
		"timestamp":          time.Now().Format(time.RFC3339),
		"cluster_info":       clusterInfo,
		"failure_context":    failureContext,
		"artifacts_examined": artifactsCount,
		"status":             res.Status,
		"prompt":             res.Prompt,
		"tool_calls":         res.ToolCalls,
		"response":           res.Content,
		"metadata":           res.Metadata,
		"error":              res.Error,
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
