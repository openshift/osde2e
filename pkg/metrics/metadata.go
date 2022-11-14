package metrics

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/prometheus/common/model"
)

// ListAllMetadata will list all metadata seen in the given time range.
func (c *Client) ListAllMetadata(begin, end time.Time) ([]Metadata, error) {
	results, err := c.issueQuery("cicd_metadata", begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all metadata: %v", err)
	}

	return processMetadataResults(results)
}

// ListMetadataByJobNameAndJobID will list all metadata seen in the given time range using the given job name and job ID.
func (c *Client) ListMetadataByJobNameAndJobID(jobName string, jobID int64, begin, end time.Time) ([]Metadata, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_metadata{job=\"%s\", job_id=\"%d\"}", escapeQuotes(jobName), jobID), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all metadata by job ID: %v", err)
	}

	return processMetadataResults(results)
}

// ListMetadataByClusterID will list all metadata seen in the given time range using the given cloud provider, environment, and cluster ID.
func (c *Client) ListMetadataByClusterID(cloudProvider, environment, clusterID string, begin, end time.Time) ([]Metadata, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_metadata{cloud_provider=\"%s\", environment=\"%s\", cluster_id=\"%s\"}",
		escapeQuotes(cloudProvider), escapeQuotes(environment), escapeQuotes(clusterID)), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all metadata by cluster ID: %v", err)
	}

	return processMetadataResults(results)
}

func processMetadataResults(results model.Value) ([]Metadata, error) {
	metadatas := []Metadata{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			metadata, err := sampleToMetadata(sample)
			if err != nil {
				return nil, fmt.Errorf("error while getting event from Prometheus: %v", err)
			}
			metadatas = append(metadatas, metadata)
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	sort.Sort(Metadatas(metadatas))

	return metadatas, nil
}

func sampleToMetadata(sample *model.SampleStream) (Metadata, error) {
	installVersion, upgradeVersion, err := extractInstallAndUpgradeVersionsFromSample(sample)
	if err != nil {
		return Metadata{}, fmt.Errorf("error getting install and upgrade versions: %v", err)
	}

	jobID, err := strconv.ParseInt(extractMetricFromSample(sample, "job_id"), 0, 64)
	if err != nil {
		return Metadata{}, fmt.Errorf("error parsing job id: %v", err)
	}

	return Metadata{
		InstallVersion: installVersion,
		UpgradeVersion: upgradeVersion,
		CloudProvider:  extractMetricFromSample(sample, "cloud_provider"),
		Environment:    extractMetricFromSample(sample, "environment"),
		MetadataName:   extractMetricFromSample(sample, "metadata_name"),
		ClusterID:      extractMetricFromSample(sample, "cluster_id"),
		JobName:        extractMetricFromSample(sample, "job"),
		JobID:          jobID,
		Value:          averageValues(sample.Values),
		Timestamp:      pickFirstTimestamp(sample.Values),
	}, nil
}
