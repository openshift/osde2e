package osd

import (
	"fmt"

	"github.com/openshift-online/uhc-sdk-go/pkg/client/clustersmgmt/v1"
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
		resp, err := u.cluster(clusterId).
			Logs().
			Log(logId).
			Get().Parameter("tail", length).
			Send()

		if resp != nil {
			err = errResp(resp.Error())
		}

		if err != nil {
			return logs, fmt.Errorf("the contents of log '%s' couldn't be retrieved: %v", logId, err)
		}
		logs[logId] = []byte(resp.Body().Content())
	}
	return
}

func (u *OSD) getLogList(clusterId string) ([]string, error) {
	resp, err := u.cluster(clusterId).
		Logs().
		List().
		Send()

	if resp != nil {
		err = errResp(resp.Error())
	}

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve log list for cluster '%s': %v", clusterId, err)
	}

	var logs []string
	resp.Items().Each(func(l *v1.Log) bool {
		logs = append(logs, l.ID())
		return true
	})
	return logs, nil
}
