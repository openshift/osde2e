package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/onsi/gomega"
	"github.com/openshift/osde2e/internal/analysisengine"
	"github.com/openshift/osde2e/internal/reporter"
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/cluster"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/phase"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/debug"
	ctrlog "sigs.k8s.io/controller-runtime/pkg/log"
)

// RunTests initializes the orchestrator and runs the complete e2e test lifecycle.
// This includes provisioning, test execution, log analysis (on failure), and reporting.
func RunTests(ctx context.Context) int {
	// Create orchestrator
	orch, err := NewOrchestrator(ctx)
	if err != nil {
		log.Printf("Failed to create orchestrator: %v", err)
		return config.Failure
	}

	// Provision cluster
	if err := orch.Provision(ctx); err != nil {
		log.Printf("Provision failed: %v", err)
		if viper.GetBool(config.LogAnalysis.EnableAnalysis) {
			if err := orch.AnalyzeLogs(ctx, err); err != nil {
				log.Printf("Log analysis failed: %v", err)
			}
		}
		return config.Failure
	}

	// Execute tests
	testErr := orch.Execute(ctx)

	// Analyze logs on failure, if enabled
	if testErr != nil {
		log.Printf("Tests failed: %v", testErr)
		// Run log analysis only if enableAnalysis is true and TestSuites and AdHocTestImages are not set
		// This is to avoid double analysis when adhocTestImages or TestSuites are used
		if viper.GetBool(config.LogAnalysis.EnableAnalysis) && viper.GetString(config.Tests.TestSuites) == "" && viper.GetString(config.Tests.AdHocTestImages) == "" {
			if err := orch.AnalyzeLogs(ctx, testErr); err != nil {
				log.Printf("Log analysis failed: %v", err)
			}
		}
	}

	// Generate reports
	if err := orch.Report(ctx); err != nil {
		log.Printf("Report errors: %v", err)
	}

	// Post-process cluster
	if err := orch.PostProcessCluster(ctx); err != nil {
		log.Printf("Cluster post-processing errors: %v", err)
	}

	// Cleanup resources and delete cluster
	if err := orch.Cleanup(ctx); err != nil {
		log.Printf("Cleanup errors: %v", err)
	}

	result := orch.Result()
	return result.ExitCode
}

// E2EOrchestrator implements the orchestrator.Orchestrator interface for OSD e2e tests.
type E2EOrchestrator struct {
	provider       spi.Provider
	result         *orchestrator.Result
	suiteConfig    types.SuiteConfig
	reporterConfig types.ReporterConfig
}

// NewOrchestrator creates a new E2E orchestrator instance.
func NewOrchestrator(ctx context.Context) (orchestrator.Orchestrator, error) {
	testing.Init()

	provider, err := providers.ClusterProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster provider: %w", err)
	}

	suiteConfig, reporterConfig := ginkgo.GinkgoConfiguration()
	configureGinkgo(&suiteConfig, &reporterConfig)

	return &E2EOrchestrator{
		provider:       provider,
		suiteConfig:    suiteConfig,
		reporterConfig: reporterConfig,
		result: &orchestrator.Result{
			ExitCode: config.Success,
		},
	}, nil
}

// Provision prepares the cluster environment.
func (o *E2EOrchestrator) Provision(ctx context.Context) error {
	ctrlog.SetLogger(ginkgo.GinkgoLogr)

	// Load cluster context (kubeconfig and cluster ID)
	if err := cluster.LoadClusterContext(); err != nil {
		return err
	}

	// Provision or reuse cluster
	cl, err := cluster.ProvisionOrReuseCluster(o.provider)
	runner.ReportClusterInstallLogs(o.provider)
	if err != nil {
		return err
	}

	o.result.ClusterID = cl.ID()

	// Install addons if configured
	if _, err := cluster.InstallAddonsIfConfigured(o.provider, cl.ID()); err != nil {
		runner.ReportClusterInstallLogs(o.provider)
		return fmt.Errorf("addon installation failed: %w", err)
	}

	return nil
}

