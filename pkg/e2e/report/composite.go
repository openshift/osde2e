// Package report provides implementations of the Reporter interface
// for various reporting formats and destinations.
package report

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/runner"
)

// CompositeReporter combines multiple reporters into a single reporter.
// It delegates to all sub-reporters, continuing even if some fail.
type CompositeReporter struct {
	reporters []orchestrator.Reporter
	reportDir string
}

// NewCompositeReporter creates a new composite reporter.
func NewCompositeReporter() *CompositeReporter {
	return &CompositeReporter{
		reporters: []orchestrator.Reporter{},
		reportDir: viper.GetString(config.ReportDir),
	}
}

// AddReporter adds a sub-reporter to the composite.
func (r *CompositeReporter) AddReporter(reporter orchestrator.Reporter) {
	r.reporters = append(r.reporters, reporter)
}

// Initialize sets up reporting (create directories, validate config).
func (r *CompositeReporter) Initialize(ctx context.Context) error {
	// Create report directory if it doesn't exist
	if r.reportDir == "" {
		r.reportDir = "./reports"
		viper.Set(config.ReportDir, r.reportDir)
	}

	if _, err := os.Stat(r.reportDir); os.IsNotExist(err) {
		if err := os.MkdirAll(r.reportDir, os.FileMode(0o755)); err != nil {
			return fmt.Errorf("failed to create report directory: %w", err)
		}
	}

	log.Printf("Report directory initialized: %s", r.reportDir)

	// Initialize all sub-reporters
	for _, reporter := range r.reporters {
		if err := reporter.Initialize(ctx); err != nil {
			log.Printf("Warning: Failed to initialize reporter: %v", err)
			// Continue with other reporters
		}
	}

	return nil
}

// Report generates reports from execution results.
func (r *CompositeReporter) Report(ctx context.Context, input *orchestrator.ReportInput) error {
	var firstError error

	// Report to all sub-reporters
	for _, reporter := range r.reporters {
		if err := reporter.Report(ctx, input); err != nil {
			log.Printf("Warning: Reporter failed: %v", err)
			if firstError == nil {
				firstError = err
			}
			// Continue with other reporters
		}
	}

	return firstError
}

// Finalize performs cleanup and final report generation.
func (r *CompositeReporter) Finalize(ctx context.Context) error {
	var firstError error

	// Finalize all sub-reporters
	for _, reporter := range r.reporters {
		if err := reporter.Finalize(ctx); err != nil {
			log.Printf("Warning: Failed to finalize reporter: %v", err)
			if firstError == nil {
				firstError = err
			}
		}
	}

	return firstError
}

// ArtifactCollector collects cluster artifacts (logs, must-gather, etc.)
type ArtifactCollector struct {
	reportDir string
}

// NewArtifactCollector creates a new artifact collector.
func NewArtifactCollector() *ArtifactCollector {
	return &ArtifactCollector{
		reportDir: viper.GetString(config.ReportDir),
	}
}

// Initialize sets up artifact collection.
func (c *ArtifactCollector) Initialize(ctx context.Context) error {
	return nil
}

// Report collects artifacts for the given execution.
func (c *ArtifactCollector) Report(ctx context.Context, input *orchestrator.ReportInput) error {
	// Collect cluster logs
	if err := c.collectLogs(input); err != nil {
		log.Printf("Warning: Failed to collect logs: %v", err)
	}

	// Run must-gather if not skipped
	if !viper.GetBool(config.SkipMustGather) {
		if err := c.runMustGather(ctx, input); err != nil {
			log.Printf("Warning: Must-gather failed: %v", err)
		}
	}

	return nil
}

// Finalize performs cleanup.
func (c *ArtifactCollector) Finalize(ctx context.Context) error {
	return nil
}

// collectLogs collects cluster logs from viper (populated by provisioner)
func (c *ArtifactCollector) collectLogs(input *orchestrator.ReportInput) error {
	// Check if logs are stored in viper
	logsMap := viper.AllSettings()
	for key, value := range logsMap {
		if len(key) > 5 && key[:5] == "logs." {
			logName := key[5:]
			if logBytes, ok := value.([]byte); ok {
				logPath := filepath.Join(c.reportDir, fmt.Sprintf("%s-log.txt", logName))
				if err := os.WriteFile(logPath, logBytes, os.ModePerm); err != nil {
					log.Printf("Error writing log %s: %s", logPath, err.Error())
				}
			}
		}
	}

	return nil
}

// runMustGather runs must-gather to collect cluster state
func (c *ArtifactCollector) runMustGather(ctx context.Context, input *orchestrator.ReportInput) error {
	log.Print("Running Must Gather...")
	
	h, err := helper.NewOutsideGinkgo()
	if h == nil || err != nil {
		return fmt.Errorf("unable to generate helper for must-gather: %w", err)
	}

	mustGatherTimeoutInSeconds := 1800
	h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
	r := h.Runner(fmt.Sprintf("oc adm must-gather --dest-dir=%v", runner.DefaultRunner.OutputDir))
	r.Name = "must-gather"
	r.Tarball = true
	stopCh := make(chan struct{})
	
	if err := r.Run(mustGatherTimeoutInSeconds, stopCh); err != nil {
		return fmt.Errorf("error running must-gather: %w", err)
	}

	gatherResults, err := r.RetrieveResults()
	if err != nil {
		return fmt.Errorf("error retrieving must-gather results: %w", err)
	}

	h.WriteResults(gatherResults)

	log.Print("Gathering Project States...")
	h.InspectState(ctx)

	log.Print("Gathering OLM State...")
	if err = h.InspectOLM(ctx); err != nil {
		log.Printf("Error inspecting OLM: %v", err)
	}

	return nil
}

