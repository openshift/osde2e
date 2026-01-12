package orchestrator

import (
	"context"
	"log"
	"time"
)

// Orchestrator coordinates testing workflows using the four core interfaces.
// It is generic and can be used by any testing framework.
type Orchestrator struct {
	provisioner Provisioner
	executor    Executor
	analyzer    Analyzer
	reporter    Reporter
}

// NewOrchestrator creates a new orchestrator with the given components.
func NewOrchestrator(
	provisioner Provisioner,
	executor Executor,
	analyzer Analyzer,
	reporter Reporter,
) *Orchestrator {
	return &Orchestrator{
		provisioner: provisioner,
		executor:    executor,
		analyzer:    analyzer,
		reporter:    reporter,
	}
}

// Run executes the complete testing workflow.
// It coordinates: Provision -> Initialize Reporter -> Execute -> Report -> Analyze (if failed) -> Cleanup
func (o *Orchestrator) Run(ctx context.Context) (exitCode int) {
	log.Println("Starting orchestrator...")
	startTime := time.Now()

	// 1. Provision cluster
	log.Println("=== Phase 1: Provisioning Cluster ===")
	cluster, err := o.provisioner.Provision(ctx)
	if err != nil {
		log.Printf("❌ Provisioning failed: %v", err)
		return FailureExitCode
	}
	log.Printf("✓ Cluster provisioned: %s (ID: %s)", cluster.Name, cluster.ID)

	// Ensure cleanup happens
	defer func() {
		log.Println("=== Cleanup Phase ===")
		if destroyErr := o.provisioner.Destroy(ctx, cluster); destroyErr != nil {
			log.Printf("⚠ Cleanup error: %v", destroyErr)
		}
	}()

	// 2. Initialize reporter
	log.Println("=== Phase 2: Initializing Reporter ===")
	if err := o.reporter.Initialize(ctx); err != nil {
		log.Printf("❌ Reporter initialization failed: %v", err)
		return FailureExitCode
	}
	log.Println("✓ Reporter initialized")

	defer func() {
		if finalizeErr := o.reporter.Finalize(ctx); finalizeErr != nil {
			log.Printf("⚠ Reporter finalization error: %v", finalizeErr)
		}
	}()

	// 3. Execute tests
	log.Println("=== Phase 3: Executing Tests ===")

	// Build execution target from cluster info
	target := &ExecutionTarget{
		Cluster:    cluster,
		Kubeconfig: cluster.Kubeconfig,
		Labels:     make(map[string]string),
		Timeout:    1 * time.Hour, // Default timeout, can be overridden by implementations
	}

	// Execute tests
	result, execErr := o.executor.Execute(ctx, target)
	if execErr != nil {
		log.Printf("❌ Test execution error: %v", execErr)
		// Don't return yet - we still want to report and analyze
	}

	if result != nil {
		if result.Success {
			log.Printf("✓ Tests passed (%d/%d)", result.Summary.Passed, result.Summary.Total)
		} else {
			log.Printf("❌ Tests failed (%d passed, %d failed, %d total)",
				result.Summary.Passed, result.Summary.Failed, result.Summary.Total)
		}
	}

	// 4. Report results
	log.Println("=== Phase 4: Reporting Results ===")
	reportInput := &ReportInput{
		Cluster:  cluster,
		Result:   result,
		Metadata: make(map[string]interface{}),
	}

	if err := o.reporter.Report(ctx, reportInput); err != nil {
		log.Printf("⚠ Reporting error: %v", err)
	} else {
		log.Println("✓ Results reported")
	}

	// 5. Analyze failures (if needed)
	if result != nil && o.analyzer.ShouldAnalyze(result) {
		log.Println("=== Phase 5: Analyzing Failures ===")

		analysisInput := &AnalysisInput{
			Cluster:       cluster,
			Result:        result,
			FailureReason: execErr,
		}

		analysisResult, analyzeErr := o.analyzer.Analyze(ctx, analysisInput)
		if analyzeErr != nil {
			log.Printf("⚠ Analysis error: %v", analyzeErr)
		} else if analysisResult != nil {
			log.Printf("✓ Analysis completed (confidence: %.2f)", analysisResult.Confidence)
			log.Printf("Root cause: %s", analysisResult.RootCause)

			// Report analysis results
			reportInput.Analysis = analysisResult
			if err := o.reporter.Report(ctx, reportInput); err != nil {
				log.Printf("⚠ Failed to report analysis: %v", err)
			}
		}
	}

	// Determine final exit code
	duration := time.Since(startTime)
	log.Printf("=== Workflow Complete (duration: %v) ===", duration)

	if result != nil && result.Success {
		log.Println("✓ SUCCESS")
		return SuccessExitCode
	}

	log.Println("❌ FAILURE")
	return FailureExitCode
}
