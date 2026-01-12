package analyze_test

import (
	"testing"

	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/e2e/analyze"
)

// TestNewAIAnalyzer tests the creation of a new AI analyzer
func TestNewAIAnalyzer(t *testing.T) {
	analyzer := analyze.NewAIAnalyzer()
	
	if analyzer == nil {
		t.Error("NewAIAnalyzer returned nil")
	}
}

// TestNoOpAnalyzer tests the no-op analyzer
func TestNoOpAnalyzer(t *testing.T) {
	analyzer := analyze.NewNoOpAnalyzer()
	
	if analyzer == nil {
		t.Error("NewNoOpAnalyzer returned nil")
	}
	
	// Verify it never triggers analysis
	result := &orchestrator.ExecutionResult{Success: false}
	if analyzer.ShouldAnalyze(result) {
		t.Error("NoOpAnalyzer should always return false for ShouldAnalyze")
	}
}

