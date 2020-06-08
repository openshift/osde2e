package config

import (
	"log"

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
	// FailureThreshold is the number of failures in a rolling 24h window
	FailureThreshold int
	// SlowThreshold is the average time (in seconds) a test takes in a rolling 24h window
	SlowThreshold float64
}

func (ma MetricAlerts) Run() error {
	return nil
}
