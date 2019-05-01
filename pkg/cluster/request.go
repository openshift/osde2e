package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"

	uhc "github.com/openshift-online/uhc-sdk-go/pkg/client"
)

func createClusterReq(conn *uhc.Connection, cluster interface{}) (interface{}, error) {
	params := map[string]interface{}{"provision": true}
	resp, err := doRequest(conn, "POST", "clusters", params, cluster)
	if err != nil {
		return nil, fmt.Errorf("couldn't create cluster: %v", err)
	}

	var createdCluster interface{}
	err = json.Unmarshal(resp.Bytes(), &createdCluster)
	return createdCluster, err
}

func getClusterReq(conn *uhc.Connection, clusterId string) (interface{}, error) {
	resource := fmt.Sprintf("clusters/%s", clusterId)
	resp, err := doRequest(conn, "", resource, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve cluster '%s': %v", clusterId, err)
	}

	var cluster interface{}
	err = json.Unmarshal(resp.Bytes(), &cluster)
	return cluster, err
}

func deleteClusterReq(conn *uhc.Connection, clusterId string) error {
	resource := fmt.Sprintf("clusters/%s", clusterId)
	resp, err := doRequest(conn, "DELETE", resource, nil, nil)
	if err != nil {
		return fmt.Errorf("couldn't delete cluster '%s': %v", clusterId, err)
	}

	if resp.Status() != http.StatusOK {
		return fmt.Errorf("encountered error (%d) deleting '%s': %v", resp.Status(), clusterId, err)
	}
	return nil
}

func getCredentialsReq(conn *uhc.Connection, clusterId string) (map[string]interface{}, error) {
	resource := fmt.Sprintf("clusters/%s/credentials", clusterId)
	resp, err := doRequest(conn, "", resource, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve credentials for cluster '%s': %v", clusterId, err)
	}

	creds := map[string]interface{}{}
	err = json.Unmarshal(resp.Bytes(), creds)
	return creds, err
}

func doRequest(conn *uhc.Connection, method, resource string, param map[string]interface{}, msg interface{}) (resp *uhc.Response, err error) {
	var data []byte
	if msg != nil {
		// marshal body unless bytes
		switch in := msg.(type) {
		case []byte:
			data = in
		default:
			data, err = json.Marshal(msg)
			if err != nil {
				return nil, fmt.Errorf("couldn't marshal request body: %v", err)
			}
		}
	}

	// set method type
	var req *uhc.Request
	switch method {
	case "POST":
		req = conn.Post()
	case "DELETE":
		req = conn.Delete()
	default:
		req = conn.Get()
	}

	// set path and payload
	req = req.Path(APIPrefix + "/" + APIVersion + "/" + resource)
	if msg != nil {
		req.Bytes(data)
	}

	// set params
	for k, v := range param {
		req.Parameter(k, v)
	}

	return req.Send()
}
