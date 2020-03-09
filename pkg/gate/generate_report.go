package gate

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/report"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

const (
	gateQuery = `count by (install_version, suite, testname, result) (cicd_jUnitResult{environment="%s", install_version=~"%s.*"})`

	stepDurationInHours = 4
)

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

// GenerateReleaseReportForOSD will generate a JSON report for a release of OpenShift on OSD.
func GenerateReleaseReportForOSD(environment, openshiftVersion, output string) error {
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
		return fmt.Errorf("error while creating client: %v", err)
	}

	promAPI := v1.NewAPI(client)
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	report, err := generateReport(context, promAPI, environment, openshiftVersion, queryRange)

	if err != nil {
		return fmt.Errorf("error while generating report: %v", err)
	}

	err = report.ToOutput(output)

	if err != nil {
		return fmt.Errorf("error while writing out report: %v", err)
	}

	return nil
}

func generateReport(context context.Context, promAPI v1.API, environment string, openshiftVersion string, queryRange v1.Range) (report.GateReport, error) {
	results, warnings, err := promAPI.QueryRange(context, fmt.Sprintf(gateQuery, environment, openshiftVersion), queryRange)
	if err != nil {
		return report.GateReport{Viable: false}, fmt.Errorf("error during query: %v", err)
	}

	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	if matrixResults, ok := results.(model.Matrix); ok {
		versions, failures := generateVersionsAndFailures(matrixResults)

		report := report.GateReport{
			Viable:       len(failures) == 0,
			Versions:     versions,
			FailingTests: failures,
		}

		return report, nil
	}

	log.Printf("Results not in the expected format.")
	return report.GateReport{Viable: false}, nil
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