// Execute runs the test suites.
func (o *E2EOrchestrator) Execute(ctx context.Context) error {
	gomega.RegisterFailHandler(ginkgo.Fail)
	viper.Set(config.Cluster.Passing, false)

	if viper.GetString(config.Suffix) == "" {
		viper.Set(config.Suffix, util.RandomStr(5))
	}

	// Determine test execution plan
	runInstallTests := true
	upgradeCluster := viper.GetString(config.Upgrade.Image) != "" || viper.GetString(config.Upgrade.ReleaseName) != ""

	if upgradeCluster {
		runInstallTests = viper.GetBool(config.Upgrade.RunPreUpgradeTests)
	}

	// Run install phase tests
	if runInstallTests {
		log.Println("Running e2e tests...")
		o.result.TestsPassed = o.runTestsInPhase(phase.InstallPhase, "OSD e2e suite")
		runner.ReportClusterInstallLogs(o.provider)
		viper.Set(config.Cluster.Passing, o.result.TestsPassed)
	}

	// Run upgrade and post-upgrade tests
	o.result.UpgradePassed = true
	if upgradeCluster {
		if err := o.runUpgrade(); err != nil {
			o.result.Errors = append(o.result.Errors, err)
			o.result.UpgradePassed = false
		}
	}

	// Set final result
	if !o.result.TestsPassed || !o.result.UpgradePassed {
		o.result.ExitCode = config.Failure
		return fmt.Errorf("tests failed")
	}

	return nil
}

// AnalyzeLogs performs AI-powered log analysis on test failures.
func (o *E2EOrchestrator) AnalyzeLogs(ctx context.Context, testErr error) error {
	log.Println("Running log analysis...")
	var notificationConfig *reporter.NotificationConfig
	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		return fmt.Errorf("no report directory available for log analysis")
	}

	clusterInfo := &analysisengine.ClusterInfo{
		ID:            viper.GetString(config.Cluster.ID),
		Name:          viper.GetString(config.Cluster.Name),
		Provider:      viper.GetString(config.Provider),
		Region:        viper.GetString(config.CloudProvider.Region),
		CloudProvider: viper.GetString(config.CloudProvider.CloudProviderID),
		Version:       viper.GetString(config.Cluster.Version),
	}
	if viper.GetBool(config.Tests.EnableSlackNotify) {
		notificationConfig = reporter.BuildNotificationConfig(
			viper.GetString(config.LogAnalysis.SlackWebhook),
			viper.GetString(config.LogAnalysis.SlackChannel),
			clusterInfo,
			reportDir,
		)
	}

	engineConfig := &analysisengine.Config{
		BaseConfig: analysisengine.BaseConfig{
			ArtifactsDir:       reportDir,
			APIKey:             viper.GetString(config.LogAnalysis.APIKey),
			NotificationConfig: notificationConfig,
		},
		PromptTemplate: "default",
		FailureContext: testErr.Error(),
		ClusterInfo:    clusterInfo,
	}

	engine, err := analysisengine.New(ctx, engineConfig)
	if err != nil {
		return fmt.Errorf("failed to create analysis engine: %w", err)
	}

	result, err := engine.Run(ctx)
	if err != nil {
		return fmt.Errorf("log analysis failed: %w", err)
	}

	log.Printf("Log analysis completed. Results: %s/%s/", reportDir, analysisengine.AnalysisDirName)
	log.Printf("=== Log Analysis Result ===\n%s", result.Content)

	return nil
}

// Report generates reports and collects diagnostic data.
func (o *E2EOrchestrator) Report(ctx context.Context) error {
	if o.suiteConfig.DryRun {
		return nil
	}

	runner.ReportClusterInstallLogs(o.provider)

	// Upload test artifacts to S3 if bucket configured
	if err := o.uploadToS3(); err != nil {
		log.Printf("S3 upload failed: %v", err)
		// Don't fail the overall report phase for S3 upload errors
	}

	return nil
}

// uploadToS3 uploads the report directory contents to S3.
func (o *E2EOrchestrator) uploadToS3() error {
	// Check if S3 bucket is configured
	if viper.GetString(config.Tests.LogBucket) == "" {
		return nil // S3 upload not configured, skip
	}

	component := deriveComponentFromTestImage()
	uploader, err := aws.NewS3Uploader(component)
	if err != nil {
		return fmt.Errorf("failed to create S3 uploader: %w", err)
	}

	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		return fmt.Errorf("no report directory configured")
	}

	results, err := uploader.UploadDirectory(reportDir)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	aws.LogS3UploadSummary(results)
	return nil
}

// deriveComponentFromTestImage determines the component name from the test image.
// It extracts a meaningful name from the test image path to organize S3 artifacts.
// Examples:
//
//	quay.io/org/osd-example-operator-e2e:tag -> osd-example-operator
//	quay.io/org/my-service-test:latest -> my-service
func deriveComponentFromTestImage() string {
	testSuites, err := config.GetTestSuites()
	if err == nil && len(testSuites) > 0 {
		imageName := testSuites[0].Image
		if component := extractNameFromImage(imageName); component != "" {
			log.Printf("Derived component from test image: %s -> %s", imageName, component)
			return component
		}
	}

	log.Println("Could not derive component, using fallback: unknown")
	return "unknown"
}

