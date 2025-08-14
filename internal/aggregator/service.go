// Package aggregator provides a unified interface for collecting artifacts
// and metadata from osde2e test runs for LLM analysis.
package aggregator

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

// Service provides a unified interface for collecting artifacts and metadata
type Service struct {
	artifactCollector *artifactCollector
	logger            logr.Logger
}

// NewService creates a new aggregator service
func NewService(logger logr.Logger) *Service {
	return &Service{
		artifactCollector: newArtifactCollector(logger),
		logger:            logger,
	}
}

// Collect collects all artifacts and metadata from the specified report directory
func (s *Service) Collect(ctx context.Context, reportDir string) (*AggregatedData, error) {
	s.logger.Info("collecting data", "reportDir", reportDir)

	// Collect artifacts
	data, err := s.artifactCollector.collectFromReportDir(reportDir)
	if err != nil {
		return nil, fmt.Errorf("failed to collect artifacts: %w", err)
	}

	// Extract metadata from CI artifacts using already collected log entries
	metadata := extractMetadata(data.Logs, s.logger)

	// Set the extracted metadata
	data.SetMetadata(metadata)

	s.logger.Info("completed data collection",
		"artifacts", len(data.Logs),
		"metadata_keys", len(metadata),
		"failed_tests", len(data.FailedTests))

	return data, nil
}
