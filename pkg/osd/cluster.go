package osd

import (
	"fmt"
	"log"
	"time"

	"github.com/openshift-online/uhc-sdk-go/pkg/client/clustersmgmt/v1"

	"github.com/openshift/osde2e/pkg/config"
)

// LaunchCluster setups an new cluster using the OSD API and returns it's ID.
func (u *OSD) LaunchCluster(cfg *config.Config) (string, error) {
	log.Printf("Creating cluster '%s'...", cfg.ClusterName)

	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	expiration := time.Now().Add(8 * time.Hour)

	cluster, err := v1.NewCluster().
		Name(cfg.ClusterName).
		Flavour(v1.NewFlavour().
			ID("4")).
		Region(v1.NewCloudRegion().
			ID("us-east-1")).
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
func (u *OSD) WaitForClusterReady(clusterID string) error {
	times, wait := 145, 45*time.Second
	log.Printf("Waiting %v for cluster '%s' to be ready...\n", time.Duration(times)*wait, clusterID)

	for i := 0; i < times; i++ {
		if state, err := u.ClusterState(clusterID); state == v1.ClusterStateReady {
			return nil
		} else if err != nil {
			log.Print("Encountered error waiting for cluster:", err)
		} else if state == v1.ClusterStateError {
			return fmt.Errorf("the installation of cluster '%s' has errored", clusterID)
		} else {
			log.Printf("Cluster is not ready, current status '%s'.", state)
		}

		time.Sleep(wait)
	}

	time.Sleep(time.Second)
	return fmt.Errorf("timed out waiting for cluster '%s' to be ready", clusterID)
}
