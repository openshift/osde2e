package osd

import (
	"fmt"
	"log"
	"time"

	v1 "github.com/openshift-online/uhc-sdk-go/pkg/client/clustersmgmt/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	osconfig "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// DefaultFlavour is used when no specialized configuration exists.
	DefaultFlavour = "4"
)

// LaunchCluster setups an new cluster using the OSD API and returns it's ID.
func (u *OSD) LaunchCluster(cfg *config.Config) (string, error) {
	log.Printf("Creating cluster '%s'...", cfg.ClusterName)

	// choose flavour based on config
	flavourID := u.Flavour(cfg)

	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	expiration := time.Now().Add(8 * time.Hour)

	cluster, err := v1.NewCluster().
		Name(cfg.ClusterName).
		Flavour(v1.NewFlavour().
			ID(flavourID)).
		Region(v1.NewCloudRegion().
			ID("us-east-1")).
		MultiAZ(cfg.MultiAZ).
		Version(v1.NewVersion().
			ID(cfg.ClusterVersion)).
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

	if resp != nil {
		err = errResp(resp.Error())
	}

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve cluster '%s': %v", clusterID, err)
	}
	return resp.Body(), err
}

// Flavour returns the default flavour for cfg.
func (u *OSD) Flavour(cfg *config.Config) string {
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
func (u *OSD) WaitForClusterReady(cfg *config.Config) error {
	log.Printf("Waiting %v for cluster '%s' to be ready...\n", cfg.ClusterUpTimeout, cfg.ClusterID)
	cleanRuns := 0

	return wait.PollImmediate(30*time.Second, cfg.ClusterUpTimeout, func() (bool, error) {
		if state, err := u.ClusterState(cfg.ClusterID); state == v1.ClusterStateReady {
			healthy, err := u.PollClusterHealth(cfg)
			if err != nil {
				return false, err
			}

			if !healthy {
				cleanRuns = 0
				return false, nil
			}
			cleanRuns++

			return cleanRuns >= 5, nil

		} else if err != nil {
			log.Print("Encountered error waiting for cluster:", err)
		} else if state == v1.ClusterStateError {
			return false, fmt.Errorf("the installation of cluster '%s' has errored\n", cfg.ClusterID)
		} else {
			log.Printf("Cluster is not ready, current status '%s'.\n", state)
		}
		return false, nil
	})
}

// PollClusterHealth looks at CVO data to determine if a cluster is alive/healthy or not
func (u *OSD) PollClusterHealth(cfg *config.Config) (status bool, err error) {
	log.Print("Polling Cluster Health...\n")
	if cfg.Kubeconfig == nil {
		if cfg.Kubeconfig, err = u.ClusterKubeconfig(cfg.ClusterID); err != nil {
			log.Printf("could not get kubeconfig for cluster: %v\n", err)
			return false, nil
		}
	}
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(cfg.Kubeconfig)
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
	cv, getOpts := oscfg.ConfigV1(), metav1.GetOptions{}
	cvInfo, err := cv.ClusterVersions().Get("version", getOpts)
	if err != nil {
		log.Printf("Error getting CVS: %v\n", err)
		return false, nil
	}

	for _, v := range cvInfo.Status.Conditions {
		if v.Type == "Available" && v.Status != "True" {
			log.Printf("API Server troubles: %v\n", v.Message)
			return false, nil
		}
		if v.Type == "Progressing" && v.Status != "False" {
			log.Printf("Installation or upgrade in progress: %v\n", v.Message)
			return false, nil
		}
		if v.Type == "Failing" && v.Status != "False" {
			log.Printf("Cluster is in a failed state: %v\n", v.Message)
			return false, nil
		}
	}

	if check, err := helper.CheckNodeHealth(kubeClient); !check || err != nil {
		return false, nil
	}

	if check, err := helper.CheckOperatorReadiness(oscfg); !check || err != nil {
		return false, nil
	}

	if check, err := helper.CheckPodHealth(kubeClient); !check || err != nil {
		return false, nil
	}

	return true, nil
}