// extractNameFromImage extracts a meaningful name from a container image path.
// It strips the registry, organization, tag, and common test suffixes.
// Examples:
//
//	quay.io/org/osd-example-operator-e2e:tag -> osd-example-operator
//	quay.io/org/my-service-test:latest -> my-service
//	quay.io/org/simple:v1 -> simple
func extractNameFromImage(image string) string {
	if image == "" {
		return ""
	}

	// Remove tag (everything after :)
	if idx := strings.LastIndex(image, ":"); idx != -1 {
		image = image[:idx]
	}

	// Remove registry and org (everything before last /)
	if idx := strings.LastIndex(image, "/"); idx != -1 {
		image = image[idx+1:]
	}

	// Strip common test suffixes
	suffixes := []string{"-e2e", "-test", "-tests", "-harness"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(image, suffix) {
			image = strings.TrimSuffix(image, suffix)
			break
		}
	}

	return image
}

// PostProcessCluster performs post-processing on the cluster including must-gather,
// cluster expiration extension, and cluster property updates.
func (o *E2EOrchestrator) PostProcessCluster(ctx context.Context) error {
	if o.suiteConfig.DryRun {
		return nil
	}

	h, err := helper.NewOutsideGinkgo()
	if h == nil {
		log.Printf("Failed to generate helper object for post-processing: %v", err)
		return fmt.Errorf("failed to generate helper for post-processing: %w", err)
	}

	// Post processing: must gather, cluster extension, cluster property updates (with ginkgo defer for recovery)
	defer ginkgo.GinkgoRecover()

	var errors []error
	clusterStatus := "completed-failing"

	// Run must-gather
	if !viper.GetBool(config.SkipMustGather) {
		if err := cluster.RunMustGather(ctx, h); err != nil {
			errors = append(errors, err)
			clusterStatus = "completed-error"
		}
		cluster.InspectClusterState(ctx, h)
	}

	// Update cluster properties
	if clusterID := viper.GetString(config.Cluster.ID); clusterID != "" {
		if err := cluster.UpdateClusterProperties(o.provider, clusterStatus); err != nil {
			errors = append(errors, err)
		}
	}

	// Extend expiration for nightly builds
	if err := cluster.HandleExpirationExtension(o.provider); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		o.result.Errors = append(o.result.Errors, errors...)
	}

	return nil
}

// Cleanup performs post-test cleanup and optionally destroys the cluster.
func (o *E2EOrchestrator) Cleanup(ctx context.Context) error {
	if o.suiteConfig.DryRun {
		return nil
	}

	// Delete cluster if configured
	if err := cluster.DeleteCluster(o.provider); err != nil {
		o.result.Errors = append(o.result.Errors, err)
		return err
	}

	return nil
}

// Result returns the test execution result.
func (o *E2EOrchestrator) Result() *orchestrator.Result {
	return o.result
}

// runTestsInPhase executes tests for a specific phase.
func (o *E2EOrchestrator) runTestsInPhase(phaseName, description string) bool {
	viper.Set(config.Phase, phaseName)

	reportDir := viper.GetString(config.ReportDir)
	phaseDir := filepath.Join(reportDir, phaseName)
	if err := os.MkdirAll(phaseDir, 0o755); err != nil {
		log.Printf("Error creating phase directory %s: %v", phaseDir, err)
		return false
	}

	suffix := viper.GetString(config.Suffix)

	// Setup JUnit reporter
	ginkgo.ReportAfterSuite("OSDE2E", func(report ginkgo.Report) {
		junitPath := filepath.Join(phaseDir, fmt.Sprintf("junit_%v.xml", suffix))
		err := reporters.GenerateJUnitReportWithConfig(
			report,
			junitPath,
			reporters.JunitReportConfig{OmitSpecLabels: true, OmitLeafNodeType: true},
		)
		if err != nil {
			log.Printf("Error creating junit report: %v", err)
		}
	})

	// Setup Konflux reporter if configured
	if konfluxFile := viper.GetString(config.KonfluxTestOutputFile); konfluxFile != "" {
		ginkgo.ReportAfterSuite("OSDE2E konflux results", func(report ginkgo.Report) {
			o.writeKonfluxReport(report, konfluxFile)
		})
	}

	// Run tests
	var passed bool
	func() {
		defer ginkgo.GinkgoRecover()
		passed = ginkgo.RunSpecs(ginkgo.GinkgoT(), description, o.suiteConfig, o.reporterConfig)
	}()

	// Generate dependencies for periodic jobs
	o.generateDependencies(phaseDir, phaseName)

	return passed
}

