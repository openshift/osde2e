package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/osde2e/internal/reporter"
	clusterutil "github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// ProvisionOrReuseCluster either provisions a new cluster or retrieves an existing one
// based on whether a kubeconfig is already available.
func ProvisionOrReuseCluster(provider spi.Provider) (*spi.Cluster, error) {
	var cluster *spi.Cluster
	var err error

	if viper.GetString(config.Kubeconfig.Contents) == "" {
		// Provision new cluster
		cluster, err = clusterutil.Provision(provider)
		if err != nil {
			return nil, fmt.Errorf("cluster provisioning failed: %w", err)
		}
	} else {
		// Reuse existing cluster
		log.Println("Using provided kubeconfig")
		clusterID := viper.GetString(config.Cluster.ID)
		cluster, err = provider.GetCluster(clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster %s: %w", clusterID, err)
		}
	}

	// Set cluster into viper config for downstream usage
	clusterutil.SetClusterIntoViperConfig(cluster)
	return cluster, nil
}

// LoadClusterContext loads kubeconfig and cluster ID from configuration.
// This should be called before provisioning to ensure context is available.
func LoadClusterContext() error {
	// Load kubeconfig if available
	if err := config.LoadKubeconfig(); err != nil {
		log.Printf("Not loading kubeconfig: %v", err)
	}

	// Load cluster ID from shared directory
	if err := config.LoadClusterId(); err != nil {
		log.Printf("Not loading cluster id: %v", err)
		return fmt.Errorf("failed to load cluster ID: %w", err)
	}

	return nil
}

// InstallAddonsIfConfigured installs addons on the cluster if configured in viper.
// Returns true if addons were installed, false otherwise.
func InstallAddonsIfConfigured(provider spi.Provider, clusterID string) (bool, error) {
	addonIDsStr := viper.GetString(config.Addons.IDs)
	if len(addonIDsStr) == 0 {
		return false, nil
	}

	// Skip addon installation for mock provider
	if viper.GetString(config.Provider) == "mock" {
		log.Println("Skipping addon installation for mock provider")
		return false, nil
	}

	addonIDs := strings.Split(addonIDsStr, ",")
	params := make(map[string]map[string]string)

	// Parse addon parameters if provided
	if strParams := viper.GetString(config.Addons.Parameters); strParams != "" {
		if err := json.Unmarshal([]byte(strParams), &params); err != nil {
			return false, fmt.Errorf("failed to unmarshal addon parameters: %w", err)
		}
	}

	// Install addons
	num, err := provider.InstallAddons(clusterID, addonIDs, params)
	if err != nil {
		return false, fmt.Errorf("failed to install addons: %w", err)
	}

	// Wait for cluster to be ready after addon installation
	if num > 0 {
		if err := clusterutil.WaitForClusterReadyPostInstall(clusterID, nil); err != nil {
			return false, fmt.Errorf("cluster not ready after addon installation: %w", err)
		}
	}

	return num > 0, nil
}

// CollectAndWriteLogs retrieves cluster logs from the provider and writes them to the report directory.
func CollectAndWriteLogs(provider spi.Provider) {
	if provider == nil {
		log.Println("Skipping log collection (no provider)")
		return
	}

	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" {
		log.Println("Skipping log collection (no cluster ID)")
		return
	}

	logs, err := provider.Logs(clusterID)
	if err != nil {
		log.Printf("Error collecting cluster logs: %v", err)
		return
	}

	WriteLogs(logs)
}

// WriteLogs writes logs to the report directory.
func WriteLogs(logs map[string][]byte) {
	reportDir := viper.GetString(config.ReportDir)
	for name, content := range logs {
		filePath := filepath.Join(reportDir, name+"-log.txt")
		if err := os.WriteFile(filePath, content, os.ModePerm); err != nil {
			log.Printf("Error writing log %s: %v", filePath, err)
		}
	}
}

// DeleteCluster destroys the cluster if configured.
func DeleteCluster(provider spi.Provider) error {
	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" {
		return nil
	}

	if !viper.GetBool(config.Cluster.SkipDestroyCluster) {
		log.Printf("Destroying cluster '%s'...", clusterID)
		if err := provider.DeleteCluster(clusterID); err != nil {
			return fmt.Errorf("failed to delete cluster: %w", err)
		}
	} else {
		log.Printf("Cluster %s preserved in environment %s", clusterID, provider.Environment())
	}

	return nil
}

