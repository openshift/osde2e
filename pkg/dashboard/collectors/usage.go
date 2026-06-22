package collectors

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/dashboard/models"
)

// UsageCollector collects cluster usage metrics from one or more OCM environments.
type UsageCollector struct {
	// providers maps environment name → OCMProvider for that env
	providers map[string]*ocmprovider.OCMProvider
}

// NewUsageCollector creates a UsageCollector that queries the given OCM environments
// in parallel. Each env must be a valid OCM environment name ("stage", "int", "prod", etc.).
// Environments that fail to connect are skipped with a warning.
func NewUsageCollector(envs ...string) (*UsageCollector, error) {
	if len(envs) == 0 {
		envs = []string{"stage", "int"}
	}

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

	return &UsageCollector{providers: providers}, nil
}

// CollectUsage queries all configured OCM environments in parallel and returns
// one ClusterUsage entry per environment.
func (c *UsageCollector) CollectUsage() ([]models.ClusterUsage, error) {
	type result struct {
		env   string
		usage *models.ClusterUsage
		err   error
	}

	ch := make(chan result, len(c.providers))
	var wg sync.WaitGroup

	for env, provider := range c.providers {
		wg.Add(1)
		go func(env string, p *ocmprovider.OCMProvider) {
			defer wg.Done()
			usage, err := collectUsageForEnv(env, p)
			ch <- result{env: env, usage: usage, err: err}
		}(env, provider)
	}

	wg.Wait()
	close(ch)

	var usages []models.ClusterUsage
	for r := range ch {
		if r.err != nil {
			if isAuthError(r.err) {
				log.Printf("Info: skipping env %q (OCM account not available in this environment)", r.env)
			} else {
				log.Printf("Warning: failed to collect usage for env %q: %v", r.env, r.err)
			}
			continue
		}
		usages = append(usages, *r.usage)
	}

	log.Printf("Collected usage metrics for %d environments", len(usages))
	return usages, nil
}

// collectUsageForEnv queries a single OCM environment and returns its ClusterUsage.
func collectUsageForEnv(env string, provider *ocmprovider.OCMProvider) (*models.ClusterUsage, error) {
	query := "properties.MadeByOSDe2e='true'"

	resp, err := provider.GetConnection().ClustersMgmt().V1().Clusters().List().
		Search(query).
		Size(1000).
		Send()
	if err != nil {
		return nil, fmt.Errorf("failed to query clusters: %w", err)
	}

	usage := &models.ClusterUsage{
		Environment:     env,
		ByState:         make(map[string]int),
		ByAvailability:  make(map[string]int),
		ByCloudProvider: make(map[string]int),
		ByVersion:       make(map[string]int),
		LastUpdated:     time.Now(),
	}

	resp.Items().Each(func(cluster *v1.Cluster) bool {
		usage.TotalClusters++
		usage.ByState[string(cluster.State())]++
		usage.ByCloudProvider[cluster.CloudProvider().ID()]++
		usage.ByVersion[cluster.Version().ID()]++

		if props, ok := cluster.GetProperties(); ok {
			if avail, exists := props[clusterproperties.Availability]; exists {
				usage.ByAvailability[avail]++
			}
		}

		return true
	})

	return usage, nil
}

// CollectUsageByEnvironment retrieves usage for a specific environment.
func (c *UsageCollector) CollectUsageByEnvironment(env string) (*models.ClusterUsage, error) {
	p, ok := c.providers[env]
	if !ok {
		return &models.ClusterUsage{
			Environment:     env,
			ByState:         make(map[string]int),
			ByAvailability:  make(map[string]int),
			ByCloudProvider: make(map[string]int),
			ByVersion:       make(map[string]int),
			LastUpdated:     time.Now(),
		}, nil
	}
	return collectUsageForEnv(env, p)
}

// isAuthError returns true for OCM errors that indicate the token is not valid
// for a given environment (401, 403, 422 user-not-found). These are expected when
// running with a stage/int token against prod, and should not be surfaced as warnings.
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "status is 401") ||
		strings.Contains(msg, "status is 403") ||
		(strings.Contains(msg, "status is 422") && strings.Contains(msg, "does not exist"))
}
