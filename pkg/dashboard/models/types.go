package models

import "time"

// ClusterReserve represents a reserved cluster available for testing
type ClusterReserve struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	State         string            `json:"state"`        // ready, installing, pending
	Availability  string            `json:"availability"` // reserved, claimed, used
	Version       string            `json:"version"`
	Region        string            `json:"region"`
	CloudProvider string            `json:"cloud_provider"`
	CreatedAt     time.Time         `json:"created_at"`
	ExpiresAt     time.Time         `json:"expires_at"`
	Product       string            `json:"product"` // osd, rosa
	Properties    map[string]string `json:"properties,omitempty"`
}

// IsExpiringSoon returns true if the cluster expires within the given duration.
// Returns false for zero or already-expired timestamps.
func (c *ClusterReserve) IsExpiringSoon(threshold time.Duration) bool {
	if c.ExpiresAt.IsZero() {
		return false
	}
	remaining := time.Until(c.ExpiresAt)
	return remaining >= 0 && remaining < threshold
}

// ExpiringSoon returns true if the cluster expires within 2 hours (for template use)
func (c *ClusterReserve) ExpiringSoon() bool {
	return !c.ExpiresAt.IsZero() && time.Until(c.ExpiresAt) < 2*time.Hour
}

// ClusterUsage represents aggregate cluster usage metrics
type ClusterUsage struct {
	Environment     string         `json:"environment"` // stage, prod, integration
	TotalClusters   int            `json:"total_clusters"`
	ByState         map[string]int `json:"by_state"`        // ready: 5, installing: 2
	ByAvailability  map[string]int `json:"by_availability"` // reserved: 3, claimed: 2, used: 1
	ByCloudProvider map[string]int `json:"by_cloud_provider,omitempty"`
	ByVersion       map[string]int `json:"by_version,omitempty"`
	LastUpdated     time.Time      `json:"last_updated"`
}

// TestCase holds a single test case result for rendering in the UI
type TestCase struct {
	Name     string  `json:"name"`
	Duration float64 `json:"duration_seconds"`
	Status   string  `json:"status"`            // passed, failed, error, skipped
	Message  string  `json:"message,omitempty"` // failure/error/skip message
}

// TestResult represents the outcome of a test execution
type TestResult struct {
	JobID        string     `json:"job_id"`
	JobName      string     `json:"job_name"`
	Component    string     `json:"component"`
	Date         string     `json:"date"`
	Status       string     `json:"status"` // passed, failed, error, skipped
	TotalTests   int        `json:"total_tests"`
	PassedTests  int        `json:"passed_tests"`
	FailedTests  int        `json:"failed_tests"`
	SkippedTests int        `json:"skipped_tests"`
	ErrorTests   int        `json:"error_tests"`
	Duration     float64    `json:"duration_seconds"`
	S3Path       string     `json:"s3_path"`
	LogURL       string     `json:"log_url,omitempty"`
	JUnitXMLURL  string     `json:"junit_xml_url,omitempty"`
	Timestamp    time.Time  `json:"timestamp"`
	TestCases    []TestCase `json:"test_cases,omitempty"`
}

// DashboardOverview provides a high-level summary for the main dashboard view
type DashboardOverview struct {
	TotalReservedClusters int            `json:"total_reserved_clusters"`
	ClustersExpiringSoon  int            `json:"clusters_expiring_soon"`
	ActiveTests           int            `json:"active_tests"`
	OverallSuccessRate    float64        `json:"overall_success_rate"`
	RecentTests           []TestResult   `json:"recent_tests"`
	ClusterUsageSummary   []ClusterUsage `json:"cluster_usage_summary"`
	LastUpdated           time.Time      `json:"last_updated"`
}

// APIResponse is a generic wrapper for API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// FailedTestCase holds the name and failure message of a single failed test
type FailedTestCase struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// LLMAnalysis holds the AI-generated root cause and recommendations from summary.yaml
type LLMAnalysis struct {
	RootCause       string   `json:"root_cause"`
	Recommendations []string `json:"recommendations"`
}

