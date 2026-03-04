// Package orchestrator defines the interface for managing end-to-end test lifecycles.
package orchestrator

import "context"

// Orchestrator manages the complete lifecycle of e2e test execution including
// cluster provisioning, test execution, failure analysis, and reporting.
type Orchestrator interface {
	// Provision prepares the test environment by provisioning or reusing a cluster,
	// loading kubeconfig, and installing required addons.
	Provision(ctx context.Context) error

	// Execute runs the configured test suites including install phase tests and
	// optional upgrade tests with their respective phases.
	Execute(ctx context.Context) error

	// AnalyzeLogs performs AI-powered log analysis when tests fail,
	// providing insights into failure root causes. Results are cached
	// internally for use by Report.
	AnalyzeLogs(ctx context.Context, testErr error) error

	// Report uploads artifacts, sends notifications, and generates reports.
	// It consolidates S3 uploads, Slack notifications (both built-in analysis
	// and deferred ad-hoc test suite results), and diagnostic data collection.
	Report(ctx context.Context) error

	// Cleanup performs post-test cleanup including resource cleanup and
	// optionally destroys the cluster based on configuration.
	Cleanup(ctx context.Context) error

	// PostProcessCluster performs optional post-processing on the cluster
	// after test execution but before cleanup (e.g., extending expiration,
	// updating metadata). Implementations can return nil to skip processing.
	PostProcessCluster(ctx context.Context) error

	// Result returns the outcome of the test run including exit code and status.
	Result() *Result
}

// Result encapsulates the outcome of an e2e test run.
type Result struct {
	ExitCode      int     // Exit code: 0 for success, non-zero for failure
	TestsPassed   bool    // Whether install phase tests passed
	UpgradePassed bool    // Whether upgrade phase tests passed (if run)
	ClusterID     string  // ID of the cluster used for testing
	Errors        []error // Collection of errors encountered during execution
}
