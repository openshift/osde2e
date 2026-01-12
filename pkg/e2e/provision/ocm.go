// Package provision provides implementations of the Provisioner interface
// for various cluster providers.
package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// OCMProvisioner implements the Provisioner interface using OCM (OpenShift Cluster Manager).
type OCMProvisioner struct {
	provider spi.Provider
}

// NewOCMProvisioner creates a new OCM-based provisioner.
func NewOCMProvisioner() (*OCMProvisioner, error) {
	provider, err := providers.ClusterProvider()
	if err != nil {
		return nil, fmt.Errorf("error getting cluster provider: %w", err)
	}
	return &OCMProvisioner{
		provider: provider,
	}, nil
}

// Provision ensures a cluster is available and returns connection details.
func (p *OCMProvisioner) Provision(ctx context.Context) (*orchestrator.ClusterInfo, error) {
	// Load existing cluster ID if available (for multi-step jobs)
	if err := config.LoadClusterId(); err != nil {
		log.Printf("Not loading cluster id: %v", err)
	}

	// Load kubeconfig if available
	if err := config.LoadKubeconfig(); err != nil {
		log.Printf("Not loading kubeconfig: %v", err)
	}

	var spiCluster *spi.Cluster
	var err error

	// Check if we already have a kubeconfig (reusing existing cluster)
	if viper.GetString(config.Kubeconfig.Contents) == "" {
		log.Println("Provisioning new cluster...")
		spiCluster, err = cluster.Provision(p.provider)
		if err != nil {
			// Try to get logs even on failure
			p.getLogs()
			return nil, fmt.Errorf("failed to provision cluster: %w", err)
		}
		log.Printf("Cluster provisioned successfully: %s", spiCluster.ID())
	} else {
		log.Println("Using provided kubeconfig")
		clusterID := viper.GetString(config.Cluster.ID)
		if clusterID == "" {
			return nil, fmt.Errorf("cluster ID is required when using existing kubeconfig")
		}
		spiCluster, err = p.provider.GetCluster(clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster %s: %w", clusterID, err)
		}
	}

	// Update viper config with cluster details
	cluster.SetClusterIntoViperConfig(spiCluster)

	// Install addons if configured
	if len(viper.GetString(config.Addons.IDs)) > 0 {
		if viper.GetString(config.Provider) != "mock" {
			if err := p.installAddons(spiCluster.ID()); err != nil {
				log.Printf("Failed installing addons: %v", err)
				p.getLogs()
				return nil, fmt.Errorf("addon installation failed: %w", err)
			}
		} else {
			log.Println("Skipping addon installation due to mock provider.")
		}
	}

	// Get kubeconfig
	kubeconfig := []byte(viper.GetString(config.Kubeconfig.Contents))
	if len(kubeconfig) == 0 {
		kubeconfig, err = p.provider.ClusterKubeconfig(spiCluster.ID())
		if err != nil {
			return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
		}
	}

	// Build ClusterInfo
	clusterInfo := &orchestrator.ClusterInfo{
		ID:         spiCluster.ID(),
		Name:       spiCluster.Name(),
		Provider:   p.provider.Type(),
		Region:     spiCluster.Region(),
		Version:    spiCluster.Version(),
		Kubeconfig: kubeconfig,
		Properties: map[string]interface{}{
			"state":      spiCluster.State(),
			"cloudProvider": spiCluster.CloudProvider(),
			"flavour":    spiCluster.Flavour(),
			"addons":     spiCluster.Addons(),
		},
	}

	return clusterInfo, nil
}

// Destroy tears down the cluster if configured to do so.
func (p *OCMProvisioner) Destroy(ctx context.Context, clusterInfo *orchestrator.ClusterInfo) error {
	if clusterInfo == nil || clusterInfo.ID == "" {
		log.Println("Cluster ID is empty, unable to destroy cluster")
		return nil
	}

	// Get final logs before destruction
	p.getLogs()

	// Perform cleanup operations
	if err := p.cleanup(ctx, clusterInfo); err != nil {
		log.Printf("Cleanup operations had errors: %v", err)
		// Don't fail destruction on cleanup errors
	}

	// Delete cluster if configured
	if !viper.GetBool(config.Cluster.SkipDestroyCluster) {
		log.Printf("Destroying cluster '%s'...", clusterInfo.ID)
		if err := p.provider.DeleteCluster(clusterInfo.ID); err != nil {
			return fmt.Errorf("error deleting cluster: %w", err)
		}
		log.Printf("Cluster %s deletion initiated", clusterInfo.ID)
	} else {
		log.Printf("Skipping cluster destruction. For debugging, cluster ID: %s in environment: %s",
			clusterInfo.ID, p.provider.Environment())
	}

	return nil
}

// GetKubeconfig returns raw kubeconfig bytes for cluster access.
func (p *OCMProvisioner) GetKubeconfig(ctx context.Context, clusterInfo *orchestrator.ClusterInfo) ([]byte, error) {
	if clusterInfo == nil || clusterInfo.ID == "" {
		return nil, fmt.Errorf("cluster info is required")
	}

	// Return cached kubeconfig if available
	if len(clusterInfo.Kubeconfig) > 0 {
		return clusterInfo.Kubeconfig, nil
	}

	// Fetch from provider
	kubeconfig, err := p.provider.ClusterKubeconfig(clusterInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig for cluster %s: %w", clusterInfo.ID, err)
	}

	return kubeconfig, nil
}

