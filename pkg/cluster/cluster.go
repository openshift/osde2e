package cluster

import (
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
func (u *UHC) LaunchCluster(name, awsId, awsKey string) (string, error) {
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
			"id": "openshift-v4.0-beta4",
		},
	}

	newCluster, err := createClusterReq(u.conn, cluster)
	if err != nil {
		return "", err
	}

	return getStr(newCluster, "id")
}

func (u *UHC) ClusterState(clusterId string) (string, error) {
	cluster, err := getClusterReq(u.conn, clusterId)
	if err != nil {
		return "", fmt.Errorf("couldn't get cluster '%s': %v", clusterId, err)
	}

	state, err := getStr(cluster, "state")
	if err != nil {
		return "", fmt.Errorf("couldn't get cluster state for '%s': %v", clusterId, err)
	}

	return state, nil
}

func (u *UHC) ClusterKubeconfig(clusterId string) (kubeconfig []byte, err error) {
	creds, err := getCredentialsReq(u.conn, clusterId)
	if err != nil {
		return nil, err
	}

	kubeconfigStr, err := getStr(creds, "kubeconfig")
	if err != nil {
		kubeconfig = []byte(kubeconfigStr)
	}
	return kubeconfig, err
}

func (u *UHC) DeleteCluster(clusterId string) error {
	if err := deleteClusterReq(u.conn, clusterId); err != nil {
		return fmt.Errorf("failed to destroy cluster '%s': %v", clusterId, err)
	}
	return nil
}

func (u *UHC) WaitForClusterReady(clusterId string) error {
	times, wait := 45, 45*time.Second
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

func getStr(d interface{}, k string) (string, error) {
	m, ok := d.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected a map: %+v", d)
	}

	v, ok := m[k]
	if !ok || v == nil {
		return "", fmt.Errorf("key '%s' is not set", k)
	}

	str, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value for key '%s' was not a string", k)
	}

	return str, nil
}
