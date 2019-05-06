package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"

	uhc "github.com/openshift-online/uhc-sdk-go/pkg/client"
)

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
	expectedStatus := http.StatusOK
	switch method {
	case "POST":
		req = conn.Post()
		expectedStatus = http.StatusCreated
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

	resp, err = req.Send()
	if err == nil && resp.Status() != expectedStatus {
		err = errResp(resp)
	}
	return
}

type mgmtError struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

func errResp(resp *uhc.Response) error {
	errResp := new(mgmtError)
	if err := json.Unmarshal(resp.Bytes(), errResp); err != nil {
		return fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	return fmt.Errorf("api error: %s", errResp.Reason)
}

func getListOfField(d interface{}, k, field string) (idList []string, err error) {
	v, err := getVal(d, k)
	if err != nil {
		return idList, err
	}

	list, ok := v.([]interface{})
	if !ok {
		return idList, fmt.Errorf("value for key %s' is not a list'", k)
	}

	for _, i := range list {
		str, err := getStr(i, field)
		if err != nil {
			return idList, err
		}

		idList = append(idList, str)
	}

	return
}

func getStr(d interface{}, k string) (string, error) {
	v, err := getVal(d, k)
	if err != nil {
		return "", err
	}

	str, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value for key '%s' was not a string", k)
	}

	return str, nil
}

func getVal(d interface{}, k string) (interface{}, error) {
	m, ok := d.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected a map: %+v", d)
	}

	v, ok := m[k]
	if !ok || v == nil {
		return "", fmt.Errorf("key '%s' is not set", k)
	}
	return v, nil
}
