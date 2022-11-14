package state

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/util"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers"

	"github.com/openshift/osde2e/pkg/common/prometheus"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// A mapping of alerts to ignore by cluster provider and environment.
var ignoreAlerts = map[string]map[string][]string{
	"ocm": {
		"int": {"MetricsClientSendFailingSRE"},
	},
}

var clusterStateTestName string = "[Suite: e2e] Cluster state"

func init() {
	alert.RegisterGinkgoAlert(clusterStateTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(clusterStateTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	alertsTimeoutInSeconds := 900
	util.GinkgoIt("should have no alerts", func(ctx context.Context) {
		// Set up prometheus client
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
		promClient, err := prometheus.CreateClusterClient(h)
		Expect(err).NotTo(HaveOccurred(), "error creating a prometheus client")
		promAPI := promv1.NewAPI(promClient)

		var queryresult []byte

		// Query for alerts with a retry count of 40 and timeout of 20 minutes
		err = wait.PollImmediate(30*time.Second, 20*time.Minute, func() (bool, error) {
			query := "ALERTS{alertstate!=\"pending\",alertname!=\"Watchdog\"}"
			context, cancel := context.WithTimeout(ctx, 1*time.Minute)
			defer cancel()
			value, _, err := promAPI.Query(context, query, time.Now())
			if err != nil {
				log.Printf("Unable to query prom API: %v", err)
				// try again
				return false, nil
			}
			queryresult, err = json.MarshalIndent(value, "", "  ")
			if err != nil {
				log.Printf("Error marshaling results: %v", err)
				// try again
				return false, nil
			}
			return true, nil
		})
		Expect(err).NotTo(HaveOccurred(), "error retrieving results from alert gatherer")

		// Store JSON query results in an object
		queryJSON := []result{}
		err = json.Unmarshal(queryresult, &queryJSON)
		Expect(err).NotTo(HaveOccurred(), "error unmarshalling json from alert gatherer")

		clusterProvider, err := providers.ClusterProvider()
		Expect(err).NotTo(HaveOccurred(), "error retrieving cluster provider")

		Expect(!findCriticalAlerts(queryJSON, viper.GetString(config.Provider), clusterProvider.Environment())).Should(BeTrue(), "never able to find zero alerts")
	}, float64(alertsTimeoutInSeconds+30))
})

func findCriticalAlerts(results []result, provider, environment string) bool {
	log.Printf("Alerts found: %v", results)
	foundCritical := false
	for _, result := range results {
		ignoredCritical := false
		if result.Metric.Severity == "critical" {
			// If there alerts to ignore for this provider, let's look through them.
			if ignoreForEnv, ok := ignoreAlerts[provider]; ok {
				// If we can find alerts to ignore for this environment, let's look through those, too.
				if ignoreAlertList, ok := ignoreForEnv[environment]; ok {
					for _, alertToIgnore := range ignoreAlertList {
						// If we find an alert in our ignore alert list, set this flag. This will indicate that the presence
						// of this alert will not fail this test.
						if alertToIgnore == result.Metric.AlertName {
							ignoredCritical = true
							break
						}
					}
				}
			}

			if !ignoredCritical {
				foundCritical = true
			}
		}

		if ignoredCritical {
			log.Printf("Active alert: %s, Severity: %s (known to be consistently critical, ignoring)", result.Metric.AlertName, result.Metric.Severity)
		} else {
			log.Printf("Active alert: %s, Severity: %s", result.Metric.AlertName, result.Metric.Severity)
		}
	}

	return foundCritical
}

type result struct {
	Metric metric `json:"metric"`
}

type metric struct {
	AlertName string `json:"alertname"`
	Severity  string `json:"severity"`
}
