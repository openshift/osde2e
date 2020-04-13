package report

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

const (
	gateQuery = `count by (job, install_version, suite, testname, result) (cicd_jUnitResult)`

	stepDurationInHours = 4
)

type reportData struct {
	Versions []string
	Failures map[string]int
}

// weatherRoundTripper is like api.DefaultRoundTripper with an added stripping of cert verification
// and adding the bearer token to the HTTP request
var weatherRoundTripper http.RoundTripper = &http.Transport{
	Proxy: func(request *http.Request) (*url.URL, error) {
		request.Header.Add("Authorization", "Bearer "+config.Instance.Weather.PrometheusBearerToken)
		return http.ProxyFromEnvironment(request)
	},
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
	TLSHandshakeTimeout: 10 * time.Second,
}

// GenerateReport generates a weather report.
func GenerateReport() (WeatherReport, error) {
	// Range for the queries issued to Prometheus
	queryRange := v1.Range{
		Start: time.Now().Add(-time.Hour * config.Instance.Weather.StartOfTimeWindowInHours),
		End:   time.Now(),
		Step:  stepDurationInHours * time.Hour,
	}

	client, err := api.NewClient(api.Config{
		Address:      config.Instance.Weather.PrometheusAddress,
		RoundTripper: weatherRoundTripper,
	})

	if err != nil {
		return WeatherReport{}, fmt.Errorf("error while creating client: %v", err)
	}

	promAPI := v1.NewAPI(client)
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Assemble the whitelist regexes. We'll only produce a report based on these regexes.
	whitelistRegexes := []*regexp.Regexp{}
	for _, whitelistRegex := range config.Instance.Weather.JobWhitelist {
		whitelistRegexes = append(whitelistRegexes, regexp.MustCompile(whitelistRegex))
	}

	results, warnings, err := promAPI.QueryRange(context, gateQuery, queryRange)
	if err != nil {
		return WeatherReport{}, fmt.Errorf("error during query: %v", err)
	}

	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	// Generate report from query results.
	if matrixResults, ok := results.(model.Matrix); ok {
		jobReportData, err := generateVersionsAndFailures(matrixResults)

		if err != nil {
			return WeatherReport{}, err
		}

		weatherReport := WeatherReport{
			ReportDate: time.Now().UTC(),
		}
		for job, reportData := range jobReportData {
			whitelisted := false
			// If a job matches the whitelist, include it in the weather report.
			for _, whitelistRegex := range whitelistRegexes {
				if whitelistRegex.MatchString(job) {
					whitelisted = true
					break
				}
			}

			if whitelisted {
				weatherReport.Jobs = append(weatherReport.Jobs, JobReport{
					Name:         job,
					Viable:       len(reportData.Failures) == 0,
					Versions:     reportData.Versions,
					FailingTests: arrayFromMapKeys(reportData.Failures),
				})
			}
		}

		sort.Stable(weatherReport)

		return weatherReport, nil
	}

	return WeatherReport{}, fmt.Errorf("results not in the expected format")
}

// generateVersionsAndFailures generates an intermediary data structure from the results that can be used to populate
// the weather report.
func generateVersionsAndFailures(matrixResults model.Matrix) (map[string]*reportData, error) {
	jobReportData := map[string]*reportData{}
	for _, sample := range matrixResults {
		job := fmt.Sprintf("%s", sample.Metric["job"])

		// If there's no corresponding report data for a given job, make an empty struct.
		if _, ok := jobReportData[job]; !ok {
			jobReportData[job] = &reportData{
				Versions: []string{},
				Failures: map[string]int{},
			}
		}

		jobReportData[job].addVersion(fmt.Sprintf("%s", sample.Metric["install_version"]))
		key := fmt.Sprintf("%s", sample.Metric["testname"])

		if sample.Metric["result"] == "failed" {
			// Initialize the failure count for the key if it doesn't exist
			if _, ok := jobReportData[job].Failures[key]; !ok {
				jobReportData[job].Failures[key] = 0
			}

			jobReportData[job].Failures[key] = jobReportData[job].Failures[key] + len(sample.Values)
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
		if failureCount >= (config.Instance.Weather.NumberOfSamplesNecessary - 1) {
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
