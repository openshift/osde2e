package metrics

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/prometheus/common/model"
)

// ListAllEvents will list all events seen in the given time range.
func (c *Client) ListAllEvents(begin, end time.Time) ([]Event, error) {
	results, err := c.issueQuery("cicd_event", begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all events: %v", err)
	}

	return c.processEventResults(results)
}

// ListEventsByJobNameAndJobID will list all events seen in the given time range using the given job ID.
func (c *Client) ListEventsByJobNameAndJobID(jobName string, jobID int64, begin, end time.Time) ([]Event, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_event{job=\"%s\", job_id=\"%d\"}", escapeQuotes(jobName), jobID), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all events by job ID: %v", err)
	}

	return c.processEventResults(results)
}

// ListEventsByClusterID will list all events seen in the given time range using the given cloud provider, environment, and cluster ID.
func (c *Client) ListEventsByClusterID(cloudProvider, environment, clusterID string, begin, end time.Time) ([]Event, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_event{cloud_provider=\"%s\", environment=\"%s\", cluster_id=\"%s\"}",
		escapeQuotes(cloudProvider), escapeQuotes(environment), escapeQuotes(clusterID)), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all events by cluster ID: %v", err)
	}

	return c.processEventResults(results)
}

func (c *Client) processEventResults(results model.Value) ([]Event, error) {
	events := []Event{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			event, err := sampleToEvent(sample)
			if err != nil {
				return nil, fmt.Errorf("error while getting event from Prometheus: %v", err)
			}
			events = append(events, event)
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	sort.Sort(Events(events))

	return events, nil
}

func sampleToEvent(sample *model.SampleStream) (Event, error) {
	installVersion, upgradeVersion, err := extractInstallAndUpgradeVersionsFromSample(sample)
	if err != nil {
		return Event{}, fmt.Errorf("error getting install and upgrade versions: %v", err)
	}

	jobIDString := extractMetricFromSample(sample, "job_id")

	if err != nil {
		return Event{}, fmt.Errorf("error getting job ID metric from sample: %v", err)
	}

	jobID, err := strconv.ParseInt(jobIDString, 0, 64)
	if err != nil {
		return Event{}, fmt.Errorf("error parsing job id: %v", err)
	}

	return Event{
		InstallVersion: installVersion,
		UpgradeVersion: upgradeVersion,
		CloudProvider:  extractMetricFromSample(sample, "cloud_provider"),
		Environment:    extractMetricFromSample(sample, "environment"),
		Event:          extractMetricFromSample(sample, "event"),
		ClusterID:      extractMetricFromSample(sample, "cluster_id"),
		JobName:        extractMetricFromSample(sample, "job"),
		JobID:          jobID,
		Timestamp:      pickFirstTimestamp(sample.Values),
	}, nil
}
