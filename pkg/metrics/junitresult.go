package metrics

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
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

// ListPassRatesByJobID will return a map of job IDs to their corresponding pass rates given a job name.
func (c *Client) ListPassRatesByJobID(jobName string, begin, end time.Time) (map[int64]float64, error) {
	results, err := c.ListJUnitResultsByJobName(jobName, begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing JUnit results while calculating pass rates: %v", err)
	}

	resultsByJobID := map[int64][]JUnitResult{}

	for _, result := range results {
		if _, ok := resultsByJobID[result.JobID]; !ok {
			resultsByJobID[result.JobID] = []JUnitResult{}
		}

		resultsByJobID[result.JobID] = append(resultsByJobID[result.JobID], result)
	}

	passRatesByJobID := map[int64]float64{}

	for jobID, jobIDResults := range resultsByJobID {
		passRatesByJobID[jobID] = calculatePassRates(jobIDResults)[jobName]
	}

	return passRatesByJobID, nil
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
		numPasses        int
		numTests         int
		upgradeFailed    bool
		upgradeTestsSeen bool
	}

	countsByJob := map[string]*counts{}

	for _, result := range results {
		// Ignore all log metrics results for calculating pass rates
		if strings.HasPrefix(result.TestName, "[Log Metrics]") {
			continue
		}

		if countsByJob[result.JobName] == nil {
			countsByJob[result.JobName] = &counts{}
		}

		if result.Result == Passed {
			countsByJob[result.JobName].numPasses++
		}

		if result.Result != Skipped {
			countsByJob[result.JobName].numTests++
		}

		// If the upgrade fails, we want to munge the number of total tests so that the passrate shows that installCount tests failed instead of just
		// this singular [upgrade] BeforeSuite test.
		if result.TestName == "[upgrade] BeforeSuite" {
			countsByJob[result.JobName].upgradeFailed = true
		} else if strings.HasPrefix(result.TestName, "[upgrade]") {
			countsByJob[result.JobName].upgradeTestsSeen = true
		}

	}

	passRates := map[string]float64{}

	for jobName, count := range countsByJob {
		if count.numTests == 0 {
			passRates[jobName] = 0
		} else {
			numTests := count.numTests

			// If an upgrade has failed, remove the upgrade BeforeSuite test failure and double the count of total tests.
			if count.upgradeFailed && !count.upgradeTestsSeen {
				numTests = (numTests - 1) * 2
			}

			passRates[jobName] = float64(count.numPasses) / float64(numTests)
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