// Health checks cluster health and readiness.
func (p *OCMProvisioner) Health(ctx context.Context, clusterInfo *orchestrator.ClusterInfo) (*orchestrator.HealthStatus, error) {
	if clusterInfo == nil || clusterInfo.ID == "" {
		return nil, fmt.Errorf("cluster info is required")
	}

	spiCluster, err := p.provider.GetCluster(clusterInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster state: %w", err)
	}

	state := spiCluster.State()
	ready := state == spi.ClusterStateReady

	return &orchestrator.HealthStatus{
		Ready:   ready,
		Message: fmt.Sprintf("Cluster state: %v", state),
		Conditions: map[string]bool{
			"ready": ready,
		},
	}, nil
}

// installAddons installs addons onto the cluster.
func (p *OCMProvisioner) installAddons(clusterID string) error {
	params := make(map[string]map[string]string)
	strParams := viper.GetString(config.Addons.Parameters)
	if strParams != "" {
		if err := json.Unmarshal([]byte(strParams), &params); err != nil {
			return fmt.Errorf("failed unmarshalling addon parameters %s: %w", strParams, err)
		}
	}

	addonIDs := strings.Split(viper.GetString(config.Addons.IDs), ",")
	num, err := p.provider.InstallAddons(clusterID, addonIDs, params)
	if err != nil {
		return fmt.Errorf("could not install addons: %w", err)
	}

	if num > 0 {
		log.Printf("Installed %d addon(s), waiting for cluster ready...", num)
		if err = cluster.WaitForClusterReadyPostInstall(clusterID, nil); err != nil {
			return fmt.Errorf("failed waiting for cluster ready: %w", err)
		}
	}

	return nil
}

// getLogs retrieves cluster logs from the provider.
func (p *OCMProvisioner) getLogs() {
	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" {
		log.Println("CLUSTER_ID is not set, skipping log collection...")
		return
	}

	logs, err := p.provider.Logs(clusterID)
	if err != nil {
		log.Printf("Error collecting cluster logs: %s", err.Error())
		return
	}

	// Logs will be written by the reporter
	// Store them in viper for now (this maintains existing behavior)
	// TODO: Consider moving this to reporter package
	for k, v := range logs {
		viper.Set(fmt.Sprintf("logs.%s", k), v)
	}
}

// cleanup performs cleanup operations before cluster destruction.
func (p *OCMProvisioner) cleanup(ctx context.Context, clusterInfo *orchestrator.ClusterInfo) error {
	spiCluster, err := p.provider.GetCluster(clusterInfo.ID)
	if err != nil {
		log.Printf("Error getting cluster for cleanup: %s", err.Error())
		return err
	}

	// Determine cluster status
	clusterStatus := clusterproperties.StatusCompletedFailing
	if viper.GetBool(config.Cluster.Passing) {
		clusterStatus = clusterproperties.StatusCompletedPassing
	}

	// Log cluster information
	log.Printf("Cluster addons: %v", spiCluster.Addons())
	log.Printf("Cluster cloud provider: %v", spiCluster.CloudProvider())
	log.Printf("Cluster expiration: %v", spiCluster.ExpirationTimestamp())
	log.Printf("Cluster flavor: %s", spiCluster.Flavour())
	log.Printf("Cluster state: %v", spiCluster.State())

	// Set cluster properties
	_ = p.provider.AddProperty(spiCluster, clusterproperties.Status, clusterStatus)
	_ = p.provider.AddProperty(spiCluster, clusterproperties.JobID, "")
	_ = p.provider.AddProperty(spiCluster, clusterproperties.JobName, "")
	_ = p.provider.AddProperty(spiCluster, clusterproperties.Availability, clusterproperties.Used)

	// Handle expiration for nightly tests
	if viper.GetString(config.Cluster.InstallSpecificNightly) != "" ||
		viper.GetString(config.Cluster.ReleaseImageLatest) != "" {
		if err := p.provider.Expire(clusterInfo.ID, 30*time.Minute); err != nil {
			log.Printf("Error setting cluster expiration: %v", err)
		}
	}

	// Extend expiry if cluster is being preserved and conditions are met
	if viper.GetBool(config.Cluster.SkipDestroyCluster) &&
		!viper.GetBool(config.Cluster.ClaimedFromReserve) &&
		clusterStatus != clusterproperties.StatusCompletedError &&
		viper.GetString(config.Addons.IDs) == "" {
		
		if !spiCluster.ExpirationTimestamp().Add(6 * time.Hour).After(spiCluster.CreationTimestamp().Add(24 * time.Hour)) {
			if err := p.provider.ExtendExpiry(clusterInfo.ID, 6, 0, 0); err != nil {
				log.Printf("Error extending cluster expiration: %s", err.Error())
			}
		}
	}

	return nil
}

