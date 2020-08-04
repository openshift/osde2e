package metrics

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/prometheus/common/model"
)

// ListAllJUnitResults will return all JUnitResults in the given time range.
func (c *Client) ListAllJUnitResults(begin, end time.Time) ([]JUnitResult, error) {
	results, err := c.issueQuery("cicd_jUnitResult", begin, end)

	if err != nil {
		return nil, fmt.Errorf("error listing all JUnit results: %v", err)
	}

	return processJUnitResults(results)
}

// ListPassRatesByJob will return a map of job names to their corresponding pass rates.
func (c *Client) ListPassRatesByJob(begin, end time.Time) (map[string]float64, error) {
	results, err := c.ListAllJUnitResults(begin, end)

	if err != nil {
		return nil, fmt.Errorf("error listing JUnit results while calculating pass rates: %v", err)
	}

	return calculatePassRates(results), nil
}

// GetPassRateForJob will return the pass rate for a given job.
func (c *Client) GetPassRateForJob(jobName string, begin, end time.Time) (float64, error) {
	results, err := c.ListJUnitResultsByJobName(jobName, begin, end)

	if err != nil {
		return 0, fmt.Errorf("error listing JUnit results for job %s while calculating pass rates: %v", jobName, err)
	}

	return calculatePassRates(results)[jobName], nil
}

// ListJUnitResultsByJobName will return all JUnitResults in the given time range for the given job name across job IDs.
func (c *Client) ListJUnitResultsByJobName(jobName string, begin, end time.Time) ([]JUnitResult, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_jUnitResult{job=\"%s\"}", escapeQuotes(jobName)), begin, end)

	if err != nil {
		return nil, fmt.Errorf("error listing all JUnit results: %v", err)
	}

	return processJUnitResults(results)
}

// ListJUnitResultsByJobNameAndJobID will return all JUnitResults in the given time range for the given job name and ID.
func (c *Client) ListJUnitResultsByJobNameAndJobID(jobName string, jobID int64, begin, end time.Time) ([]JUnitResult, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_jUnitResult{job=\"%s\", job_id=\"%d\"}", escapeQuotes(jobName), jobID), begin, end)

	if err != nil {
		return nil, fmt.Errorf("error listing all JUnit results: %v", err)
	}

	return processJUnitResults(results)
}

// ListJUnitResultsByClusterID will return all JUnitResults in the given time range for the given cloud provider, environment, and cluster ID.
func (c *Client) ListJUnitResultsByClusterID(cloudProvider, environment, clusterID string, begin, end time.Time) ([]JUnitResult, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_jUnitResult{cloud_provider=\"%s\", environment=\"%s\", cluster_id=\"%s\"}",
		escapeQuotes(cloudProvider), escapeQuotes(environment), escapeQuotes(clusterID)), begin, end)

	if err != nil {
		return nil, fmt.Errorf("error listing all JUnit results: %v", err)
	}

	return processJUnitResults(results)
}

// ListFailedJUnitResultsByTestName will return all JUnitResults in a given time range for a given test name.
func (c *Client) ListFailedJUnitResultsByTestName(testName string, begin, end time.Time) ([]JUnitResult, error) {
	results, err := c.issueQuery(fmt.Sprintf("cicd_jUnitResult{result=\"failed\", testname=~\".*%s.*\"}", escapeQuotes(testName)), begin, end)

	if err != nil {
		return nil, fmt.Errorf("error listing all JUnit results: %v", err)
	}

	return processJUnitResults(results)
}

func calculatePassRates(results []JUnitResult) map[string]float64 {
	type counts struct {
		numPasses int
		numTests  int
	}

	countsByJob := map[string]*counts{}

	for _, result := range results {
		if countsByJob[result.JobName] == nil {
			countsByJob[result.JobName] = &counts{}
		}

		if result.Result == Passed {
			countsByJob[result.JobName].numPasses++
		}

		if result.Result != Skipped {
			countsByJob[result.JobName].numTests++
		}
	}

	passRates := map[string]float64{}

	for jobName, count := range countsByJob {
		if count.numTests == 0 {
			passRates[jobName] = 0
		} else {
			passRates[jobName] = float64(count.numPasses) / float64(count.numTests)
		}
	}

	return passRates
}

func processJUnitResults(results model.Value) ([]JUnitResult, error) {
	jUnitResults := []JUnitResult{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			jUnitResult, err := sampleToJUnitResult(sample)

			if err != nil {
				return nil, fmt.Errorf("error while getting event from Prometheus: %v", err)
			}
			jUnitResults = append(jUnitResults, jUnitResult)
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	sort.Sort(JUnitResults(jUnitResults))

	return jUnitResults, nil
}

func sampleToJUnitResult(sample *model.SampleStream) (JUnitResult, error) {
	installVersion, upgradeVersion, err := extractInstallAndUpgradeVersionsFromSample(sample)

	if err != nil {
		return JUnitResult{}, fmt.Errorf("error getting install and upgrade versions: %v", err)
	}

	jobID, err := strconv.ParseInt(extractMetricFromSample(sample, "job_id"), 0, 64)

	if err != nil {
		return JUnitResult{}, fmt.Errorf("error parsing job id: %v", err)
	}

	return JUnitResult{
		InstallVersion: installVersion,
		UpgradeVersion: upgradeVersion,
		CloudProvider:  extractMetricFromSample(sample, "cloud_provider"),
		Environment:    extractMetricFromSample(sample, "environment"),
		Suite:          extractMetricFromSample(sample, "suite"),
		TestName:       extractMetricFromSample(sample, "testname"),
		Result:         stringToResult(extractMetricFromSample(sample, "result")),
		ClusterID:      extractMetricFromSample(sample, "cluster_id"),
		JobName:        extractMetricFromSample(sample, "job"),
		JobID:          jobID,
		Phase:          stringToPhase(extractMetricFromSample(sample, "phase")),
		Duration:       time.Duration(averageValues(sample.Values)) * time.Second,
		Timestamp:      pickFirstTimestamp(sample.Values),
	}, nil
}
