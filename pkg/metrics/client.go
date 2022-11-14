package metrics

import (
	"context"
	"fmt"
	"log"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/prometheus"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

const (
	maxQueryTimeoutInSeconds = "osde2e.metricsLib.maxQueryTimeoutInSeconds"

	stepDurationInHours = "osde2e.metricsLib.stepDurationInHours"
)

func init() {
	// Set our max query timeout to 2 minutes for now.
	viper.SetDefault(maxQueryTimeoutInSeconds, 120)
	viper.BindEnv(maxQueryTimeoutInSeconds, "OSDE2E_METRICSLIB_MAX_QUERY_TIMEOUT_IN_SECONDS")

	// Hard code our step duration to 4 for now. Our jobs are pretty coarse -- running every 4+ hours.
	// We'll bake this into our client to prevent our users from getting oversampled data.
	viper.SetDefault(stepDurationInHours, 4)
	viper.BindEnv(stepDurationInHours, "OSDE2E_METRICSLIB_STEP_DURATION_IN_HOURS")
}

// Client is a metrics client that can be used to query osde2e's metrics.
type Client struct {
	client api.Client
}

// NewClient returns a new metrics client.
// If no arguments are supplied, the global config will be used.
// If one argument is supplied, it will be used as the address for Prometheus, but will use the global config for the bearer token.
// If two arguments are supplied, the first will be used as the address for Prometheus and the second will be used as the bearer token.
func NewClient(args ...string) (*Client, error) {
	client, err := prometheus.CreateClient(args...)
	if err != nil {
		return nil, fmt.Errorf("error trying to create the metrics client: %v", err)
	}

	return &Client{
		client: client,
	}, nil
}

// ListAllJobNames will give a list of all of the osde2e jobs names seen in the given range.
func (c *Client) ListAllJobNames(begin, end time.Time) ([]string, error) {
	results, err := c.issueQuery("count by (job) (cicd_jUnitResult)", begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing all jobs: %v", err)
	}

	jobNames := []string{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			if err != nil {
				return nil, fmt.Errorf("error getting job name from sample: %v", err)
			}
			jobNames = append(jobNames, extractMetricFromSample(sample, "job"))
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	return sort.StringSlice(jobNames), nil
}

// ListAllJobIDs will list all of the individual job IDs (individual job runs) for a given job in the given range.
func (c *Client) ListAllJobIDs(jobName string, begin, end time.Time) ([]int64, error) {
	results, err := c.issueQuery(fmt.Sprintf("count by (job_id) (cicd_jUnitResult{job=\"%s\"})", escapeQuotes(jobName)), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing job IDs: %v", err)
	}

	jobIDs := []int64{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			jobID, err := strconv.ParseInt(extractMetricFromSample(sample, "job_id"), 0, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing job id: %v", err)
			}

			jobIDs = append(jobIDs, jobID)
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	sort.SliceStable(jobIDs, func(i, j int) bool { return jobIDs[i] < jobIDs[j] })

	return jobIDs, nil
}

// ListAllCloudProviders will list all of the individual cloud providers in the given range.
func (c *Client) ListAllCloudProviders(begin, end time.Time) ([]string, error) {
	results, err := c.issueQuery("count by (cloud_provider) (cicd_jUnitResult)", begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing cloud providers: %v", err)
	}

	cloudProviders := []string{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			cloudProviders = append(cloudProviders, extractMetricFromSample(sample, "cloud_provider"))
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	return sort.StringSlice(cloudProviders), nil
}

// ListAllEnvironments will list all of the environments for a cloud provider in the given range.
func (c *Client) ListAllEnvironments(cloudProvider string, begin, end time.Time) ([]string, error) {
	results, err := c.issueQuery(fmt.Sprintf("count by (environment) (cicd_jUnitResult{cloud_provider=\"%s\"})", escapeQuotes(cloudProvider)), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing environments: %v", err)
	}

	environments := []string{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			environments = append(environments, extractMetricFromSample(sample, "environment"))
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	return sort.StringSlice(environments), nil
}

// ListAllClusterIDs will list all of the individual cluster IDs for an provider and environment in the given range.
func (c *Client) ListAllClusterIDs(cloudProvider, environment string, begin, end time.Time) ([]string, error) {
	results, err := c.issueQuery(fmt.Sprintf("count by (cluster_id) (cicd_jUnitResult{cloud_provider=\"%s\", environment=\"%s\"})",
		escapeQuotes(cloudProvider), escapeQuotes(environment)), begin, end)
	if err != nil {
		return nil, fmt.Errorf("error listing cluster IDs: %v", err)
	}

	clusterIDs := []string{}

	if matrixResults, ok := results.(model.Matrix); ok {
		for _, sample := range matrixResults {
			clusterIDs = append(clusterIDs, extractMetricFromSample(sample, "cluster_id"))
		}
	} else {
		return nil, fmt.Errorf("unrecognized result type: %v", reflect.TypeOf(results))
	}

	return sort.StringSlice(clusterIDs), nil
}

// Issues a query and prints out the associated warnings.
func (c *Client) issueQuery(query string, begin, end time.Time) (model.Value, error) {
	promAPI := v1.NewAPI(c.client)
	context, cancel := context.WithTimeout(context.Background(), viper.GetDuration(maxQueryTimeoutInSeconds)*time.Second)
	defer cancel()

	results, warnings, err := promAPI.QueryRange(context, query, makeRange(begin, end))

	if len(warnings) > 0 {
		log.Printf("Job query warnings: %v", warnings)
	}

	return results, err
}

// Just some syntactic sugar to extract a sample metric.
func extractMetricFromSample(sample *model.SampleStream, metricName model.LabelName) string {
	return string(sample.Metric[metricName])
}

func extractInstallAndUpgradeVersionsFromSample(sample *model.SampleStream) (*semver.Version, *semver.Version, error) {
	var err error
	var installVersion *semver.Version

	installVersionString := extractMetricFromSample(sample, "install_version")

	if installVersionString == "" {
		return nil, nil, fmt.Errorf("unable to find install version")
	}

	installVersion, err = util.OpenshiftVersionToSemver(installVersionString)

	if err != nil {
		return nil, nil, fmt.Errorf("error parsing install version: %v", err)
	}

	var upgradeVersion *semver.Version
	upgradeVersionString := extractMetricFromSample(sample, "upgrade_version")

	if upgradeVersionString != "" {
		upgradeVersion, err = util.OpenshiftVersionToSemver(upgradeVersionString)

		if err != nil {
			return nil, nil, fmt.Errorf("error parsing upgrade version: %v", err)
		}
	}

	return installVersion, upgradeVersion, nil
}

// Converts a string into a test result.
func stringToResult(inputString string) Result {
	lowerCaseInput := strings.ToLower(inputString)

	switch lowerCaseInput {
	case "passed":
		return Passed
	case "failed":
		return Failed
	case "skipped":
		return Skipped
	}

	return UnknownResult
}

// Converts a string into a test phase.
func stringToPhase(inputString string) Phase {
	lowerCaseInput := strings.ToLower(inputString)

	switch lowerCaseInput {
	case "install":
		return Install
	case "upgrade":
		return Upgrade
	}

	return UnknownPhase
}

func escapeQuotes(stringToEscape string) string {
	return strings.Replace(stringToEscape, `"`, `\"`, -1)
}

// makeRange will make a query range for metrics queries and bake in the 4 hour step, as it's the lowest granularity we have for any of our jobs.
func makeRange(begin, end time.Time) v1.Range {
	return v1.Range{
		Start: begin,
		End:   end,
		Step:  viper.GetDuration(stepDurationInHours) * time.Hour,
	}
}

// averageValues will just average a bunch of sample pair values together. Our granularity is extremely low, but it's lowest is 4 hours.
// We have some jobs that execute every 8 hours or 12 hours. If we have multiple values with the sample set of labels, it's likey this is
// an oversampled metric from prometheus. To get around this, we'll average the values together. We could pick the first value we see,
// but I figure this way if we do introduce metrics that produce the same set of labels and multiple metrics, we'll at least get
// an average of those values in the 4 hour step.
func averageValues(values []model.SamplePair) float64 {
	var valueSum float64 = 0
	for _, value := range values {
		valueSum = valueSum + float64(value.Value)
	}

	if valueSum == 0 {
		return 0
	}

	return valueSum / float64(len(values))
}

// pickFirstTimestamp will pick the earliest timestamp from a list of sample pairs. This is because, if we have multiple values for a singular
// sample, this is likely because the value is oversampled. This will make sure we get the earliest recorded instance of the sample. If the
// input array is empty, then 0 is returned.
func pickFirstTimestamp(values []model.SamplePair) int64 {
	if len(values) == 0 {
		return 0
	}

	var timestamp int64 = math.MaxInt64
	for _, value := range values {
		int64Timestamp := (int64)(value.Timestamp)
		if int64Timestamp < timestamp {
			timestamp = int64Timestamp
		}
	}
	return timestamp
}
