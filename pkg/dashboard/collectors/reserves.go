package collectors

import (
	"fmt"
	"log"
	"time"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/dashboard/models"
)

// ReserveCollector collects cluster reserve information from one or more OCM environments.
type ReserveCollector struct {
	providers map[string]*ocmprovider.OCMProvider
}

// NewReserveCollector creates a new reserve collector for the given OCM environments.
func NewReserveCollector(envs ...string) (*ReserveCollector, error) {
	providers := make(map[string]*ocmprovider.OCMProvider, len(envs))
	for _, env := range envs {
		p, err := ocmprovider.NewWithEnv(env)
		if err != nil {
			log.Printf("Warning: could not create provider for environment %s: %v (skipping)", env, err)
			continue
		}
		providers[env] = p
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("could not connect to any OCM environment")
	}
	return &ReserveCollector{providers: providers}, nil
}

// CollectReserves retrieves reserved clusters from all configured OCM environments.
func (c *ReserveCollector) CollectReserves() ([]models.ClusterReserve, error) {
	query := fmt.Sprintf(
		"properties.MadeByOSDe2e='true' AND properties.Availability like '%s%%'",
		clusterproperties.Reserved,
	)

	var all []models.ClusterReserve
	for env, p := range c.providers {
		resp, err := p.GetConnection().ClustersMgmt().V1().Clusters().List().
			Search(query).
			Size(500).
			Send()
		if err != nil {
			if isAuthError(err) {
				log.Printf("Info: skipping reserves for env %q (OCM account not available)", env)
			} else {
				log.Printf("Warning: failed to query reserved clusters for env %q: %v", env, err)
			}
			continue
		}
		resp.Items().Each(func(cluster *v1.Cluster) bool {
			all = append(all, c.ocmClusterToReserve(cluster))
			return true
		})
	}

	log.Printf("Collected %d reserved clusters from OCM", len(all))
	return all, nil
}

// ocmClusterToReserve converts an OCM cluster to a ClusterReserve model
func (c *ReserveCollector) ocmClusterToReserve(cluster *v1.Cluster) models.ClusterReserve {
	reserve := models.ClusterReserve{
		ID:            cluster.ID(),
		Name:          cluster.Name(),
		State:         string(cluster.State()),
		Version:       cluster.Version().ID(),
		Region:        cluster.Region().ID(),
		CloudProvider: cluster.CloudProvider().ID(),
		CreatedAt:     cluster.CreationTimestamp(),
		ExpiresAt:     cluster.ExpirationTimestamp(),
		Product:       cluster.Product().ID(),
		Properties:    make(map[string]string),
	}

	// Extract availability from properties
	if props, ok := cluster.GetProperties(); ok {
		for k, v := range props {
			reserve.Properties[k] = v
			if k == clusterproperties.Availability {
				reserve.Availability = v
			}
		}
	}

	return reserve
}

// CollectClustersPerEnv returns all osde2e clusters grouped by environment name.
func (c *ReserveCollector) CollectClustersPerEnv() (map[string][]models.ClusterReserve, error) {
	result := make(map[string][]models.ClusterReserve)
	for env, p := range c.providers {
		resp, err := p.GetConnection().ClustersMgmt().V1().Clusters().List().
			Search("properties.MadeByOSDe2e='true'").
			Size(1000).
			Send()
		if err != nil {
			if isAuthError(err) {
				log.Printf("Info: skipping clusters for env %q (OCM account not available)", env)
			} else {
				log.Printf("Warning: failed to query clusters for env %q: %v", env, err)
			}
			continue
		}
		var clusters []models.ClusterReserve
		resp.Items().Each(func(cluster *v1.Cluster) bool {
			clusters = append(clusters, c.ocmClusterToReserve(cluster))
			return true
		})
		result[env] = clusters
	}
	return result, nil
}

// CountExpiringSoon counts clusters expiring within the given threshold
func (c *ReserveCollector) CountExpiringSoon(reserves []models.ClusterReserve, threshold time.Duration) int {
	count := 0
	for _, r := range reserves {
		if r.IsExpiringSoon(threshold) {
			count++
		}
	}
	return count
}