// runUpgrade performs cluster upgrade and runs post-upgrade tests.
func (o *E2EOrchestrator) runUpgrade() error {
	if viper.GetString(config.Kubeconfig.Contents) == "" {
		return fmt.Errorf("unable to perform upgrade: no kubeconfig found")
	}

	h, err := helper.NewOutsideGinkgo()
	if err != nil {
		return fmt.Errorf("failed to generate helper for upgrade: %w", err)
	}

	if err := upgrade.RunUpgrade(h); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}

	if viper.GetBool(config.Upgrade.RunPostUpgradeTests) {
		log.Println("Running e2e tests POST-UPGRADE...")
		viper.Set(config.Cluster.Passing, false)
		o.result.UpgradePassed = o.runTestsInPhase(phase.UpgradePhase, "OSD e2e suite post-upgrade")
		viper.Set(config.Cluster.Passing, o.result.UpgradePassed)
	}

	return nil
}

// generateDependencies creates dependency reports for periodic jobs.
func (o *E2EOrchestrator) generateDependencies(phaseDir, phaseName string) {
	if o.suiteConfig.DryRun || viper.GetString(config.JobName) == "" || viper.GetString(config.JobType) != "periodic" {
		return
	}

	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" {
		return
	}

	cluster, err := o.provider.GetCluster(clusterID)
	if err != nil || cluster.State() != spi.ClusterStateReady {
		return
	}

	h, err := helper.NewOutsideGinkgo()
	if err != nil {
		return
	}

	dependencies, err := debug.GenerateDependencies(h.Kube())
	if err != nil {
		log.Printf("Error generating dependencies: %v", err)
		return
	}

	depFile := filepath.Join(phaseDir, "dependencies.txt")
	if err := os.WriteFile(depFile, []byte(dependencies), 0o644); err != nil {
		log.Printf("Error writing dependencies: %v", err)
		return
	}

	if err := debug.GenerateDiff(phaseName, dependencies); err != nil {
		log.Printf("Error generating diff: %v", err)
	}
}

// writeKonfluxReport generates Konflux-compatible test results.
func (o *E2EOrchestrator) writeKonfluxReport(report ginkgo.Report, outputFile string) {
	result := "FAILURE"
	if report.SuiteSucceeded {
		result = "SUCCESS"
	}

	var successes, failures int
	for _, spec := range report.SpecReports {
		switch spec.State {
		case types.SpecStatePassed:
			successes++
		case types.SpecStateFailed:
			failures++
		}
	}

	results := map[string]any{
		"result":    result,
		"timestamp": report.EndTime.Format(time.RFC3339),
		"warnings":  0,
		"successes": successes,
		"failures":  failures,
	}

	data, err := json.Marshal(results)
	if err != nil {
		log.Printf("Failed to marshal konflux results: %v", err)
		return
	}

	if err := os.WriteFile(outputFile, data, os.ModePerm); err != nil {
		log.Printf("Failed to write konflux results: %v", err)
	}
}

// configureGinkgo sets up Ginkgo configuration from viper settings.
func configureGinkgo(suiteConfig *types.SuiteConfig, reporterConfig *types.ReporterConfig) {
	suiteConfig.Timeout = time.Hour * time.Duration(viper.GetInt(config.Tests.SuiteTimeout))
	suiteConfig.DryRun = viper.GetBool(config.DryRun)

	if skip := viper.GetString(config.Tests.GinkgoSkip); skip != "" {
		suiteConfig.SkipStrings = append(suiteConfig.SkipStrings, skip)
	}

	if labels := viper.GetString(config.Tests.GinkgoLabelFilter); labels != "" {
		suiteConfig.LabelFilter = labels
	}

	if tests := viper.GetStringSlice(config.Tests.TestsToRun); len(tests) > 0 {
		suiteConfig.FocusStrings = tests
	}

	if focus := viper.GetString(config.Tests.GinkgoFocus); focus != "" {
		suiteConfig.FocusStrings = append(suiteConfig.FocusStrings, focus)
	}

	reporterConfig.NoColor = true
	switch viper.GetString(config.Tests.GinkgoLogLevel) {
	case "v":
		reporterConfig.Verbose = true
	case "vv":
		reporterConfig.VeryVerbose = true
	default:
		reporterConfig.Succinct = true
	}

	if suiteConfig.DryRun {
		log.Println("\x1b[33mWARNING! This is a DRY RUN. Review this state if outcome is unexpected.\033[0m")
	}
}
