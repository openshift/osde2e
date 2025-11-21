package reporter

import (
	"context"
	"fmt"
)

// AnalysisResult represents the analysis output (simplified to avoid import cycle)
type AnalysisResult struct {
	Status   string         `json:"status"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Error    string         `json:"error,omitempty"`
	Prompt   string         `json:"prompt,omitempty"`
}

// Reporter interface defines the contract for sending analysis results to external systems
type Reporter interface {
	// Report sends the analysis result to the configured destination
	Report(ctx context.Context, result *AnalysisResult, config *ReporterConfig) error

	// Name returns the reporter's identifier
	Name() string
}

// ReporterConfig holds configuration for different reporter implementations
type ReporterConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// ReporterRegistry manages available reporters
type ReporterRegistry struct {
	reporters map[string]Reporter
}

// NewReporterRegistry creates a new reporter registry
func NewReporterRegistry() *ReporterRegistry {
	return &ReporterRegistry{
		reporters: make(map[string]Reporter),
	}
}

// Register adds a reporter to the registry
func (r *ReporterRegistry) Register(reporter Reporter) {
	r.reporters[reporter.Name()] = reporter
}

// Get retrieves a reporter by name
func (r *ReporterRegistry) Get(name string) (Reporter, bool) {
	reporter, exists := r.reporters[name]
	return reporter, exists
}

// List returns all registered reporter names
func (r *ReporterRegistry) List() []string {
	names := make([]string, 0, len(r.reporters))
	for name := range r.reporters {
		names = append(names, name)
	}
	return names
}

// SendNotification sends analysis result to a specific reporter
func (r *ReporterRegistry) SendNotification(ctx context.Context, result *AnalysisResult, config *ReporterConfig) error {
	if !config.Enabled {
		return nil // Skip disabled reporters
	}

	reporter, exists := r.Get(config.Type)
	if !exists {
		return fmt.Errorf("reporter type %s not found", config.Type)
	}

	return reporter.Report(ctx, result, config)
}

// NotificationConfig holds configuration for notification settings
type NotificationConfig struct {
	Enabled   bool             `json:"enabled" yaml:"enabled"`
	Reporters []ReporterConfig `json:"reporters" yaml:"reporters"`
}
