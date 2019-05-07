package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const (
	// ClusterStateReady is returned by UHC when a cluster is ready for operations.
	ClusterStateReady = "ready"

	// ClusterStateError is returned by UHC when there is an unrecoverable problem.
	ClusterStateError = "error"
)

// LaunchCluster setups an new cluster using the UHC API and returns it's ID.
func (u *UHC) LaunchCluster(name, version, awsId, awsKey string) (string, error) {
	log.Printf("Creating cluster '%s'...", name)
	cluster := map[string]interface{}{
		"name": name,
		"aws": map[string]interface{}{
			"access_key_id":     awsId,
			"secret_access_key": awsKey,
		},
		"dns": map[string]interface{}{
			"base_domain": "devcluster.openshift.com",
		},
		"flavour": map[string]interface{}{
			"id": "4",
		},
		"region": map[string]interface{}{
			"id": "us-east-1",
		},
		"version": map[string]interface{}{
			"id": version,
		},
	}

	params := map[string]interface{}{"provision": true}
	resp, err := doRequest(u.conn, "POST", "clusters", params, cluster)
	if err != nil {
		return "", fmt.Errorf("couldn't create cluster: %v", err)
	}

	var newCluster interface{}
	err = json.Unmarshal(resp.Bytes(), &newCluster)

	return getStr(newCluster, "id")
}

// GetCluster returns the information about clusterId.
func (u *UHC) GetCluster(clusterId string) (interface{}, error) {
	resource := fmt.Sprintf("clusters/%s", clusterId)
	resp, err := doRequest(u.conn, "", resource, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve cluster '%s': %v", clusterId, err)
	}

	var cluster interface{}
	err = json.Unmarshal(resp.Bytes(), &cluster)
	return cluster, err
}

// ClusterState retrieves the state of clusterId.
func (u *UHC) ClusterState(clusterId string) (string, error) {
	cluster, err := u.GetCluster(clusterId)
	if err != nil {
		return "", fmt.Errorf("couldn't get cluster '%s': %v", clusterId, err)
	}

	state, err := getStr(cluster, "state")
	if err != nil {
		return "", fmt.Errorf("couldn't get cluster state for '%s': %v", clusterId, err)
	}

	return state, nil
}

// ClusterKubeconfig retrieves the kubeconfig of clusterId.
func (u *UHC) ClusterKubeconfig(clusterId string) (kubeconfig []byte, err error) {
	resource := fmt.Sprintf("clusters/%s/credentials", clusterId)
	resp, err := doRequest(u.conn, "", resource, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve credentials for cluster '%s': %v", clusterId, err)
	}

	creds := map[string]interface{}{}
	err = json.Unmarshal(resp.Bytes(), &creds)

	kubeconfigStr, err := getStr(creds, "kubeconfig")
	if err == nil {
		kubeconfig = []byte(kubeconfigStr)
	}
	return kubeconfig, err
}

// DeleteCluster requests the deletion of clusterID.
func (u *UHC) DeleteCluster(clusterId string) error {
	resource := fmt.Sprintf("clusters/%s", clusterId)
	_, err := doRequest(u.conn, "DELETE", resource, nil, nil)
	if err != nil {
		return fmt.Errorf("couldn't delete cluster '%s': %v", clusterId, err)
	}
	return nil
}

// WaitForClusterReady blocks until clusterId is ready or a number of retries has been attempted.
func (u *UHC) WaitForClusterReady(clusterId string) error {
	times, wait := 145, 45*time.Second
	log.Printf("Waiting %v for cluster '%s' to be ready...\n", time.Duration(times)*wait, clusterId)

	for i := 0; i < times; i++ {
		if state, err := u.ClusterState(clusterId); state == ClusterStateReady {
			return nil
		} else if err != nil {
			log.Print("Encountered error waiting for cluster:", err)
		} else if state == ClusterStateError {
			return fmt.Errorf("the installation of cluster '%s' has errored", clusterId)
		} else {
			log.Printf("Cluster is not ready, current status '%s'.", state)
		}

		time.Sleep(wait)
	}

	time.Sleep(time.Second)
	return fmt.Errorf("timed out waiting for cluster '%s' to be ready", clusterId)
}
