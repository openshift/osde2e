// Package orchestrator provides generic interfaces for orchestrating
// end-to-end testing workflows. These interfaces can be used by any
// testing framework, not just osde2e.
package orchestrator

import (
	"context"
)

// Provisioner manages cluster infrastructure lifecycle.
// Implementations can provision clusters from various providers (OCM, EKS, GKE, local kind, etc.)
type Provisioner interface {
	// Provision ensures a cluster is available and returns connection details.
	// Implementation may create a new cluster or reuse an existing one based on config.
	Provision(ctx context.Context) (*ClusterInfo, error)

	// Destroy tears down the cluster if configured to do so.
	// May be a no-op if the cluster should be preserved for debugging.
	Destroy(ctx context.Context, cluster *ClusterInfo) error

	// GetKubeconfig returns raw kubeconfig bytes for cluster access.
	GetKubeconfig(ctx context.Context, cluster *ClusterInfo) ([]byte, error)

	// Health checks cluster health and readiness.
	Health(ctx context.Context, cluster *ClusterInfo) (*HealthStatus, error)
}

// Executor runs tests or workloads against a cluster.
// Implementations can use different test frameworks (Ginkgo, pytest, shell scripts, etc.)
type Executor interface {
	// Execute runs the configured test suite against the cluster.
	// Returns results that can be interpreted by reporters.
	Execute(ctx context.Context, target *ExecutionTarget) (*ExecutionResult, error)
}

// Analyzer performs post-execution analysis.
// Implementations can use AI/LLM, rule-based analysis, ML models, external services, etc.
type Analyzer interface {
	// Analyze examines test results and artifacts to provide insights.
	// Returns analysis results or nil if no analysis was performed.
	Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisResult, error)

	// ShouldAnalyze determines if analysis should run based on results.
	ShouldAnalyze(result *ExecutionResult) bool
}

// Reporter handles test result reporting.
// Implementations can generate multiple report formats (JUnit, HTML, Slack, etc.)
// using the composite pattern to combine multiple reporters.
type Reporter interface {
	// Initialize sets up reporting (create directories, validate config).
	Initialize(ctx context.Context) error

	// Report generates reports from execution results.
	Report(ctx context.Context, input *ReportInput) error

	// Finalize performs cleanup and final report generation.
	Finalize(ctx context.Context) error
}

