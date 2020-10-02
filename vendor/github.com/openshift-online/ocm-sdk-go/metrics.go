/*
Copyright (c) 2019 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file contains the implementations of the Prometheus metrics.

package sdk

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// registerMetrics registers the metrics with the Prometheus library.
func (c *Connection) registerMetrics(subsystem string) error {
	var err error

	// Register the token request count metric:
	c.tokenCountMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "token_request_count",
			Help:      "Number of token requests sent.",
		},
		tokenMetricsLabels,
	)
	err = prometheus.Register(c.tokenCountMetric)
	if err != nil {
		registered, ok := err.(prometheus.AlreadyRegisteredError)
		if ok {
			c.tokenCountMetric = registered.ExistingCollector.(*prometheus.CounterVec)
		} else {
			return err
		}
	}

	// Description of the token request duration metric:
	c.tokenDurationMetric = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "token_request_duration",
			Help:      "Token request duration in seconds.",
			Buckets: []float64{
				0.1,
				1.0,
				10.0,
				30.0,
			},
		},
		tokenMetricsLabels,
	)
	err = prometheus.Register(c.tokenDurationMetric)
	if err != nil {
		registered, ok := err.(prometheus.AlreadyRegisteredError)
		if ok {
			c.tokenDurationMetric = registered.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			return err
		}
	}

	// Register the call count metric:
	c.callCountMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "request_count",
			Help:      "Number of requests sent.",
		},
		callMetricsLabels,
	)
	err = prometheus.Register(c.callCountMetric)
	if err != nil {
		registered, ok := err.(prometheus.AlreadyRegisteredError)
		if ok {
			c.callCountMetric = registered.ExistingCollector.(*prometheus.CounterVec)
		} else {
			return err
		}
	}

	// Description of the call duration metric:
	c.callDurationMetric = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_duration",
			Help:      "Request duration in seconds.",
			Buckets: []float64{
				0.1,
				1.0,
				10.0,
				30.0,
			},
		},
		callMetricsLabels,
	)
	err = prometheus.Register(c.callDurationMetric)
	if err != nil {
		registered, ok := err.(prometheus.AlreadyRegisteredError)
		if ok {
			c.callDurationMetric = registered.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			return err
		}
	}

	return nil
}

func (c *Connection) GetAPIServiceLabelFromPath(path string) string {
	if strings.HasPrefix(path, "/api/accounts_mgmt") {
		return "ocm-accounts-service"
	} else if strings.HasPrefix(path, "/api/clusters_mgmt") {
		return "ocm-clusters-service"
	} else if strings.HasPrefix(path, "/api/authorizations") {
		return "ocm-authorizations-service"
	} else if strings.HasPrefix(path, "/api/service_logs") {
		return "ocm-logs-service"
	} else {
		pathParts := strings.Split(path, "/")
		if len(pathParts) > 3 {
			pathParts = pathParts[:3]
		}
		return "ocm-" + strings.Join(pathParts, "/")
	}
}

// Names of the labels added to metrics:
const (
	metricsAPIServiceLabel = "apiservice"
	metricsAttemptLabel    = "attempt"
	metricsCodeLabel       = "code"
	metricsMethodLabel     = "method"
	metricsPathLabel       = "path"
)

// Array of labels added to token metrics:
var tokenMetricsLabels = []string{
	metricsAttemptLabel,
	metricsCodeLabel,
}

// Array of labels added to call metrics:
var callMetricsLabels = []string{
	metricsAPIServiceLabel,
	metricsCodeLabel,
	metricsMethodLabel,
	metricsPathLabel,
}

// Name of the header that contains the metrics path:
const metricHeader = "X-Metric"
