package slack

// AnalysisResult represents the analysis output passed to reporters.
type AnalysisResult struct {
	Status   string         `json:"status"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Error    string         `json:"error,omitempty"`
	Prompt   string         `json:"prompt,omitempty"`
}

// ReporterConfig holds configuration for different reporter implementations
type ReporterConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// NotificationConfig holds configuration for notification settings
type NotificationConfig struct {
	Enabled   bool             `json:"enabled" yaml:"enabled"`
	Reporters []ReporterConfig `json:"reporters" yaml:"reporters"`
}
