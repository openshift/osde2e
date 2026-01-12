// Package analyze provides implementations of the Analyzer interface
// for various analysis strategies.
package analyze

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openshift/osde2e/internal/analysisengine"
	"github.com/openshift/osde2e/internal/reporter"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
)

// AIAnalyzer implements the Analyzer interface using AI-powered log analysis.
type AIAnalyzer struct {
	enabled bool
}

// NewAIAnalyzer creates a new AI-powered analyzer.
func NewAIAnalyzer() *AIAnalyzer {
	return &AIAnalyzer{
		enabled: viper.GetBool(config.LogAnalysis.EnableAnalysis),
	}
}

// Analyze examines test results and artifacts to provide insights.
func (a *AIAnalyzer) Analyze(ctx context.Context, input *orchestrator.AnalysisInput) (*orchestrator.AnalysisResult, error) {
	if !a.enabled {
		return nil, nil
	}

	log.Println("Running AI-powered log analysis...")

	if input.ArtifactsDir == "" {
		return nil, fmt.Errorf("artifacts directory is required for analysis")
	}

	// Build cluster info for analysis engine
	clusterInfo := &analysisengine.ClusterInfo{
		ID:            input.Cluster.ID,
		Name:          input.Cluster.Name,
		Provider:      input.Cluster.Provider,
		Region:        input.Cluster.Region,
		CloudProvider: input.Cluster.Provider,
		Version:       input.Cluster.Version,
	}

	// Setup notification config
	var notificationConfig *reporter.NotificationConfig
	var reporters []reporter.ReporterConfig

	// Add Slack reporter if enabled
	enableSlackNotify := viper.GetBool(config.Tests.EnableSlackNotify)
	slackWebhook := viper.GetString(config.LogAnalysis.SlackWebhook)
	defaultChannel := viper.GetString(config.LogAnalysis.SlackChannel)
	if enableSlackNotify && slackWebhook != "" && defaultChannel != "" {
		slackConfig := reporter.SlackReporterConfig(slackWebhook, true)
		slackConfig.Settings["channel"] = defaultChannel
		reporters = append(reporters, slackConfig)
	}

	// Create notification config if we have any reporters
	if len(reporters) > 0 {
		notificationConfig = &reporter.NotificationConfig{
			Enabled:   true,
			Reporters: reporters,
		}
	}

	// Build failure context
	failureContext := "Test execution failed"
	if input.FailureReason != nil {
		failureContext = input.FailureReason.Error()
	}
	if input.Result != nil && input.Result.Summary != nil && len(input.Result.Summary.Errors) > 0 {
		failureContext = fmt.Sprintf("%s. First error: %s",
			failureContext, input.Result.Summary.Errors[0].Message)
	}

	// Create analysis engine config
	engineConfig := &analysisengine.Config{
		ArtifactsDir:       input.ArtifactsDir,
		PromptTemplate:     "default",
		APIKey:             viper.GetString(config.LogAnalysis.APIKey),
		FailureContext:     failureContext,
		ClusterInfo:        clusterInfo,
		NotificationConfig: notificationConfig,
	}

	// Create and run analysis engine
	engine, err := analysisengine.New(ctx, engineConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create analysis engine: %w", err)
	}

	startTime := time.Now()
	engineResult, err := engine.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("log analysis failed: %w", err)
	}
	analysisTime := time.Since(startTime)

	log.Printf("Log analysis completed successfully. Results written to %s/%s/",
		input.ArtifactsDir, analysisengine.AnalysisDirName)
	log.Printf("=== Log Analysis Result ===\n%s", engineResult.Content)

	// Convert engine result to generic AnalysisResult
	// Parse root cause and suggestions from content if available
	rootCause := ""
	suggestions := []string{}
	if engineResult.Content != "" {
		// For now, just use the full content as summary
		// Future enhancement: parse structured output from LLM
		rootCause = "See analysis summary"
		suggestions = []string{"Review the detailed analysis"}
	}

	result := &orchestrator.AnalysisResult{
		Summary:      engineResult.Content,
		RootCause:    rootCause,
		Suggestions:  suggestions,
		Confidence:   0.8, // Default confidence
		AnalysisTime: analysisTime,
		Metadata: map[string]interface{}{
			"analyzer": "ai-gemini",
			"model":    "gemini",
			"status":   engineResult.Status,
		},
	}

	return result, nil
}

// ShouldAnalyze determines if analysis should run based on results.
func (a *AIAnalyzer) ShouldAnalyze(result *orchestrator.ExecutionResult) bool {
	if !a.enabled {
		return false
	}

	// Only analyze on failure
	return !result.Success
}

