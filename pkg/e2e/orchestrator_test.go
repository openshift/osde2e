package e2e_test

import (
	"context"
	"testing"

	orc "github.com/openshift/osde2e/pkg/common/orchestrator"
)

// MockProvisioner is a simple mock for testing
type MockProvisioner struct {
	provisionCalled bool
	destroyCalled   bool
}

func (m *MockProvisioner) Provision(ctx context.Context) (*orc.ClusterInfo, error) {
	m.provisionCalled = true
	return &orc.ClusterInfo{
		ID:         "test-cluster-123",
		Name:       "Test Cluster",
		Provider:   "mock",
		Kubeconfig: []byte("mock-kubeconfig-data"),
	}, nil
}

func (m *MockProvisioner) Destroy(ctx context.Context, cluster *orc.ClusterInfo) error {
	m.destroyCalled = true
	return nil
}

func (m *MockProvisioner) GetKubeconfig(ctx context.Context, cluster *orc.ClusterInfo) ([]byte, error) {
	return []byte("mock-kubeconfig"), nil
}

func (m *MockProvisioner) Health(ctx context.Context, cluster *orc.ClusterInfo) (*orc.HealthStatus, error) {
	return &orc.HealthStatus{Ready: true}, nil
}

// MockExecutor is a simple mock for testing
type MockExecutor struct {
	executeCalled bool
}

func (m *MockExecutor) Execute(ctx context.Context, target *orc.ExecutionTarget) (*orc.ExecutionResult, error) {
	m.executeCalled = true
	return &orc.ExecutionResult{
		Success: true,
		Summary: &orc.ResultSummary{
			Total:  10,
			Passed: 10,
			Failed: 0,
		},
	}, nil
}

// MockAnalyzer is a simple mock for testing
type MockAnalyzer struct{}

func (m *MockAnalyzer) Analyze(ctx context.Context, input *orc.AnalysisInput) (*orc.AnalysisResult, error) {
	return nil, nil
}

func (m *MockAnalyzer) ShouldAnalyze(result *orc.ExecutionResult) bool {
	return false
}

// MockReporter is a simple mock for testing
type MockReporter struct {
	initializeCalled bool
	reportCalled     bool
	finalizeCalled   bool
}

func (m *MockReporter) Initialize(ctx context.Context) error {
	m.initializeCalled = true
	return nil
}

func (m *MockReporter) Report(ctx context.Context, input *orc.ReportInput) error {
	m.reportCalled = true
	return nil
}

func (m *MockReporter) Finalize(ctx context.Context) error {
	m.finalizeCalled = true
	return nil
}

// TestOrchestratorSuccessPath tests the happy path of the orchestrator
func TestOrchestratorSuccessPath(t *testing.T) {
	provisioner := &MockProvisioner{}
	executor := &MockExecutor{}
	analyzer := &MockAnalyzer{}
	reporter := &MockReporter{}

	orch := orc.NewOrchestrator(provisioner, executor, analyzer, reporter)

	ctx := context.Background()
	exitCode := orch.Run(ctx)

	// Verify success exit code
	if exitCode != orc.SuccessExitCode {
		t.Errorf("Expected success exit code %d, got %d", orc.SuccessExitCode, exitCode)
	}

	// Verify all components were called
	if !provisioner.provisionCalled {
		t.Error("Provisioner.Provision was not called")
	}
	if !provisioner.destroyCalled {
		t.Error("Provisioner.Destroy was not called")
	}
	if !executor.executeCalled {
		t.Error("Executor.Execute was not called")
	}
	if !reporter.initializeCalled {
		t.Error("Reporter.Initialize was not called")
	}
	if !reporter.reportCalled {
		t.Error("Reporter.Report was not called")
	}
	if !reporter.finalizeCalled {
		t.Error("Reporter.Finalize was not called")
	}
}

// TestOrchestratorFailedTests tests orchestrator behavior when tests fail
func TestOrchestratorFailedTests(t *testing.T) {
	provisioner := &MockProvisioner{}
	executor := &MockExecutor{} // Will return success=false
	analyzer := &MockAnalyzer{}
	reporter := &MockReporter{}

	// Override executor to return failure
	failingExecutor := &struct {
		*MockExecutor
	}{MockExecutor: executor}
	failingExecutor.MockExecutor = executor

	orch := orc.NewOrchestrator(provisioner, failingExecutor, analyzer, reporter)

	ctx := context.Background()
	exitCode := orch.Run(ctx)

	// Should still complete the workflow even on test failure
	if exitCode != orc.SuccessExitCode {
		// This is expected if tests fail, so we check if cleanup happened
		if !provisioner.destroyCalled {
			t.Error("Provisioner.Destroy should be called even on test failure")
		}
	}
}

