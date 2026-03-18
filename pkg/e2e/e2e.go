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
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/cluster"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/phase"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/slack"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/debug"
	"github.com/openshift/osde2e/pkg/e2e/adhoctestimages"
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
		if err := orch.Report(ctx); err != nil {
			log.Printf("Report errors: %v", err)
		}
		return config.Failure
	}

	// Execute tests
	testErr := orch.Execute(ctx)

	// On failure: analyze logs (results are cached for Report)
	if testErr != nil {
		log.Printf("Tests failed: %v", testErr)

		if viper.GetBool(config.LogAnalysis.EnableAnalysis) && viper.GetString(config.Tests.TestSuites) == "" && viper.GetString(config.Tests.AdHocTestImages) == "" {
			if err := orch.AnalyzeLogs(ctx, testErr); err != nil {
				log.Printf("Log analysis failed: %v", err)
			}
		}
	}

	// Post-process cluster: must-gather, cluster state inspection, property
	// updates. Runs before Report so diagnostic data is included in the
	// S3 artifact upload.
	if err := orch.PostProcessCluster(ctx); err != nil {
		log.Printf("Cluster post-processing errors: %v", err)
	}

	// Report: upload artifacts, send notifications, generate reports
	if err := orch.Report(ctx); err != nil {
		log.Printf("Report errors: %v", err)
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
	s3Results      []aws.S3UploadResult
	analysisResult *analysisengine.Result
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
		o.result.ExitCode = config.Failure
		return err
	}

	// Provision or reuse cluster
	cl, err := cluster.ProvisionOrReuseCluster(o.provider)
	runner.ReportClusterInstallLogs(o.provider)
	if err != nil {
		o.result.ExitCode = config.Failure
		return err
	}

	o.result.ClusterID = cl.ID()

	// Install addons if configured
	if _, err := cluster.InstallAddonsIfConfigured(o.provider, cl.ID()); err != nil {
		runner.ReportClusterInstallLogs(o.provider)
		o.result.ExitCode = config.Failure
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

// cleanStaleJunitFiles removes junit XML files from previous runs in the report directory,
// keeping only the file matching the current run suffix.
func cleanStaleJunitFiles() {
	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		return
	}
	suffix := viper.GetString(config.Suffix)
	currentJunit := "junit_" + suffix + ".xml"

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := info.Name()
		if strings.HasPrefix(name, "junit_") && strings.HasSuffix(name, ".xml") && name != currentJunit {
			os.Remove(path)
		}
		return nil
	}
	_ = filepath.Walk(reportDir, walkFn)
}

// AnalyzeLogs performs AI-powered log analysis on test failures.
// Results are cached on the orchestrator for use by Report.
func (o *E2EOrchestrator) AnalyzeLogs(ctx context.Context, testErr error) error {
	log.Println("Running log analysis...")
	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		return fmt.Errorf("no report directory available for log analysis")
	}

	engineConfig := &analysisengine.Config{
		BaseConfig: analysisengine.BaseConfig{
			ArtifactsDir: reportDir,
			APIKey:       viper.GetString(config.LogAnalysis.APIKey),
			ClusterInfo: &analysisengine.ClusterInfo{
				ID:            viper.GetString(config.Cluster.ID),
				Name:          viper.GetString(config.Cluster.Name),
				Provider:      viper.GetString(config.Provider),
				Region:        viper.GetString(config.CloudProvider.Region),
				CloudProvider: viper.GetString(config.CloudProvider.CloudProviderID),
				Version:       viper.GetString(config.Cluster.Version),
			},
		},
		PromptTemplate: "default",
		FailureContext: testErr.Error(),
	}

	engine, err := analysisengine.New(ctx, engineConfig)
	if err != nil {
		return fmt.Errorf("failed to create analysis engine: %w", err)
	}

	result, err := engine.Run(ctx)
	if err != nil {
		return fmt.Errorf("log analysis failed: %w", err)
	}

	o.analysisResult = result
	log.Printf("Log analysis completed. Results: %s/%s/", reportDir, analysisengine.AnalysisDirName)
	log.Printf("=== Log Analysis Result ===\n%s", result.Content)

	return nil
}

