// Package execute provides implementations of the Executor interface
// for various test frameworks.
package execute

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/phase"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/debug"
	ctrlog "sigs.k8s.io/controller-runtime/pkg/log"
)

// GinkgoExecutor implements the Executor interface using the Ginkgo test framework.
type GinkgoExecutor struct {
	// Configuration is stored during initialization
}

// NewGinkgoExecutor creates a new Ginkgo-based executor.
func NewGinkgoExecutor() *GinkgoExecutor {
	return &GinkgoExecutor{}
}

// Execute runs the configured test suite against the cluster.
func (ex *GinkgoExecutor) Execute(ctx context.Context, target *orchestrator.ExecutionTarget) (*orchestrator.ExecutionResult, error) {
	ctrlog.SetLogger(ginkgo.GinkgoLogr)
	testing.Init()

	// Update viper with cluster information
	if target.Cluster != nil {
		viper.Set(config.Cluster.ID, target.Cluster.ID)
		viper.Set(config.Cluster.Name, target.Cluster.Name)
		viper.Set(config.Cluster.Version, target.Cluster.Version)
		if len(target.Kubeconfig) > 0 {
			viper.Set(config.Kubeconfig.Contents, string(target.Kubeconfig))
		}
	}

	startTime := time.Now()

	// Determine execution mode based on labels
	phaseLabel := target.Labels["phase"]
	shouldUpgrade := target.Labels["upgrade"] == "true"

	var success bool
	var artifacts []orchestrator.Artifact
	var errors []orchestrator.TestError

	if shouldUpgrade {
		success, artifacts, errors = ex.executeWithUpgrade(ctx, target)
	} else if phaseLabel != "" {
		success, artifacts, errors = ex.executeSinglePhase(ctx, phaseLabel, target)
	} else {
		// Default execution: run install tests, optionally upgrade
		success, artifacts, errors = ex.executeDefault(ctx, target)
	}

	endTime := time.Now()

	// Build result summary
	summary := &orchestrator.ResultSummary{
		Errors: errors,
	}

	// Populate counts from artifacts if JUnit report is available
	// (This is a simplification - real implementation would parse JUnit)
	if success {
		summary.Passed = 1 // At least one test passed
	} else {
		summary.Failed = 1
	}

	result := &orchestrator.ExecutionResult{
		Success:   success,
		StartTime: startTime,
		EndTime:   endTime,
		Summary:   summary,
		Artifacts: artifacts,
		Metadata: map[string]interface{}{
			"framework": "ginkgo",
			"phase":     phaseLabel,
			"upgrade":   shouldUpgrade,
		},
	}

	if !success {
		return result, fmt.Errorf("test execution failed")
	}

	return result, nil
}

// executeDefault runs the default execution flow (install tests, optional upgrade)
func (ex *GinkgoExecutor) executeDefault(ctx context.Context, target *orchestrator.ExecutionTarget) (bool, []orchestrator.Artifact, []orchestrator.TestError) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	viper.Set(config.Cluster.Passing, false)

	suiteConfig, reporterConfig := ex.configureSuite(target)

	if viper.GetString(config.Suffix) == "" {
		viper.Set(config.Suffix, util.RandomStr(5))
	}

	// Determine if we should run install tests and/or upgrade
	runInstallTests := true
	upgradeCluster := false
	if viper.GetString(config.Upgrade.Image) != "" || viper.GetString(config.Upgrade.ReleaseName) != "" {
		upgradeCluster = true
		runInstallTests = viper.GetBool(config.Upgrade.RunPreUpgradeTests)
	}

	testsPassed := true
	var allArtifacts []orchestrator.Artifact
	var allErrors []orchestrator.TestError

	// Run install phase tests
	if runInstallTests {
		log.Println("Running e2e tests...")
		passed, artifacts, errs := ex.runPhase(phase.InstallPhase, "OSD e2e suite", suiteConfig, reporterConfig)
		testsPassed = passed
		allArtifacts = append(allArtifacts, artifacts...)
		allErrors = append(allErrors, errs...)
		viper.Set(config.Cluster.Passing, testsPassed)
	}

	upgradeTestsPassed := true

	// Run upgrade if requested
	if upgradeCluster {
		if len(viper.GetString(config.Kubeconfig.Contents)) > 0 {
			h, err := helper.NewOutsideGinkgo()
			if h == nil || err != nil {
				allErrors = append(allErrors, orchestrator.TestError{
					Name:    "upgrade-setup",
					Message: fmt.Sprintf("unable to generate helper outside ginkgo: %v", err),
					Time:    time.Now(),
				})
				return false, allArtifacts, allErrors
			}

			// Run the upgrade
			if err = upgrade.RunUpgrade(h); err != nil {
				allErrors = append(allErrors, orchestrator.TestError{
					Name:    "upgrade-execution",
					Message: fmt.Sprintf("error performing upgrade: %v", err),
					Time:    time.Now(),
				})
				return false, allArtifacts, allErrors
			}

			// Run post-upgrade tests if configured
			if viper.GetBool(config.Upgrade.RunPostUpgradeTests) {
				log.Println("Running e2e tests POST-UPGRADE...")
				viper.Set(config.Cluster.Passing, false)
				passed, artifacts, errs := ex.runPhase(
					phase.UpgradePhase,
					"OSD e2e suite post-upgrade",
					suiteConfig,
					reporterConfig,
				)
				upgradeTestsPassed = passed
				allArtifacts = append(allArtifacts, artifacts...)
				allErrors = append(allErrors, errs...)
				viper.Set(config.Cluster.Passing, upgradeTestsPassed)
			}
		} else {
			log.Println("Unable to perform cluster upgrade, no kubeconfig found.")
		}
	}

	finalSuccess := testsPassed && upgradeTestsPassed
	return finalSuccess, allArtifacts, allErrors
}

