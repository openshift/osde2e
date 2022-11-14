package weather

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/metrics"
	"github.com/openshift/osde2e/pkg/reporting/spi"
	"github.com/openshift/osde2e/pkg/reporting/templates"
)

// Intermediary structs
type reportData struct {
	Versions    []string
	Environment string
	Failures    map[string]int
}

type summaryReportData struct {
	summedPassRates float64
	numTests        int64
}

// weatherReport is the weather report.
type weatherReport struct {
	ReportDate time.Time   `json:"reportDate"`
	Provider   string      `json:"provider"`
	Jobs       []jobReport `json:"jobs"`
	Summary    string      `json:"summary"`

	// We want the sort interface so that we can sort jobs and produce stable, comparable reports.
	sort.Interface `json:"-"`
}

// jobReport is a report for an individual job.
type jobReport struct {
	Name         string        `json:"name"`
	Viable       bool          `json:"viable"`
	Color        string        `json:"color"`
	JobIDsReport []jobIDReport `json:"jobIDsReport"`
	Versions     []string      `json:"versions"`
	PassRate     float64       `json:"passRate"`
	FailingTests []string      `json:"failingTests,omitempty"`
}

// jobIDReport combines the job ID, pass rate, and a color for the job run together.
type jobIDReport struct {
	JobID          int64    `json:"jobID"`
	PassRate       float64  `json:"passRate"`
	JobColor       string   `json:"jobColor"`
	InstallVersion string   `json:"installVersion"`
	UpgradeVersion string   `json:"upgradeVersion"`
	FailingTests   []string `json:"failingTests,omitempty"`
}

// Len is the number of jobs in the weather report.
func (w weatherReport) Len() int {
	return len(w.Jobs)
}

// Less reports whether the element with index i should sort before the element with index j.
func (w weatherReport) Less(i, j int) bool {
	return w.Jobs[i].Name < w.Jobs[j].Name
}

// Swap swaps the elements with indexes i and j.
func (w weatherReport) Swap(i, j int) {
	w.Jobs[i], w.Jobs[j] = w.Jobs[j], w.Jobs[i]
}

// ToJSON will convert the weather report into a JSON object.
func (w weatherReport) ToJSON() ([]byte, error) {
	jsonReport, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error while marshaling report into JSON: %v", err)
	}

	return append(jsonReport, '\n'), nil
}

// Reporter will write out the actual weather report.
type Reporter struct{}

func init() {
	spi.RegisterReporter(Reporter{})
}

// Name will return the name of the weather reporter.
func (w Reporter) Name() string {
	return "weather-report"
}

// GenerateReport generates a weather report.
func (w Reporter) GenerateReport(reportType string) ([]byte, error) {
	// Range for the queries issued to Prometheus
	end := time.Now()
	start := end.Add(-time.Hour * (viper.GetDuration(StartOfTimeWindowInHours)))

	client, err := metrics.NewClient()
	if err != nil {
		return nil, fmt.Errorf("error while creating client: %v", err)
	}

	// Assemble the allowlist regexes. We'll only produce a report based on these regexes.
	allowlistRegexes := []*regexp.Regexp{}
	jobAllowlistString := viper.GetString(JobAllowlist)
	for _, allowlistRegex := range strings.Split(jobAllowlistString, ",") {
		allowlistRegexes = append(allowlistRegexes, regexp.MustCompile(allowlistRegex))
	}

	provider := viper.GetString(Provider)

	results, err := client.ListAllJUnitResults(start, end)
	if err != nil {
		return nil, fmt.Errorf("error during query: %v", err)
	}

	// Generate report from query results.
	jobReportData, err := generateVersionsAndFailures(results)
	if err != nil {
		return nil, err
	}

	summary := map[string]*summaryReportData{}

	weatherReport := weatherReport{
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
				return nil, err
			}

			jobIDsAndPassRatesReport := []jobIDReport{}
			passRate := 0.0
			environment := reportData.Environment

			for jobID, jobPassRate := range jobIDsAndPassRates {
				passRate += jobPassRate

				jobIDResults, err := client.ListJUnitResultsByJobNameAndJobID(job, jobID, start, end)
				if err != nil {
					return nil, err
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

				if _, ok := summary[environment]; !ok {
					summary[environment] = &summaryReportData{}
				}

				summary[environment].summedPassRates += jobPassRate
				summary[environment].numTests++

				jobIDsAndPassRatesReport = append(jobIDsAndPassRatesReport, jobIDReport{
					JobID:          jobID,
					PassRate:       jobPassRate * 100,
					JobColor:       getPassRateColor(jobPassRate),
					FailingTests:   failingResults,
					InstallVersion: installVersion,
					UpgradeVersion: upgradeVersion,
				})
			}
			passRate = passRate / float64(len(jobIDsAndPassRates))

			weatherReport.Jobs = append(weatherReport.Jobs, jobReport{
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

	return templates.WriteReport(weatherReport, w.Name(), reportType)
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

		jobReportData[job].Environment = result.Environment
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
		if failureCount >= (viper.GetInt(NumberOfSamplesNecessary) - 1) {
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
