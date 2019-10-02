package osd

import (
	"fmt"
	"math"

	v1 "github.com/openshift-online/uhc-sdk-go/pkg/client/clustersmgmt/v1"
)

// Logs provides all logs available for clusterID, ids can be optionally provided for only specific logs.
func (u *OSD) Logs(clusterID string, length int, ids ...string) (logs map[string][]byte, err error) {
	if len(ids) == 0 {
		if ids, err = u.getLogList(clusterID); err != nil {
			return logs, fmt.Errorf("couldn't get log list: %v", err)
		}
	}

	logs = make(map[string][]byte, len(ids))
	for _, logID := range ids {
		resp, err := u.cluster(clusterID).
			Logs().
			Log(logID).
			Get().Parameter("tail", length).
			Send()

		if resp != nil {
			err = errResp(resp.Error())
		}

		if err != nil {
			return logs, fmt.Errorf("the contents of log '%s' couldn't be retrieved: %v", logID, err)
		}
		logs[logID] = []byte(resp.Body().Content())
	}
	return
}

// FullLogs returns as much Logs as it can.
func (u *OSD) FullLogs(clusterID string, ids ...string) (map[string][]byte, error) {
	return u.Logs(clusterID, math.MaxInt32-1, ids...)
}

func (u *OSD) getLogList(clusterID string) ([]string, error) {
	resp, err := u.cluster(clusterID).
		Logs().
		List().
		Send()

	if resp != nil {
		err = errResp(resp.Error())
	}

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve log list for cluster '%s': %v", clusterID, err)
	}

	var logs []string
	resp.Items().Each(func(l *v1.Log) bool {
		logs = append(logs, l.ID())
		return true
	})
	return logs, nil
}
