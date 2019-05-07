package osd

import (
	"encoding/json"
	"fmt"
)

// Logs provides all logs available for a cluster, ids can be optionally provided for only specific logs.
func (u *OSD) Logs(clusterId string, length int, ids ...string) (logs map[string][]byte, err error) {
	if ids == nil || len(ids) == 0 {
		if ids, err = u.getLogList(clusterId); err != nil {
			return logs, fmt.Errorf("couldn't get log list: %v", err)
		}
	}

	logs = make(map[string][]byte, len(ids))
	for _, logId := range ids {
		params := map[string]interface{}{"tail": length}
		resource := fmt.Sprintf("clusters/%s/logs/%s", clusterId, logId)
		resp, err := doRequest(u.conn, "", resource, params, nil)
		if err != nil {
			return logs, fmt.Errorf("couldn't retrieve log list for cluster '%s': %v", clusterId, err)
		}

		body := map[string]interface{}{}
		if err = json.Unmarshal(resp.Bytes(), &body); err != nil {
			return logs, fmt.Errorf("couldn't unmarshal response: %v", err)
		}

		contentStr, err := getStr(body, "content")
		if err != nil {
			return logs, fmt.Errorf("the cotents of log '%s' couldn't be retrieved: %v", logId, err)
		}
		logs[logId] = []byte(contentStr)
	}
	return
}

func (u *OSD) getLogList(clusterId string) ([]string, error) {
	resource := fmt.Sprintf("clusters/%s/logs", clusterId)
	resp, err := doRequest(u.conn, "", resource, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve log list for cluster '%s': %v", clusterId, err)
	}

	body := map[string]interface{}{}
	if err = json.Unmarshal(resp.Bytes(), &body); err != nil {
		return []string{}, fmt.Errorf("couldn't unmarshal response: %v", err)
	}

	return getListOfField(body, "items", "id")
}
