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
	Versions    []string
	Environment string
	Failures    map[string]int
}

type summaryReportData struct {
	summedPassRates float64
	numTests        int64
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

	provider := viper.GetString(config.Weather.Provider)

	results, err := client.ListAllJUnitResults(start, end)
	if err != nil {
		return WeatherReport{}, fmt.Errorf("error during query: %v", err)
	}

	// Generate report from query results.
	jobReportData, err := generateVersionsAndFailures(results)

	if err != nil {
		return WeatherReport{}, err
	}

	summary := map[string]*summaryReportData{}

	weatherReport := WeatherReport{
		ReportDate: time.Now().UTC(),
		Provider:   provider,
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
			jobIDsAndPassRates, err := client.ListPassRatesByJobID(job, start, end)

			if err != nil {
				return WeatherReport{}, err
			}

			jobIDsAndPassRatesReport := []JobIDReport{}
			passRate := 0.0
			for jobID, jobPassRate := range jobIDsAndPassRates {
				passRate += jobPassRate

				jobIDResults, err := client.ListJUnitResultsByJobNameAndJobID(job, jobID, start, end)

				if err != nil {
					return WeatherReport{}, err
				}

				failingResults := []string{}

				installVersion := ""
				upgradeVersion := ""

				for _, jobIDResult := range jobIDResults {
					if installVersion == "" && jobIDResult.InstallVersion != nil {
						installVersion = jobIDResult.InstallVersion.String()
					}

					if upgradeVersion == "" && jobIDResult.UpgradeVersion != nil {
						upgradeVersion = jobIDResult.UpgradeVersion.String()
					}

					if jobIDResult.Result == metrics.Failed {
						failingResults = append(failingResults, jobIDResult.TestName)
					}
				}

				jobIDsAndPassRatesReport = append(jobIDsAndPassRatesReport, JobIDReport{
					JobID:          jobID,
					PassRate:       jobPassRate * 100,
					JobColor:       getPassRateColor(jobPassRate),
					FailingTests:   failingResults,
					InstallVersion: installVersion,
					UpgradeVersion: upgradeVersion,
				})
			}
			passRate = passRate / float64(len(jobIDsAndPassRates))

			environment := reportData.Environment

			// For some reason the environment label doesn't seem to be set for AWS prod jobs
			if environment == "" {
				environment = "prod"
			}

			if _, ok := summary[environment]; !ok {
				summary[environment] = &summaryReportData{}
			}

			summary[environment].summedPassRates += passRate
			summary[environment].numTests++

			weatherReport.Jobs = append(weatherReport.Jobs, JobReport{
				Name:         job,
				Viable:       len(reportData.Failures) == 0,
				Color:        getPassRateColor(passRate),
				JobIDsReport: jobIDsAndPassRatesReport,
				Versions:     reportData.Versions,
				PassRate:     passRate * 100,
				FailingTests: arrayFromMapKeys(reportData.Failures),
			})
		}
	}

	weatherReport.Summary = generateSummaryTable(summary)
	sort.Stable(weatherReport)

	return weatherReport, nil
}

func getPassRateColor(passRate float64) string {
	colorPercentage := passRate - .9
	if colorPercentage < 0 {
		colorPercentage = 0
	} else {
		colorPercentage = colorPercentage / 0.1
	}

	green := int(255 * colorPercentage)
	red := 255 - green

	return fmt.Sprintf("#%02x%02x00", red, green)
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

			jobReportData[job].Environment = result.Environment
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

// Generates a raw HTML summary
func generateSummaryTable(summary map[string]*summaryReportData) string {

	// Sort the keys
	environments := []string{}

	for environment := range summary {
		environments = append(environments, environment)
	}

	sort.Sort(sort.StringSlice(environments))

	summaryTable := strings.Builder{}

	summaryTable.WriteString("<table class=\\\"summary\\\">")

	for _, environment := range environments {
		summaryReportData := summary[environment]
		environmentPassRate := summaryReportData.summedPassRates / float64(summaryReportData.numTests)
		tableColor := getPassRateColor(environmentPassRate)

		summaryTable.WriteString(fmt.Sprintf("<tr><td bgcolor=\\\"%s\\\"></td><td>%s (Pass rate: %.2f)</td></tr>", tableColor, environment, environmentPassRate*100))
	}

	summaryTable.WriteString("</table>")

	return summaryTable.String()
}
