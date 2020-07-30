package report

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/metrics"
	"github.com/spf13/viper"
)

type reportData struct {
	Versions []string
	Failures map[string]int
}

// GenerateReport generates a weather report.
func GenerateReport() (WeatherReport, error) {
	// Range for the queries issued to Prometheus
	end := time.Now()
	start := end.Add(-time.Hour * (viper.GetDuration(config.Weather.StartOfTimeWindowInHours)))

	client, err := metrics.NewClient()

	if err != nil {
		return WeatherReport{}, fmt.Errorf("error while creating client: %v", err)
	}

	// Assemble the allowlist regexes. We'll only produce a report based on these regexes.
	allowlistRegexes := []*regexp.Regexp{}
	jobAllowlistString := viper.GetString(config.Weather.JobAllowlist)
	for _, allowlistRegex := range strings.Split(jobAllowlistString, ",") {
		allowlistRegexes = append(allowlistRegexes, regexp.MustCompile(allowlistRegex))
	}

	results, err := client.ListAllJUnitResults(start, end)
	if err != nil {
		return WeatherReport{}, fmt.Errorf("error during query: %v", err)
	}

	// Generate report from query results.
	jobReportData, err := generateVersionsAndFailures(results)

	if err != nil {
		return WeatherReport{}, err
	}

	weatherReport := WeatherReport{
		ReportDate: time.Now().UTC(),
	}
	for job, reportData := range jobReportData {
		allowed := false
		// If a job matches the allowlist, include it in the weather report.
		for _, allowlistRegex := range allowlistRegexes {
			if allowlistRegex.MatchString(job) {
				allowed = true
				break
			}
		}

		if allowed {
			passRate, err := client.GetPassRateForJob(job, start, end)

			if err != nil {
				return WeatherReport{}, err
			}
			weatherReport.Jobs = append(weatherReport.Jobs, JobReport{
				Name:         job,
				Viable:       len(reportData.Failures) == 0,
				Versions:     reportData.Versions,
				PassRate:     passRate,
				FailingTests: arrayFromMapKeys(reportData.Failures),
			})
		}
	}

	sort.Stable(weatherReport)

	return weatherReport, nil
}

// generateVersionsAndFailures generates an intermediary data structure from the results that can be used to populate
// the weather report.
func generateVersionsAndFailures(results []metrics.JUnitResult) (map[string]*reportData, error) {
	jobReportData := map[string]*reportData{}
	for _, result := range results {
		job := result.JobName

		// If there's no corresponding report data for a given job, make an empty struct.
		if _, ok := jobReportData[job]; !ok {
			jobReportData[job] = &reportData{
				Versions: []string{},
				Failures: map[string]int{},
			}
		}

		jobReportData[job].addVersion(result.InstallVersion.String())
		key := result.TestName

		if result.Result == metrics.Failed {
			// Initialize the failure count for the key if it doesn't exist
			if _, ok := jobReportData[job].Failures[key]; !ok {
				jobReportData[job].Failures[key] = 0
			}

			jobReportData[job].Failures[key] = jobReportData[job].Failures[key] + 1
		}
	}

	// Filter the failure results so that only results that cross the threshold are included.
	for _, r := range jobReportData {
		r.filterFailureResults()
	}

	return jobReportData, nil
}

// addVersion adds versions to the reportData, eliminating duplicates.
func (r *reportData) addVersion(versionToAdd string) {
	for _, version := range r.Versions {
		if version == versionToAdd {
			return
		}
	}

	r.Versions = append(r.Versions, versionToAdd)
}

// filterFailureResults eliminates results from the report that don't match the failure criteria.
// At the moment, this is pretty simple: just if tests fail more than once over the timeframe.
func (r *reportData) filterFailureResults() {
	filteredFailures := map[string]int{}
	for testname, failureCount := range r.Failures {
		if failureCount >= (viper.GetInt(config.Weather.NumberOfSamplesNecessary) - 1) {
			filteredFailures[testname] = failureCount
		}
	}

	r.Failures = filteredFailures
}

func arrayFromMapKeys(mapToExtractFrom map[string]int) []string {
	keys := []string{}
	for key := range mapToExtractFrom {
		keys = append(keys, key)
	}

	return keys
}
