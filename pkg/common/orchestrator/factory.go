package orchestrator

// NewOrchestratorWithComponents creates an orchestrator with the provided component instances.
func NewOrchestratorWithComponents(
	provisioner Provisioner,
	executor Executor,
	analyzer Analyzer,
	reporter Reporter,
) *Orchestrator {
	return NewOrchestrator(provisioner, executor, analyzer, reporter)
}
