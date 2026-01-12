package e2e

import (
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/e2e/analyze"
	"github.com/openshift/osde2e/pkg/e2e/execute"
	"github.com/openshift/osde2e/pkg/e2e/provision"
	"github.com/openshift/osde2e/pkg/e2e/report"
)

// NewOrchestrator creates an E2E orchestrator with standard components:
// - OCM Provisioner (cluster lifecycle management)
// - Ginkgo Executor (test execution)
// - AI Analyzer (log analysis on failures)
// - Composite Reporter (artifacts + must-gather)
func NewOrchestrator() (*orchestrator.Orchestrator, error) {
	// Create provisioner
	provisioner, err := provision.NewOCMProvisioner()
	if err != nil {
		return nil, err
	}

	// Create executor
	executor := execute.NewGinkgoExecutor()

	// Create analyzer
	analyzer := analyze.NewAIAnalyzer()

	// Create reporter
	compositeReporter := report.NewCompositeReporter()
	compositeReporter.AddReporter(report.NewArtifactCollector())

	return orchestrator.NewOrchestratorWithComponents(provisioner, executor, analyzer, compositeReporter), nil
}

// NewOrchestratorWithComponents creates an orchestrator with custom component instances.
func NewOrchestratorWithComponents(
	provisioner orchestrator.Provisioner,
	executor orchestrator.Executor,
	analyzer orchestrator.Analyzer,
	reporter orchestrator.Reporter,
) *orchestrator.Orchestrator {
	return orchestrator.NewOrchestratorWithComponents(provisioner, executor, analyzer, reporter)
}
