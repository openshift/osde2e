package config

import (
	"time"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	commonconfig "github.com/openshift/osde2e/pkg/common/config"
)

// Dashboard configuration keys
const (
	// Port is the HTTP port the dashboard server listens on
	Port = "dashboard.port"

	// Environment filters clusters by environment (stage, prod, integration, all)
	Environment = "dashboard.environment"

	// RefreshInterval is how often to refresh data (in seconds)
	RefreshInterval = "dashboard.refreshInterval"

	// ExpirationWarningThreshold is the duration before expiration to warn about
	ExpirationWarningThreshold = "dashboard.expirationWarningThreshold"

	// MaxTestResults is the maximum number of test results to return
	MaxTestResults = "dashboard.maxTestResults"

	// LookbackDays is the number of days of S3 data to scan for operator status
	LookbackDays = "dashboard.lookbackDays"

	// SQSQueueURL is the URL of the SQS queue receiving S3 ObjectCreated events
	SQSQueueURL = "dashboard.sqsQueueURL"

	// DBPath is the path to the SQLite database file
	DBPath = "dashboard.dbPath"
)

// Default values
const (
	DefaultPort                       = 8080
	DefaultEnvironment                = "all"
	DefaultRefreshInterval            = 300 // 5 minutes
	DefaultExpirationWarningThreshold = 2 * time.Hour
	DefaultMaxTestResults             = 100
	DefaultLookbackDays               = 30
)

// Config holds dashboard configuration
type Config struct {
	Port                       int
	S3Bucket                   string // Reuses commonconfig.Tests.LogBucket
	S3Region                   string // Reuses commonconfig.AWSRegion
	OCMConfigPath              string // Reuses commonconfig.OcmConfig
	Environment                string
	RefreshInterval            int
	ExpirationWarningThreshold time.Duration
	MaxTestResults             int
	LookbackDays               int
	SQSQueueURL                string // SQS queue URL for S3 event notifications
	DBPath                     string // Path to SQLite database file
}

// LoadConfig loads dashboard configuration from viper
// Reuses existing AWS and OCM configuration from common config
func LoadConfig() *Config {
	return &Config{
		Port:                       viper.GetInt(Port),
		S3Bucket:                   viper.GetString(commonconfig.Tests.LogBucket),
		S3Region:                   viper.GetString(commonconfig.AWSRegion),
		OCMConfigPath:              viper.GetString(commonconfig.OcmConfig),
		Environment:                viper.GetString(Environment),
		RefreshInterval:            viper.GetInt(RefreshInterval),
		ExpirationWarningThreshold: viper.GetDuration(ExpirationWarningThreshold),
		MaxTestResults:             viper.GetInt(MaxTestResults),
		LookbackDays:               viper.GetInt(LookbackDays),
		SQSQueueURL:                viper.GetString(SQSQueueURL),
		DBPath:                     viper.GetString(DBPath),
	}
}

// OCMEnvironments returns the list of OCM environments to query.
// "all" expands to stage + int + prod; a specific env returns just that one.
func (c *Config) OCMEnvironments() []string {
	switch c.Environment {
	case "all", "":
		return []string{"stage", "int", "prod"}
	default:
		return []string{c.Environment}
	}
}

// SetDefaults sets default configuration values
func SetDefaults() {
	viper.SetDefault(Port, DefaultPort)
	viper.SetDefault(Environment, DefaultEnvironment)
	viper.SetDefault(RefreshInterval, DefaultRefreshInterval)
	viper.SetDefault(ExpirationWarningThreshold, DefaultExpirationWarningThreshold)
	viper.SetDefault(MaxTestResults, DefaultMaxTestResults)
	viper.SetDefault(LookbackDays, DefaultLookbackDays)

	// Set defaults for S3 bucket if not already set
	if viper.GetString(commonconfig.Tests.LogBucket) == "" {
		viper.SetDefault(commonconfig.Tests.LogBucket, "osde2e-logs")
	}

	// The log bucket lives in us-east-1; fall back to that when no region is configured
	if viper.GetString(commonconfig.AWSRegion) == "" {
		viper.SetDefault(commonconfig.AWSRegion, "us-east-1")
	}
}