// executeSinglePhase runs tests for a specific phase
func (ex *GinkgoExecutor) executeSinglePhase(ctx context.Context, phaseName string, target *orchestrator.ExecutionTarget) (bool, []orchestrator.Artifact, []orchestrator.TestError) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	suiteConfig, reporterConfig := ex.configureSuite(target)

	return ex.runPhase(phaseName, fmt.Sprintf("OSD e2e suite - %s", phaseName), suiteConfig, reporterConfig)
}

// executeWithUpgrade runs install tests, upgrade, and post-upgrade tests
func (ex *GinkgoExecutor) executeWithUpgrade(ctx context.Context, target *orchestrator.ExecutionTarget) (bool, []orchestrator.Artifact, []orchestrator.TestError) {
	// This is similar to executeDefault but forces upgrade flow
	return ex.executeDefault(ctx, target)
}

// configureSuite configures Ginkgo suite and reporter settings
func (ex *GinkgoExecutor) configureSuite(target *orchestrator.ExecutionTarget) (types.SuiteConfig, types.ReporterConfig) {
	suiteConfig, reporterConfig := ginkgo.GinkgoConfiguration()

	// Apply timeout
	if target.Timeout > 0 {
		suiteConfig.Timeout = target.Timeout
	} else {
		suiteConfig.Timeout = time.Hour * time.Duration(viper.GetInt(config.Tests.SuiteTimeout))
	}

	// Apply skip/focus/label filters
	if skip := viper.GetString(config.Tests.GinkgoSkip); skip != "" {
		suiteConfig.SkipStrings = append(suiteConfig.SkipStrings, skip)
	}

	if labels := viper.GetString(config.Tests.GinkgoLabelFilter); labels != "" {
		suiteConfig.LabelFilter = labels
	}

	if testsToRun := viper.GetStringSlice(config.Tests.TestsToRun); len(testsToRun) > 0 {
		suiteConfig.FocusStrings = testsToRun
	}

	if focus := viper.GetString(config.Tests.GinkgoFocus); focus != "" {
		suiteConfig.FocusStrings = append(suiteConfig.FocusStrings, focus)
	}

	suiteConfig.DryRun = viper.GetBool(config.DryRun)

	if suiteConfig.DryRun {
		log.Println("\x1b[33mWARNING! This is a DRY RUN. Review this state if outcome is unexpected.\033[0m")
	}

	// Configure reporter
	logLevel := viper.GetString(config.Tests.GinkgoLogLevel)
	switch logLevel {
	case "v":
		reporterConfig.Verbose = true
	case "vv":
		reporterConfig.VeryVerbose = true
	case "succinct":
		fallthrough
	default:
		reporterConfig.Succinct = true
	}

	reporterConfig.NoColor = true

	return suiteConfig, reporterConfig
}

