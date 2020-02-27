package gate

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

const (
	gateQuery = `count by (install_version, suite, testname, result) (cicd_jUnitResult{environment="%s", install_version=~"%s.*"})`

	stepDurationInHours = 4
)

// Report is the gating report.
type Report struct {
	Viable       bool     `json:"viable"`
	Versions     []string `json:"versions"`
	FailingTests []string `json:"failingTests,omitempty"`
}

// gateRoundTripper is like api.DefaultRoundTripper with an added stripping of cert verification
// and adding the bearer token to the HTTP request
var gateRoundTripper http.RoundTripper = &http.Transport{
	Proxy: func(request *http.Request) (*url.URL, error) {
		request.Header.Add("Authorization", "Bearer "+config.Instance.Gate.PrometheusBearerToken)
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

// GenerateReleaseReportForOSD will return true if a release of OCP is viable for OSD.
func GenerateReleaseReportForOSD(environment, openshiftVersion, output string) (bool, error) {
	// Range for the queries issued to Prometheus
	queryRange := v1.Range{
		Start: time.Now().Add(-time.Hour * config.Instance.Gate.StartOfTimeWindowInHours),
		End:   time.Now(),
		Step:  stepDurationInHours * time.Hour,
	}

	client, err := api.NewClient(api.Config{
		Address:      config.Instance.Gate.PrometheusAddress,
		RoundTripper: gateRoundTripper,
	})

	if err != nil {
		return false, fmt.Errorf("error while creating client: %v", err)
	}

	promAPI := v1.NewAPI(client)
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var writer io.Writer
	if output == "-" {
		writer = os.Stdout
	} else {
		file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

		if err != nil {
			return false, fmt.Errorf("error opening output file for writing: %v", err)
		}

		defer file.Close()

		writer = file
	}

	report, err := generateReport(context, promAPI, environment, openshiftVersion, queryRange)

	if err != nil {
		return false, fmt.Errorf("error while generating report: %v", err)
	}

	jsonReport, err := json.MarshalIndent(report, "", "  ")

	if err != nil {
		return false, fmt.Errorf("error while marshaling report into JSON: %v", err)
	}

	_, err = writer.Write(append(jsonReport, '\n'))

	if err != nil {
		return false, fmt.Errorf("error while writing report to output: %v", err)
	}

	return report.Viable, nil
}

func generateReport(context context.Context, promAPI v1.API, environment string, openshiftVersion string, queryRange v1.Range) (Report, error) {
	results, warnings, err := promAPI.QueryRange(context, fmt.Sprintf(gateQuery, environment, openshiftVersion), queryRange)
	if err != nil {
		return Report{Viable: false}, fmt.Errorf("error during query: %v", err)
	}

	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	if matrixResults, ok := results.(model.Matrix); ok {
		versions, failures := generateVersionsAndFailures(matrixResults)

		report := Report{
			Viable:       len(failures) == 0,
			Versions:     versions,
			FailingTests: failures,
		}

		return report, nil
	}

	log.Printf("Results not in the expected format.")
	return Report{Viable: false}, nil
}

func generateVersionsAndFailures(matrixResults model.Matrix) ([]string, []string) {
	versions := map[string]bool{}
	failures := map[string]int{}

	for _, sample := range matrixResults {
		versions[fmt.Sprintf("%s", sample.Metric["install_version"])] = true
		key := fmt.Sprintf("%s", sample.Metric["testname"])

		if sample.Metric["result"] == "failed" {
			// Initialize the failure count for the key if it doesn't exist
			if _, ok := failures[key]; !ok {
				failures[key] = 0
			}

			failures[key] = failures[key] + len(sample.Values)
		}
	}

	versionArray := []string{}
	for version := range versions {
		versionArray = append(versionArray, version)
	}

	failureArray := []string{}
	for testname, failureCount := range failures {
		if failureCount > (config.Instance.Gate.NumberOfSamplesNecessary - 1) {
			failureArray = append(failureArray, testname)
		}
	}

	return versionArray, failureArray
}
