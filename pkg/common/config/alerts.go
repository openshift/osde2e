package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/spf13/viper"
)

// MetricAlerts is an array of LogMetric types with an easier lookup method
type MetricAlerts []MetricAlert

// once is already declared in log_metrics

var metricAlerts = MetricAlerts{}

// GetMetricAlerts will return the log metrics.
func GetMetricAlerts() MetricAlerts {
	once.Do(func() {
		viper.Set("metricAlerts", metricAlerts)
	})

	tmp := viper.Get("metricAlerts")
	ma, ok := tmp.(MetricAlerts)
	if !ok {
		log.Println("Error casting metricAlerts from Viper")
	}

	return ma
}

// AddAlert adds an alert to an existing MetricAlerts object
func (ma MetricAlerts) AddAlert(alert MetricAlert) MetricAlerts {
	ma = append(ma, alert)
	viper.Set("metricAlerts", ma)
	return ma
}

// MetricAlert lets you define a test name and the criteria to alert
// an owner via an alert channel of some sort.
type MetricAlert struct {
	// --- Description of Test ---
	// Name of the metric to look for
	Name string

	// -- Description of Test Owner ---
	// TeamOwner describes which RedHat team may own this test
	TeamOwner string
	// PrimaryContact is a point person or SME for this set of tests.
	// If there isn't one, it should default to the person committing these tests.
	PrimaryContact string

	// --- Description of Alert Channels ---
	// SlackChannel is the channel in slack to message with an alert
	SlackChannel string
	// SlackUser is the user to @ in a slack channel with an alert
	SlackUser string
	// Email is the email address to send alerts to.
	// TODO: Make this work.
	// This does not work yet.
	Email string

	// --- Description of Alert Triggers ---
	// FailureThreshold is the number of failures in a rolling window
	FailureThreshold int
}

// Notify prepares and then iterates through MetricAlerts to generate notifications
func (mas MetricAlerts) Notify() error {
	client, err := api.NewClient(api.Config{
		Address:      viper.GetString(Prometheus.Address),
		RoundTripper: WeatherRoundTripper,
	})

	if err != nil {
		return fmt.Errorf("unable to create Prometheus client: %v", err)
	}

	promAPI := v1.NewAPI(client)
	for _, ma := range mas {
		log.Printf("%v", ma)
		if err := ma.Check(promAPI); err != nil {
			return err
		}
	}

	return nil
}

// Check will query and notify depending on query results
func (ma MetricAlert) Check(prom v1.API) error {
	context, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	query := fmt.Sprintf("cicd_jUnitResult{result=\"failed\", testname=~\".*%s.*\"}[1d:4h]", ma.QuerySafeName())

	log.Printf("Query: %s", query)

	value, warnings, err := prom.Query(context, query, time.Now())
	if err != nil {
		return fmt.Errorf("error issuing query: %v", err)
	}
	for _, warning := range warnings {
		log.Printf("warning: %s", warning)
	}

	vector, _ := value.(model.Vector)

	log.Printf("%v Failures found", len(vector))

	if len(vector) >= ma.FailureThreshold {
		log.Printf("Alert triggered.")
	}

	return nil
}

// QuerySafeName is a helper function that returns a regex prometheus safe query string
func (ma MetricAlert) QuerySafeName() string {
	tmp := strings.Replace(ma.Name, "[", "\\\\[", -1)
	tmp = strings.Replace(tmp, "]", "\\\\]", -1)
	tmp = strings.Replace(tmp, "(", "\\\\(", -1)
	tmp = strings.Replace(tmp, ")", "\\\\)", -1)
	tmp = strings.Replace(tmp, "-", "\\\\-", -1)
	tmp = strings.Replace(tmp, ".", "\\\\.", -1)
	tmp = strings.Replace(tmp, ":", "\\\\:", -1)
	return tmp

}

// WeatherRoundTripper is like api.DefaultRoundTripper with an added stripping of cert verification
// and adding the bearer token to the HTTP request
var WeatherRoundTripper http.RoundTripper = &http.Transport{
	Proxy: func(request *http.Request) (*url.URL, error) {
		request.Header.Add("Authorization", "Bearer "+viper.GetString(Prometheus.BearerToken))
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
