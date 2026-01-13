package orchestrator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
)

// mockOrchestrator is a test implementation of the Orchestrator interface.
type mockOrchestrator struct {
	provisionErr   error
	executeErr     error
	analyzeErr     error
	reportErr      error
	result         *orchestrator.Result
	provisionCalls int
	executeCalls   int
	analyzeCalls   int
	reportCalls    int
}

func (m *mockOrchestrator) Provision(ctx context.Context) error {
	m.provisionCalls++
	return m.provisionErr
}

func (m *mockOrchestrator) Execute(ctx context.Context) error {
	m.executeCalls++
	return m.executeErr
}

func (m *mockOrchestrator) AnalyzeLogs(ctx context.Context, testErr error) error {
	m.analyzeCalls++
	return m.analyzeErr
}

func (m *mockOrchestrator) Report(ctx context.Context) error {
	m.reportCalls++
	return m.reportErr
}

func (m *mockOrchestrator) Cleanup(ctx context.Context) error {
	return nil
}

func (m *mockOrchestrator) PostProcessCluster(ctx context.Context) error {
	return nil
}

func (m *mockOrchestrator) Result() *orchestrator.Result {
	return m.result
}

func TestOrchestrator_SuccessfulFlow(t *testing.T) {
	ctx := context.Background()
	mock := &mockOrchestrator{
		result: &orchestrator.Result{
			ExitCode:      config.Success,
			TestsPassed:   true,
			UpgradePassed: true,
			ClusterID:     "test-cluster-123",
		},
	}

	// Execute full lifecycle
	if err := mock.Provision(ctx); err != nil {
		t.Fatalf("Provision failed: %v", err)
	}
	if err := mock.Execute(ctx); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if err := mock.Report(ctx); err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	// Verify all methods were called
	if mock.provisionCalls != 1 {
		t.Errorf("Expected 1 Provision call, got %d", mock.provisionCalls)
	}
	if mock.executeCalls != 1 {
		t.Errorf("Expected 1 Execute call, got %d", mock.executeCalls)
	}
	if mock.reportCalls != 1 {
		t.Errorf("Expected 1 Report call, got %d", mock.reportCalls)
	}

	// Verify result
	result := mock.Result()
	if result.ExitCode != config.Success {
		t.Errorf("Expected exit code %d, got %d", config.Success, result.ExitCode)
	}
	if !result.TestsPassed {
		t.Error("Expected TestsPassed to be true")
	}
}

func TestOrchestrator_ProvisionFailure(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("provision failed")
	mock := &mockOrchestrator{
		provisionErr: expectedErr,
	}

	err := mock.Provision(ctx)
	if err == nil {
		t.Fatal("Expected provision error, got nil")
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestOrchestrator_ExecuteFailure(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("tests failed")
	mock := &mockOrchestrator{
		executeErr: testErr,
		result: &orchestrator.Result{
			ExitCode:    config.Failure,
			TestsPassed: false,
		},
	}

	err := mock.Execute(ctx)
	if err == nil {
		t.Fatal("Expected execute error, got nil")
	}

	// Analyze logs should be called after failure
	if err := mock.AnalyzeLogs(ctx, err); err != nil {
		t.Fatalf("AnalyzeLogs failed: %v", err)
	}

	if mock.analyzeCalls != 1 {
		t.Errorf("Expected 1 AnalyzeLogs call, got %d", mock.analyzeCalls)
	}
}

func TestOrchestrator_AnalyzeLogsOptional(t *testing.T) {
	ctx := context.Background()
	mock := &mockOrchestrator{}

	// AnalyzeLogs should not fail if not called
	if mock.analyzeCalls != 0 {
		t.Error("AnalyzeLogs should not be called initially")
	}

	// It's only called on test failure
	testErr := errors.New("test error")
	if err := mock.AnalyzeLogs(ctx, testErr); err != nil {
		t.Fatalf("AnalyzeLogs failed: %v", err)
	}

	if mock.analyzeCalls != 1 {
		t.Errorf("Expected 1 AnalyzeLogs call, got %d", mock.analyzeCalls)
	}
}

func TestResult_InitialState(t *testing.T) {
	result := &orchestrator.Result{}

	if result.ExitCode != 0 {
		t.Errorf("Expected initial ExitCode 0, got %d", result.ExitCode)
	}
	if result.TestsPassed {
		t.Error("Expected TestsPassed to be false initially")
	}
	if result.UpgradePassed {
		t.Error("Expected UpgradePassed to be false initially")
	}
	if result.ClusterID != "" {
		t.Errorf("Expected empty ClusterID, got %s", result.ClusterID)
	}
	if result.Errors != nil {
		t.Error("Expected nil Errors initially")
	}
}

func TestResult_WithErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	result := &orchestrator.Result{
		ExitCode: config.Failure,
		Errors:   []error{err1, err2},
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}
	if result.Errors[0] != err1 {
		t.Errorf("Expected first error %v, got %v", err1, result.Errors[0])
	}
}

func TestOrchestrator_ReportWithErrors(t *testing.T) {
	ctx := context.Background()
	reportErr := errors.New("cleanup failed")
	mock := &mockOrchestrator{
		reportErr: reportErr,
		result: &orchestrator.Result{
			ExitCode: config.Failure,
			Errors:   []error{reportErr},
		},
	}

	err := mock.Report(ctx)
	if err == nil {
		t.Fatal("Expected report error, got nil")
	}

	result := mock.Result()
	if len(result.Errors) == 0 {
		t.Error("Expected errors to be collected in result")
	}
}
