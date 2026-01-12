package analyze

import (
	"context"

	"github.com/openshift/osde2e/pkg/common/orchestrator"
)

// NoOpAnalyzer is a no-op implementation of the Analyzer interface.
// Useful for testing or when analysis is disabled.
type NoOpAnalyzer struct{}

// NewNoOpAnalyzer creates a new no-op analyzer.
func NewNoOpAnalyzer() *NoOpAnalyzer {
	return &NoOpAnalyzer{}
}

// Analyze always returns nil (no analysis performed).
func (a *NoOpAnalyzer) Analyze(ctx context.Context, input *orchestrator.AnalysisInput) (*orchestrator.AnalysisResult, error) {
	return nil, nil
}

// ShouldAnalyze always returns false.
func (a *NoOpAnalyzer) ShouldAnalyze(result *orchestrator.ExecutionResult) bool {
	return false
}

