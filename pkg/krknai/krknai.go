// Package krknai provides an orchestrator implementation for krkn AI chaos testing.
package krknai

import (
	"context"
	"fmt"
	"log"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/internal/analysisengine"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	ctrlog "sigs.k8s.io/controller-runtime/pkg/log"
)

// KrknAIOrchestrator implements the orchestrator.Orchestrator interface for krkn AI chaos testing.
type KrknAIOrchestrator struct {
	provider spi.Provider
	result   *orchestrator.Result
	dryRun   bool
}

// NewOrchestrator creates a new KrknAI orchestrator instance.
func NewOrchestrator(ctx context.Context) (orchestrator.Orchestrator, error) {
	provider, err := providers.ClusterProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster provider: %w", err)
	}

	return &KrknAIOrchestrator{
		provider: provider,
		dryRun:   viper.GetBool(config.DryRun),
		result:   &orchestrator.Result{ExitCode: config.Success},
	}, nil
}

// Provision prepares the cluster environment.
func (o *KrknAIOrchestrator) Provision(ctx context.Context) error {
	ctrlog.SetLogger(ginkgo.GinkgoLogr)

	if err := orchestrator.LoadClusterContext(); err != nil {
		return err
	}

	cluster, err := orchestrator.ProvisionOrReuseCluster(o.provider)
	orchestrator.CollectAndWriteLogs(o.provider)
	if err != nil {
		return err
	}

	o.result.ClusterID = cluster.ID()

	if _, err := orchestrator.InstallAddonsIfConfigured(o.provider, cluster.ID()); err != nil {
		orchestrator.CollectAndWriteLogs(o.provider)
		return fmt.Errorf("addon installation failed: %w", err)
	}

	return nil
}

// Execute runs krkn AI chaos tests - currently a no-op placeholder.
func (o *KrknAIOrchestrator) Execute(ctx context.Context) error {
	o.result.TestsPassed = true
	o.result.UpgradePassed = true
	return nil
}

// AnalyzeLogs performs AI-powered log analysis on test failures.
func (o *KrknAIOrchestrator) AnalyzeLogs(ctx context.Context, testErr error) error {
	if !viper.GetBool(config.LogAnalysis.EnableAnalysis) {
		return nil
	}

	log.Println("Running log analysis...")

	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		return fmt.Errorf("no report directory available for log analysis")
	}

	engineConfig := &analysisengine.Config{
		ArtifactsDir:   reportDir,
		PromptTemplate: "default",
		APIKey:         viper.GetString(config.LogAnalysis.APIKey),
		FailureContext: testErr.Error(),
		ClusterInfo: &analysisengine.ClusterInfo{
			ID:            viper.GetString(config.Cluster.ID),
			Name:          viper.GetString(config.Cluster.Name),
			Provider:      viper.GetString(config.Provider),
			Region:        viper.GetString(config.CloudProvider.Region),
			CloudProvider: viper.GetString(config.CloudProvider.CloudProviderID),
			Version:       viper.GetString(config.Cluster.Version),
		},
		NotificationConfig: orchestrator.BuildNotificationConfig(),
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
func (o *KrknAIOrchestrator) Report(ctx context.Context) error {
	if o.dryRun {
		return nil
	}
	orchestrator.CollectAndWriteLogs(o.provider)
	return nil
}

// PostProcessCluster performs post-processing on the cluster.
func (o *KrknAIOrchestrator) PostProcessCluster(ctx context.Context) error {
	if o.dryRun {
		return nil
	}

	h, err := helper.NewOutsideGinkgo()
	if h == nil {
		return fmt.Errorf("failed to generate helper for post-processing: %w", err)
	}

	defer ginkgo.GinkgoRecover()
	if errs := orchestrator.PostProcessE2E(ctx, o.provider, h); len(errs) > 0 {
		o.result.Errors = append(o.result.Errors, errs...)
	}

	return nil
}

// Cleanup performs post-test cleanup and optionally destroys the cluster.
func (o *KrknAIOrchestrator) Cleanup(ctx context.Context) error {
	if o.dryRun {
		return nil
	}

	if err := orchestrator.DeleteCluster(o.provider); err != nil {
		o.result.Errors = append(o.result.Errors, err)
		return err
	}

	return nil
}

// Result returns the test execution result.
func (o *KrknAIOrchestrator) Result() *orchestrator.Result {
	return o.result
}