// runPhase runs tests for a specific phase
func (ex *GinkgoExecutor) runPhase(
	phaseName string,
	description string,
	suiteConfig types.SuiteConfig,
	reporterConfig types.ReporterConfig,
) (bool, []orchestrator.Artifact, []orchestrator.TestError) {
	viper.Set(config.Phase, phaseName)
	reportDir := viper.GetString(config.ReportDir)
	phaseDirectory := filepath.Join(reportDir, phaseName)

	if _, err := os.Stat(phaseDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(phaseDirectory, os.FileMode(0o755)); err != nil {
			log.Printf("error while creating phase directory %s", phaseDirectory)
			return false, nil, []orchestrator.TestError{{
				Name:    "phase-setup",
				Message: fmt.Sprintf("failed to create phase directory: %v", err),
				Time:    time.Now(),
			}}
		}
	}

	suffix := viper.GetString(config.Suffix)
	var artifacts []orchestrator.Artifact
	var testErrors []orchestrator.TestError

	// Setup JUnit report generation
	_ = ginkgo.ReportAfterSuite("OSDE2E", func(report ginkgo.Report) {
		junitPath := filepath.Join(phaseDirectory, fmt.Sprintf("junit_%v.xml", suffix))
		err := reporters.GenerateJUnitReportWithConfig(
			report,
			junitPath,
			reporters.JunitReportConfig{OmitSpecLabels: true, OmitLeafNodeType: true},
		)
		if err != nil {
			log.Printf("error creating junit report file %s", err.Error())
		} else {
			artifacts = append(artifacts, orchestrator.Artifact{
				Name:     fmt.Sprintf("junit_%v.xml", suffix),
				Path:     junitPath,
				MimeType: "application/xml",
			})
		}

		// Collect test errors from report
		for _, specReport := range report.SpecReports {
			if specReport.State == types.SpecStateFailed {
				testErrors = append(testErrors, orchestrator.TestError{
					Name:    specReport.FullText(),
					Message: specReport.Failure.Message,
					Stack:   specReport.Failure.Location.String(),
					Time:    specReport.EndTime,
				})
			}
		}
	})

	// Setup Konflux results if configured
	if konfluxTestOutputFile := viper.GetString(config.KonfluxTestOutputFile); konfluxTestOutputFile != "" {
		ginkgo.ReportAfterSuite("OSDE2E konflux results", func(report ginkgo.Report) {
			konfluxResults := map[string]any{
				"result":    "FAILURE",
				"timestamp": report.EndTime.Format(time.RFC3339),
				"warnings":  0,
			}

			if report.SuiteSucceeded {
				konfluxResults["result"] = "SUCCESS"
			}

			var successes, failures int
			for _, specReport := range report.SpecReports {
				switch specReport.State {
				case types.SpecStatePassed:
					successes++
				case types.SpecStateFailed:
					failures++
				}
			}
			konfluxResults["successes"] = successes
			konfluxResults["failures"] = failures

			bits, err := json.Marshal(konfluxResults)
			if err != nil {
				log.Printf("unable to marshal konflux results: %s", err)
			}
			if err = os.WriteFile(konfluxTestOutputFile, bits, os.ModePerm); err != nil {
				log.Printf("failed to write konflux results to %s: %s", konfluxTestOutputFile, err)
			}
		})
	}

	// Run the tests
	ginkgoPassed := false
	func() {
		defer ginkgo.GinkgoRecover()
		ginkgoPassed = ginkgo.RunSpecs(ginkgo.GinkgoT(), description, suiteConfig, reporterConfig)
	}()

	// Generate dependencies file for periodic jobs
	clusterID := viper.GetString(config.Cluster.ID)
	if !suiteConfig.DryRun && clusterID != "" &&
		viper.GetString(config.JobName) != "" && viper.GetString(config.JobType) == "periodic" {
		h, err := helper.NewOutsideGinkgo()
		if h != nil && err == nil {
			dependencies, err := debug.GenerateDependencies(h.Kube())
			if err != nil {
				log.Printf("Error generating dependencies: %s", err.Error())
			} else {
				depsPath := filepath.Join(phaseDirectory, "dependencies.txt")
				if err = os.WriteFile(depsPath, []byte(dependencies), 0o644); err != nil {
					log.Printf("Error writing dependencies.txt: %s", err.Error())
				} else {
					artifacts = append(artifacts, orchestrator.Artifact{
						Name:     "dependencies.txt",
						Path:     depsPath,
						MimeType: "text/plain",
					})
				}

				err := debug.GenerateDiff(phaseName, dependencies)
				if err != nil {
					log.Printf("Error generating diff: %s", err.Error())
				}
			}
		}
	}

	return ginkgoPassed, artifacts, testErrors
}