// Report uploads artifacts, sends notifications, and generates diagnostic reports.
func (o *E2EOrchestrator) Report(ctx context.Context) error {
	if o.suiteConfig.DryRun {
		return nil
	}

	// Upload artifacts to S3
	if viper.GetString(config.Tests.LogBucket) != "" {
		cleanStaleJunitFiles()
		if err := o.uploadToS3(); err != nil {
			log.Printf("S3 upload failed: %v", err)
		}
	}

	// Drain per-suite pending notifications. If any were queued (suites ran),
	// send them; otherwise fall back to global failure notification.
	pending := adhoctestimages.DrainPendingNotifications()
	if len(pending) > 0 {
		o.sendDeferredNotifications(ctx, pending)
	} else if o.result.ExitCode != config.Success && viper.GetBool(config.Tests.EnableSlackNotify) {
		o.sendFailureNotification(ctx)
	}

	runner.ReportClusterInstallLogs(o.provider)
	return nil
}

// sendFailureNotification sends a test failure notification via Slack.
// If LLM analysis results are available they are included; otherwise a
// basic failure notice is sent. Called by Report after S3 upload so that
// presigned URLs are available.
func (o *E2EOrchestrator) sendFailureNotification(ctx context.Context) {
	reportDir := viper.GetString(config.ReportDir)
	notificationConfig := slack.BuildNotificationConfig(
		viper.GetString(config.LogAnalysis.SlackWebhook),
		viper.GetString(config.Tests.SlackChannel),
		&slack.ClusterInfo{
			ID:       viper.GetString(config.Cluster.ID),
			Provider: viper.GetString(config.Provider),
			Version:  viper.GetString(config.Cluster.Version),
		},
		reportDir,
	)
	if notificationConfig == nil {
		return
	}

	if len(o.s3Results) > 0 {
		artifactLinks := s3ResultsToArtifactLinks(o.s3Results)
		for i := range notificationConfig.Reporters {
			notificationConfig.Reporters[i].Settings["artifact_links"] = artifactLinks
		}
	}

	var result *slack.AnalysisResult
	if o.analysisResult != nil {
		result = &slack.AnalysisResult{
			Status:   o.analysisResult.Status,
			Content:  o.analysisResult.Content,
			Metadata: o.analysisResult.Metadata,
			Error:    o.analysisResult.Error,
			Prompt:   o.analysisResult.Prompt,
		}
	} else {
		result = &slack.AnalysisResult{
			Status:  "skipped",
			Content: "Log analysis was not enabled for this run.",
		}
	}

	slackReporter := slack.NewSlackReporter()
	for _, cfg := range notificationConfig.Reporters {
		if err := slackReporter.Report(ctx, result, &cfg); err != nil {
			log.Printf("Failed to send failure notification via %s: %v", cfg.Type, err)
		}
	}
}

