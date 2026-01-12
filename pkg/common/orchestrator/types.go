package orchestrator

import (
	"time"
)

// ClusterInfo represents generic cluster connection information.
// The Properties map allows for provider-specific extensions without
// modifying the core interface.
type ClusterInfo struct {
	ID         string
	Name       string
	Provider   string // e.g., "ocm", "eks", "gke", "kind"
	Region     string
	Version    string
	Kubeconfig []byte
	Properties map[string]interface{} // Extensible for provider-specific data
}

// HealthStatus represents cluster health state.
type HealthStatus struct {
	Ready      bool
	Message    string
	Conditions map[string]bool // e.g., {"nodes": true, "operators": false}
}

// ExecutionTarget defines what and where to execute.
type ExecutionTarget struct {
	Cluster    *ClusterInfo
	Kubeconfig []byte
	Labels     map[string]string // For filtering/categorizing (e.g., "phase:install", "suite:e2e")
	Timeout    time.Duration
}

// ExecutionResult is a generic representation of test execution outcome.
// The Metadata map allows framework-specific data without modifying the interface.
type ExecutionResult struct {
	Success   bool
	StartTime time.Time
	EndTime   time.Time
	Summary   *ResultSummary
	Artifacts []Artifact             // Logs, reports, screenshots, etc.
	Metadata  map[string]interface{} // Framework-specific data
}

// ResultSummary provides high-level test results.
type ResultSummary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
	Errors  []TestError
}

// TestError represents a test failure.
type TestError struct {
	Name    string
	Message string
	Stack   string
	Time    time.Time
}

// Artifact represents a test artifact (log, report, etc.)
type Artifact struct {
	Name     string
	Path     string
	MimeType string
	Size     int64
}

// AnalysisInput provides data for analysis.
type AnalysisInput struct {
	Cluster       *ClusterInfo
	Result        *ExecutionResult
	ArtifactsDir  string
	FailureReason error // Original failure that triggered analysis
}

// AnalysisResult contains analysis findings.
type AnalysisResult struct {
	Summary      string
	RootCause    string
	Suggestions  []string
	Confidence   float64 // 0.0 to 1.0
	AnalysisTime time.Duration
	Metadata     map[string]interface{}
}

// ReportInput provides data for reporting.
type ReportInput struct {
	Cluster  *ClusterInfo
	Result   *ExecutionResult
	Analysis *AnalysisResult // May be nil
	Phase    string          // Contextual label (e.g., "install", "upgrade")
	Metadata map[string]interface{}
}

// Exit codes for orchestrator workflows
const (
	SuccessExitCode = 0
	FailureExitCode = 1
)

