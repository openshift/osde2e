package ocmprovider

import (
	"fmt"
	"math"
	"strings"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// Logs provides all logs available for clusterID, ids can be optionally provided for only specific logs.
func (o *OCMProvider) Logs(clusterID string) (logs map[string][]byte, err error) {
	var ids []string
	if ids, err = o.getLogList(clusterID); err != nil {
		return logs, fmt.Errorf("couldn't get log list: %v", err)
	}

	logs = make(map[string][]byte, len(ids))
	for _, logID := range ids {
		var resp *v1.LogGetResponse

		found := false
		err = retryer().Do(func() error {
			var err error
			resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
				Logs().Install().
				Get().Parameter("tail", math.MaxInt32-1).
				Send()

			if err != nil {
				// Log is just not found, so skip this log
				if strings.Contains(err.Error(), "'404'") {
					return nil
				}
				return err
			}

			if resp != nil && resp.Error() != nil {
				// Log is just not found, so skip this log
				if resp.Error().ID() == "404" {
					return nil
				}

				return errResp(resp.Error())
			}

			found = true
			return nil
		})

		if err != nil {
			return logs, fmt.Errorf("the contents of log '%s' couldn't be retrieved: %v", logID, err)
		}

		if found {
			logs[logID] = []byte(resp.Body().Content())
		}
	}
	return
}

func (o *OCMProvider) getLogList(clusterID string) ([]string, error) {
	var resp *v1.LogsListResponse

	err := retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
			Logs().
			List().
			Send()

		if err != nil {
			return err
		}

		if resp != nil && resp.Error() != nil {
			return errResp(resp.Error())
		}

		return nil
	})
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
