package main

import (
	"fmt"
)

const (
	StagingURL = "https://api.stage.openshift.com"
	APIPrefix  = "/api/clusters_mgmt"
	APIVersion = "v1"
)

func LaunchCluster(name, awsId, awsKey string) (interface{}, error) {
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

	return createClusterReq(cluster)
}

func ClusterState(clusterId string) (string, error) {
	cluster, err := getClusterReq(clusterId)
	if err != nil {
		return "", fmt.Errorf("couldn't get cluster '%s': %v", clusterId, err)
	}

	state, err := getStr(cluster, "state")
	if err != nil {
		return "", fmt.Errorf("couldn't get cluster state for '%s': %v", clusterId, err)
	}

	return state, nil
}

func ClusterKubeconfig(clusterId string) (kubeconfig []byte, err error) {
	creds, err := getCredentialsReq(clusterId)
	if err != nil {
		return nil, err
	}

	kubeconfigStr, err := getStr(creds, "kubeconfig")
	if err != nil {
		kubeconfig = []byte(kubeconfigStr)
	}
	return kubeconfig, err
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
