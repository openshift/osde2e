package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	"github.com/openshift/osde2e-common/pkg/clients/prometheus"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/providers"
	"k8s.io/apimachinery/pkg/util/wait"
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

var _ = ginkgo.Describe(clusterStateTestName, ginkgo.Ordered, label.E2E, func() {
	var (
		oc   *openshift.Client
		prom *prometheus.Client
	)

	ginkgo.BeforeAll(func(ctx context.Context) {
		var err error
		oc, err = openshift.New(ginkgo.GinkgoLogr)
		Expect(err).NotTo(HaveOccurred(), "unable to create openshift client")

		prom, err = prometheus.New(ctx, oc)
		Expect(err).NotTo(HaveOccurred(), "unable to create prometheus client")
	})

	ginkgo.It("should have no alerts", func(ctx context.Context) {
		var queryresult []byte

		// Query for alerts with a retry count of 40 and timeout of 20 minutes
		err := wait.PollImmediate(30*time.Second, 20*time.Minute, func() (bool, error) {
			query := "ALERTS{alertstate!=\"pending\",alertname!=\"Watchdog\"}"
			context, cancel := context.WithTimeout(ctx, 1*time.Minute)
			defer cancel()
			value, err := prom.InstantQuery(context, query)
			if err != nil {
				ginkgo.GinkgoLogr.Error(err, "Unable to query prom API")
				// try again
				return false, nil
			}
			queryresult, err = json.MarshalIndent(value, "", "  ")
			if err != nil {
				ginkgo.GinkgoLogr.Error(err, "Error marshaling results")
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
	})
})

func findCriticalAlerts(results []result, provider, environment string) bool {
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
			ginkgo.GinkgoLogr.Info(fmt.Sprintf("Active alert: %s, Severity: %s (known to be consistently critical, ignoring)", result.Metric.AlertName, result.Metric.Severity))
		} else {
			ginkgo.GinkgoLogr.Info(fmt.Sprintf("Active alert: %s, Severity: %s", result.Metric.AlertName, result.Metric.Severity))
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
