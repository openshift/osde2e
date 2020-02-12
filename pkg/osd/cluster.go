package osd

import (
	"fmt"
	"log"
	"time"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	osconfig "github.com/openshift/client-go/config/clientset/versioned"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/helper"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/state"
)

const (
	// DefaultFlavour is used when no specialized configuration exists.
	DefaultFlavour = "osd-4"

	// cleanRunWindow is the number of checks made to determine if a cluster is ready.
	cleanRunWindow = 20

	// errorWindow is the number of checks made to determine if a cluster has truly failed.
	errorWindow = 5
)

// LaunchCluster setups an new cluster using the OSD API and returns it's ID.
func (u *OSD) LaunchCluster() (string, error) {
	cfg := config.Instance
	state := state.Instance

	log.Printf("Creating cluster '%s'...", state.Cluster.Name)

	// choose flavour based on config
	flavourID := u.Flavour()

	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	expiration := time.Now().Add(time.Duration(cfg.Cluster.ExpiryInMinutes) * time.Minute).UTC() // UTC() to workaround SDA-1567

	cluster, err := v1.NewCluster().
		Name(state.Cluster.Name).
		Flavour(v1.NewFlavour().
			ID(flavourID)).
		Region(v1.NewCloudRegion().
			ID("us-east-1")).
		MultiAZ(cfg.Cluster.MultiAZ).
		Version(v1.NewVersion().
			ID(state.Cluster.Version)).
		ExpirationTimestamp(expiration).
		Build()
	if err != nil {
		return "", fmt.Errorf("couldn't build cluster description: %v", err)
	}

	resp, err := u.clusters().Add().
		Body(cluster).
		Send()

	if resp != nil {
		err = errResp(resp.Error())
	}

	if err != nil {
		return "", fmt.Errorf("couldn't create cluster: %v", err)
	}
	return resp.Body().ID(), nil
}

// GetCluster returns the information about clusterID.
func (u *OSD) GetCluster(clusterID string) (*v1.Cluster, error) {
	resp, err := u.cluster(clusterID).
		Get().
		Send()

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve cluster '%s': %v", clusterID, err)
	}

	if resp.Error() != nil {
		return resp.Body(), resp.Error()
	}

	return resp.Body(), err
}

// Flavour returns the default flavour for cfg.
func (u *OSD) Flavour() string {
	return DefaultFlavour
}

// ClusterState retrieves the state of clusterID.
func (u *OSD) ClusterState(clusterID string) (v1.ClusterState, error) {
	cluster, err := u.GetCluster(clusterID)
	if err != nil {
		return "", fmt.Errorf("couldn't get cluster '%s': %v", clusterID, err)
	}
	return cluster.State(), nil
}

// InstallAddons loops through the addons list in the config
// and performs the CRUD operation to trigger addon installation
func (u *OSD) InstallAddons(addonIDs []string) (num int, err error) {
	num = 0
	addonsClient := u.addons()
	clusterClient := u.cluster(state.Instance.Cluster.ID)
	for _, addonID := range addonIDs {
		addonResp, err := addonsClient.Addon(addonID).Get().Send()
		if err != nil {
			return 0, err
		}
		addon := addonResp.Body()

		if addon.Enabled() {
			addonInstallation, err := v1.NewAddOnInstallation().Addon(v1.NewAddOn().Copy(addon)).Build()
			if err != nil {
				return 0, err
			}

			aoar, err := clusterClient.Addons().Add().Body(addonInstallation).Send()
			if err != nil {
				return 0, err
			}

			if aoar.Error() != nil {
				return 0, fmt.Errorf("Error (%v) sending request: %v", aoar.Status(), aoar.Error())
			}

			num++
		}
	}

	return num, nil
}

// ClusterKubeconfig retrieves the kubeconfig of clusterID.
func (u *OSD) ClusterKubeconfig(clusterID string) (kubeconfig []byte, err error) {
	resp, err := u.cluster(clusterID).
		Credentials().
		Get().
		Send()

	if resp != nil {
		err = errResp(resp.Error())
	}

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve credentials for cluster '%s': %v", clusterID, err)
	}
	return []byte(resp.Body().Kubeconfig()), nil
}

// DeleteCluster requests the deletion of clusterID.
func (u *OSD) DeleteCluster(clusterID string) error {
	resp, err := u.cluster(clusterID).
		Delete().
		Send()

	if resp != nil {
		err = errResp(resp.Error())
	}

	if err != nil {
		return fmt.Errorf("couldn't delete cluster '%s': %v", clusterID, err)
	}
	return nil
}

// WaitForClusterReady blocks until clusterID is ready or a number of retries has been attempted.
func (u *OSD) WaitForClusterReady() error {
	cfg := config.Instance
	state := state.Instance

	log.Printf("Waiting %v minutes for cluster '%s' to be ready...\n", cfg.Cluster.InstallTimeout, state.Cluster.ID)
	cleanRuns := 0
	errRuns := 0

	clusterStarted := time.Now()
	var readinessStarted time.Time
	ocmReady := false
	if !cfg.Tests.SkipClusterHealthChecks {
		return wait.PollImmediate(30*time.Second, time.Duration(cfg.Cluster.InstallTimeout)*time.Minute, func() (bool, error) {
			if clusterState, err := u.ClusterState(state.Cluster.ID); clusterState == v1.ClusterStateReady {
				// This is the first time that we've entered this section, so we'll consider this the time until OCM has said the cluster is ready
				if !ocmReady {
					ocmReady = true
					if metadata.Instance.TimeToOCMReportingInstalled == 0 {
						metadata.Instance.SetTimeToOCMReportingInstalled(time.Since(clusterStarted).Seconds())
					}

					readinessStarted = time.Now()
				}
				if success, err := u.PollClusterHealth(); success {
					cleanRuns++
					log.Printf("Clean run %d/%d...", cleanRuns, cleanRunWindow)
					errRuns = 0
					if cleanRuns == cleanRunWindow {
						if metadata.Instance.TimeToClusterReady == 0 {
							metadata.Instance.SetTimeToClusterReady(time.Since(readinessStarted).Seconds())
						} else {
							metadata.Instance.SetTimeToClusterUpgraded(time.Since(readinessStarted).Seconds())
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
			} else if clusterState == v1.ClusterStateError {
				return false, fmt.Errorf("the installation of cluster '%s' has errored", state.Cluster.ID)
			} else {
				log.Printf("Cluster is not ready, current status '%s'.", clusterState)
			}
			return false, nil
		})
	}
	return nil
}

// PollClusterHealth looks at CVO data to determine if a cluster is alive/healthy or not
func (u *OSD) PollClusterHealth() (status bool, err error) {
	state := state.Instance

	log.Print("Polling Cluster Health...\n")
	if len(state.Kubeconfig.Contents) == 0 {
		if state.Kubeconfig.Contents, err = u.ClusterKubeconfig(state.Cluster.ID); err != nil {
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

	if check, err := helper.CheckCVOReadiness(oscfg.ConfigV1()); !check || err != nil {
		return false, nil
	}

	if check, err := helper.CheckNodeHealth(kubeClient.CoreV1()); !check || err != nil {
		return false, nil
	}

	if check, err := helper.CheckOperatorReadiness(oscfg.ConfigV1()); !check || err != nil {
		return false, nil
	}

	if check, err := helper.CheckPodHealth(kubeClient.CoreV1()); !check || err != nil {
		return false, nil
	}

	return true, nil
}