// EnvironmentResult holds the latest test result for one operator+version in one environment
type EnvironmentResult struct {
	Status      string           `json:"status"` // passed, failed, error
	Version     string           `json:"version"`
	Total       int              `json:"total"`
	Passed      int              `json:"passed"`
	Failed      int              `json:"failed"`
	Skipped     int              `json:"skipped"`
	Errors      int              `json:"errors"`
	LastRun     time.Time        `json:"last_run"`
	JobID       string           `json:"job_id"`
	LogURL      string           `json:"log_url,omitempty"`
	JUnitURL    string           `json:"junit_url,omitempty"`
	FailedTests []FailedTestCase `json:"failed_tests,omitempty"`
	LLMAnalysis *LLMAnalysis     `json:"llm_analysis,omitempty"`
}

// DeliverableStatus represents the cross-environment test status for one operator+version
type DeliverableStatus struct {
	Name        string                        `json:"name"`
	Version     string                        `json:"version"`
	Results     map[string]*EnvironmentResult `json:"results"` // key: "stage", "prod", "integration", "unknown"
	LastUpdated time.Time                     `json:"last_updated"`
}

// Stage returns the result for the stage environment, or nil if not available.
func (o DeliverableStatus) Stage() *EnvironmentResult { return o.Results["stage"] }

// Prod returns the result for the prod environment, or nil if not available.
func (o DeliverableStatus) Prod() *EnvironmentResult { return o.Results["prod"] }

// Integration returns the result for the integration environment.
// Checks both "int" (stored by SQS consumer) and "integration" (legacy).
func (o DeliverableStatus) Integration() *EnvironmentResult {
	if r := o.Results["int"]; r != nil {
		return r
	}
	return o.Results["integration"]
}

// Unknown returns results from runs where the environment could not be determined.
func (o DeliverableStatus) Unknown() *EnvironmentResult { return o.Results["unknown"] }

// PipelineRun represents one test run of an operator version in one environment
type PipelineRun struct {
	Version     string           `json:"version"`
	Env         string           `json:"env"` // stage, int, prod
	Status      string           `json:"status"`
	Date        string           `json:"date"`
	JobID       string           `json:"job_id"`
	LastRun     time.Time        `json:"last_run"`
	LogURL      string           `json:"log_url,omitempty"`
	JUnitURL    string           `json:"junit_url,omitempty"`
	Failed      []FailedTestCase `json:"failed_tests,omitempty"`
	Total       int              `json:"total"`
	Passed      int              `json:"passed"`
	LLMAnalysis *LLMAnalysis     `json:"llm_analysis,omitempty"`
}

// PipelineHistory holds all historical runs for a single operator, grouped by version
type PipelineHistory struct {
	Name     string            `json:"name"`
	Runs     []PipelineRun     `json:"runs"`     // sorted newest first (flat)
	Versions []VersionPipeline `json:"versions"` // grouped by version, newest first
}

// VersionPipeline represents one version of an operator and its run results per env
type VersionPipeline struct {
	Version string                  `json:"version"`
	Date    string                  `json:"date"`     // date of the most recent run
	LastRun time.Time               `json:"last_run"` // timestamp of most recent run
	EnvRuns map[string]*PipelineRun `json:"env_runs"` // keyed by env: "int", "stage", "prod"
}

// FailureEntry is one deliverable+version+env that shares a common failure root cause
type FailureEntry struct {
	Name    string    `json:"name"`
	Version string    `json:"version"`
	Env     string    `json:"env"`
	LastRun time.Time `json:"last_run"`
	JobID   string    `json:"job_id"`
	LogURL  string    `json:"log_url,omitempty"`
}

// FailureGroup groups deliverables that share a similar failure summary
type FailureGroup struct {
	FailureMatch    string         `json:"failure_match"`   // first sentence of LLM root cause or failure message — the grouping key
	RootCause       string         `json:"root_cause"`      // full LLM root cause (from most recent entry with analysis)
	Recommendations []string       `json:"recommendations"` // LLM recommendations
	Entries         []FailureEntry `json:"entries"`         // sorted newest first
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status       string    `json:"status"` // ok, degraded, error
	Version      string    `json:"version,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	OCMConnected bool      `json:"ocm_connected"`
	S3Connected  bool      `json:"s3_connected"`
}
