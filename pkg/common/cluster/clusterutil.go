package cluster

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	osconfig "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/openshift/osde2e/pkg/common/cluster/healthchecks"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/state"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// errorWindow is the number of checks made to determine if a cluster has truly failed.
	errorWindow = 5
)

// WaitForClusterReady blocks until the cluster is ready for testing.
func WaitForClusterReady(provider spi.Provider, clusterID string) error {
	cfg := config.Instance

	log.Printf("Waiting %v minutes for cluster '%s' to be ready...\n", cfg.Cluster.InstallTimeout, clusterID)
	cleanRuns := 0
	errRuns := 0

	clusterStarted := time.Now()
	var readinessStarted time.Time
	ocmReady := false
	if !cfg.Tests.SkipClusterHealthChecks {
		return wait.PollImmediate(30*time.Second, time.Duration(cfg.Cluster.InstallTimeout)*time.Minute, func() (bool, error) {
			if cluster, err := provider.GetCluster(clusterID); err == nil && cluster.State() == spi.ClusterStateReady {
				// This is the first time that we've entered this section, so we'll consider this the time until OCM has said the cluster is ready
				if !ocmReady {
					ocmReady = true
					if metadata.Instance.TimeToOCMReportingInstalled == 0 {
						metadata.Instance.SetTimeToOCMReportingInstalled(time.Since(clusterStarted).Seconds())
					}

					readinessStarted = time.Now()
				}
				if success, err := pollClusterHealth(provider, clusterID); success {
					cleanRuns++
					log.Printf("Clean run %d/%d...", cleanRuns, config.Instance.Cluster.CleanCheckRuns)
					errRuns = 0
					if cleanRuns == config.Instance.Cluster.CleanCheckRuns {
						if metadata.Instance.TimeToClusterReady == 0 {
							metadata.Instance.SetTimeToClusterReady(time.Since(readinessStarted).Seconds())
						} else {
							metadata.Instance.SetTimeToUpgradedClusterReady(time.Since(readinessStarted).Seconds())
						}

						return true, nil
					}
					return false, nil
				} else {
					if err != nil {
						errRuns++
						log.Printf("Error in PollClusterHealth: %v", err)
						if errRuns >= errorWindow {
							return false, fmt.Errorf("PollClusterHealth has returned an error %d times in a row. Failing osde2e", errorWindow)
						}
					}
					cleanRuns = 0
					return false, nil
				}
			} else if err != nil {
				return false, fmt.Errorf("Encountered error waiting for cluster: %v", err)
			} else if cluster.State() == spi.ClusterStateError {
				return false, fmt.Errorf("the installation of cluster '%s' has errored", clusterID)
			} else {
				log.Printf("Cluster is not ready, current status '%s'.", cluster.State())
			}
			return false, nil
		})
	}
	return nil
}

// PollClusterHealth looks at CVO data to determine if a cluster is alive/healthy or not
func pollClusterHealth(provider spi.Provider, clusterID string) (status bool, err error) {
	state := state.Instance

	log.Print("Polling Cluster Health...\n")
	if len(state.Kubeconfig.Contents) == 0 {
		if state.Kubeconfig.Contents, err = provider.ClusterKubeconfig(clusterID); err != nil {
			log.Printf("could not get kubeconfig for cluster: %v\n", err)
			return false, nil
		}
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(state.Kubeconfig.Contents)
	if err != nil {
		log.Printf("Error generating Rest Config: %v\n", err)
		return false, nil
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Printf("Error generating Kube Clientset: %v\n", err)
		return false, nil
	}

	oscfg, err := osconfig.NewForConfig(restConfig)
	if err != nil {
		log.Printf("Error generating OpenShift Clientset: %v\n", err)
		return false, nil
	}

	clusterHealthy := true

	var healthErr *multierror.Error
	if check, err := healthchecks.CheckCVOReadiness(oscfg.ConfigV1()); !check || err != nil {
		multierror.Append(healthErr, err)
		clusterHealthy = false
	}

	if check, err := healthchecks.CheckNodeHealth(kubeClient.CoreV1()); !check || err != nil {
		multierror.Append(healthErr, err)
		clusterHealthy = false
	}

	if check, err := healthchecks.CheckOperatorReadiness(oscfg.ConfigV1()); !check || err != nil {
		multierror.Append(healthErr, err)
		clusterHealthy = false
	}

	if check, err := healthchecks.CheckPodHealth(kubeClient.CoreV1()); !check || err != nil {
		multierror.Append(healthErr, err)
		clusterHealthy = false
	}

	if check, err := healthchecks.CheckCerts(kubeClient.CoreV1()); !check || err != nil {
		multierror.Append(healthErr, err)
		clusterHealthy = false
	}

	return clusterHealthy, healthErr.ErrorOrNil()
}