// sendDeferredNotifications delivers the given Slack notifications that were
// queued by adhoctestimages during test execution. Called by Report after S3
// upload so that presigned URLs are available for inclusion in the message.
func (o *E2EOrchestrator) sendDeferredNotifications(ctx context.Context, pending []adhoctestimages.PendingNotification) {
	webhook := viper.GetString(config.LogAnalysis.SlackWebhook)
	if webhook == "" || !viper.GetBool(config.Tests.EnableSlackNotify) {
		return
	}

	artifactLinks := s3ResultsToArtifactLinks(o.s3Results)
	slackReporter := slack.NewSlackReporter()
	var expiration string
	if o.provider != nil {
		if cl, err := o.provider.GetCluster(viper.GetString(config.Cluster.ID)); err != nil {
			log.Printf("failed to get cluster %s: %v", viper.GetString(config.Cluster.ID), err)
		} else if cl != nil {
			expiration = cl.ExpirationTimestamp().String()
		}
	}

	for _, p := range pending {
		if p.TestSuite.SlackChannel == "" {
			continue
		}

		cfg := slack.SlackReporterConfig(webhook, true)
		cfg.Settings["channel"] = p.TestSuite.SlackChannel
		cfg.Settings["image"] = p.TestSuite.Image
		cfg.Settings["env"] = viper.GetString(ocmprovider.Env)
		cfg.Settings["cluster_info"] = &slack.ClusterInfo{
			ID:         viper.GetString(config.Cluster.ID),
			Provider:   viper.GetString(config.Provider),
			Version:    viper.GetString(config.Cluster.Version),
			Expiration: expiration,
		}
		cfg.Settings["artifact_links"] = artifactLinks

		result := &slack.AnalysisResult{
			Status:  "completed",
			Content: p.AnalysisContent,
		}

		if err := slackReporter.Report(ctx, result, &cfg); err != nil {
			log.Printf("Failed to send deferred notification for %s: %v", p.TestSuite.Image, err)
		}
	}
}

// uploadToS3 uploads the report directory contents to S3 and caches results.
// Subsequent calls are no-ops if artifacts were already uploaded.
func (o *E2EOrchestrator) uploadToS3() error {
	if len(o.s3Results) > 0 {
		return nil
	}

	component := deriveComponentFromTestImage()
	uploader, err := aws.NewS3Uploader(component)
	if err != nil {
		return fmt.Errorf("failed to create S3 uploader: %w", err)
	}
	if uploader == nil {
		return nil
	}

	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		return fmt.Errorf("no report directory configured")
	}

	results, err := uploader.UploadDirectory(reportDir)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	o.s3Results = results
	return nil
}

// s3ResultsToArtifactLinks converts S3 upload results to artifact links for Slack.
// Returns links in a fixed order: test_output.log, junit_<suffix>.xml.
func s3ResultsToArtifactLinks(results []aws.S3UploadResult) []slack.ArtifactLink {
	suffix := viper.GetString(config.Suffix)
	currentJunit := "junit_" + suffix + ".xml"

	orderedNames := []string{"test_output.log", currentJunit}

	byName := make(map[string]aws.S3UploadResult, len(orderedNames))
	for _, r := range results {
		if r.PresignedURL == "" {
			continue
		}
		byName[filepath.Base(r.Key)] = r
	}

	links := make([]slack.ArtifactLink, 0, len(orderedNames))
	for _, name := range orderedNames {
		if r, ok := byName[name]; ok {
			links = append(links, slack.ArtifactLink{
				Name: name,
				URL:  r.PresignedURL,
				Size: r.Size,
			})
		}
	}
	return links
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

	// Post processing: must gather, cluster extension, cluster property updates (with ginkgo defer for recovery)
	defer ginkgo.GinkgoRecover()

	clusterStatus := "completed-failing"

	if !viper.GetBool(config.SkipMustGather) {
		if err := cluster.RunMustGather(ctx); err != nil {
			o.result.Errors = append(o.result.Errors, err)
			clusterStatus = "completed-error"
		}
		h, err := helper.NewOutsideGinkgo()
		if err != nil {
			log.Printf("Failed to generate helper object for cluster inspection: %v", err)
		} else {
			cluster.InspectClusterState(ctx, h)
		}
	}

	if clusterID := viper.GetString(config.Cluster.ID); clusterID != "" {
		if err := cluster.UpdateClusterProperties(o.provider, clusterStatus); err != nil {
			o.result.Errors = append(o.result.Errors, err)
		}
	}

	if err := cluster.HandleExpirationExtension(o.provider); err != nil {
		o.result.Errors = append(o.result.Errors, err)
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
