package metrics

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/prometheus/common/model"
)

// ListAllAddonMetadata will list all addon metadata seen in the given time range.
func (c *Client) ListAllAddonMetadata(begin, end time.Time) ([]AddonMetadata, error) {
	results, err := c.issueQuery("cicd_addon_metadata", begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all metadata: %v", err)
	}

	return processAddonMetadataResults(results)
}

// ListAddonMetadataByJobNameAndJobID will list all addon metadata seen in the given time range using the given job name and job ID.
func (c *Client) ListAddonMetadataByJobNameAndJobID(jobName string, jobID int64, begin, end time.Time) ([]AddonMetadata, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_addon_metadata{job=\"%s\", job_id=\"%d\"}", escapeQuotes(jobName), jobID), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all metadata by job ID: %v", err)
	}

	return processAddonMetadataResults(results)
}

// ListAddonMetadataByClusterID will list all addon metadata seen in the given time range using the given provider, environment, and cluster ID.
func (c *Client) ListAddonMetadataByClusterID(cloudProvider, environment, clusterID string, begin, end time.Time) ([]AddonMetadata, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_addon_metadata{cloud_provider=\"%s\", environment=\"%s\", cluster_id=\"%s\"}",
		escapeQuotes(cloudProvider), escapeQuotes(environment), escapeQuotes(clusterID)), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all metadata by cluster ID: %v", err)
	}

	return processAddonMetadataResults(results)
}

func processAddonMetadataResults(results model.Value) ([]AddonMetadata, error) {
	addonMetadatas := []AddonMetadata{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			addonMetadata, err := sampleToAddonMetadata(sample)
			if err != nil {
				return nil, fmt.Errorf("error while getting event from Prometheus: %v", err)
			}
			addonMetadatas = append(addonMetadatas, addonMetadata)
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	sort.Sort(AddonMetadatas(addonMetadatas))

	return addonMetadatas, nil
}

func sampleToAddonMetadata(sample *model.SampleStream) (AddonMetadata, error) {
	metadata, err := sampleToMetadata(sample)
	if err != nil {
		return AddonMetadata{}, fmt.Errorf("error parsing out metadata: %v", err)
	}

	return AddonMetadata{
		Metadata: metadata,
		Phase:    stringToPhase(extractMetricFromSample(sample, "phase")),
	}, nil
}