// RunMustGather executes must-gather and collects results.
func RunMustGather(ctx context.Context, h *helper.H) error {
	log.Print("Running must-gather...")
	h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")

	r := h.Runner(fmt.Sprintf("oc adm must-gather --dest-dir=%v", runner.DefaultRunner.OutputDir))
	r.Name = "must-gather"
	r.Tarball = true

	if err := r.Run(1800, make(chan struct{})); err != nil {
		return fmt.Errorf("must-gather failed: %w", err)
	}

	results, err := r.RetrieveResults()
	if err != nil {
		return fmt.Errorf("failed to retrieve must-gather results: %w", err)
	}

	h.WriteResults(results)
	return nil
}

// InspectClusterState gathers cluster state information.
func InspectClusterState(ctx context.Context, h *helper.H) {
	log.Print("Gathering project states...")
	h.InspectState(ctx)

	log.Print("Gathering OLM state...")
	if err := h.InspectOLM(ctx); err != nil {
		log.Printf("Error inspecting OLM: %v", err)
	}
}

// UpdateClusterProperties updates cluster metadata in the provider.
func UpdateClusterProperties(provider spi.Provider, status string) error {
	clusterID := viper.GetString(config.Cluster.ID)
	cluster, err := provider.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	log.Printf("Cluster state: %v, flavor: %s", cluster.State(), cluster.Flavour())

	if viper.GetBool(config.Cluster.Passing) {
		status = clusterproperties.StatusCompletedPassing
	}

	properties := map[string]string{
		clusterproperties.Status:       status,
		clusterproperties.JobID:        "",
		clusterproperties.JobName:      "",
		clusterproperties.Availability: clusterproperties.Used,
	}

	for key, value := range properties {
		if err := provider.AddProperty(cluster, key, value); err != nil {
			return fmt.Errorf("failed to set property %s: %w", key, err)
		}
	}

	return nil
}

// HandleExpirationExtension extends cluster expiration for specific scenarios.
func HandleExpirationExtension(provider spi.Provider) error {
	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" || !viper.GetBool(config.Cluster.SkipDestroyCluster) {
		return nil
	}

	// Extend expiration for nightly builds
	if viper.GetString(config.Cluster.InstallSpecificNightly) != "" || viper.GetString(config.Cluster.ReleaseImageLatest) != "" {
		if err := provider.Expire(clusterID, 30*time.Minute); err != nil {
			return err
		}
	}

	// Extend expiration for passing clusters without addons
	if !viper.GetBool(config.Cluster.ClaimedFromReserve) && viper.GetString(config.Addons.IDs) == "" {
		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			return err
		}

		if !cluster.ExpirationTimestamp().Add(6 * time.Hour).After(cluster.CreationTimestamp().Add(24 * time.Hour)) {
			if err := provider.ExtendExpiry(clusterID, 6, 0, 0); err != nil {
				return err
			}
		}
	}

	return nil
}

// PostProcessE2E performs post-test cleanup and diagnostics collection.
func PostProcessE2E(ctx context.Context, provider spi.Provider, h *helper.H) []error {
	var errors []error
	clusterStatus := clusterproperties.StatusCompletedFailing

	// Run must-gather
	if !viper.GetBool(config.SkipMustGather) {
		if err := RunMustGather(ctx, h); err != nil {
			errors = append(errors, err)
			clusterStatus = clusterproperties.StatusCompletedError
		}
		InspectClusterState(ctx, h)
	}

	// Update cluster properties
	if clusterID := viper.GetString(config.Cluster.ID); clusterID != "" {
		if err := UpdateClusterProperties(provider, clusterStatus); err != nil {
			errors = append(errors, err)
		}
	}

	h.Cleanup(ctx)

	// Extend expiration for nightly builds
	if err := HandleExpirationExtension(provider); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// BuildNotificationConfig creates notification configuration for log analysis.
func BuildNotificationConfig() *reporter.NotificationConfig {
	if !viper.GetBool(config.Tests.EnableSlackNotify) {
		return nil
	}

	webhook := viper.GetString(config.LogAnalysis.SlackWebhook)
	channel := viper.GetString(config.LogAnalysis.SlackChannel)
	if webhook == "" || channel == "" {
		return nil
	}

	slackConfig := reporter.SlackReporterConfig(webhook, true)
	slackConfig.Settings["channel"] = channel

	return &reporter.NotificationConfig{
		Enabled:   true,
		Reporters: []reporter.ReporterConfig{slackConfig},
	}
}
