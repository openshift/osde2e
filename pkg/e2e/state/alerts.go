package state

import (
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"log"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
)

var (
	// cmd to run get alerts from alertmanager
	alertsCmdTpl *template.Template
)

// A mapping of alerts to ignore by cluster provider and environment.
var ignoreAlerts = map[string]map[string][]string{
	"ocm": {
		"int": {"MetricsClientSendFailingSRE"},
	},
}

func init() {
	var err error

	alertsCmdTpl, err = templates.LoadTemplate("/assets/state/alerts.template")

	if err != nil {
		panic(fmt.Sprintf("error while loading alerts command: %v", err))
	}
}

var clusterStateTestName string = "[Suite: e2e] Cluster state"

func init() {
	alert.RegisterGinkgoAlert(clusterStateTestName, "SD-CICD", "Michael Wilson", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(clusterStateTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	alertsTimeoutInSeconds := 900
	ginkgo.It("should have no alerts", func() {
		// setup runner
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		r := h.RunnerWithNoCommand()

		alertsCommand, err := h.ConvertTemplateToString(alertsCmdTpl, struct {
			OutputDir string
		}{
			OutputDir: runner.DefaultRunner.OutputDir,
		})
		Expect(err).NotTo(HaveOccurred(), "failure creating templated command")

		r.Name = "alerts"
		r.Cmd = alertsCommand

		Eventually(func() bool {
			stopCh := make(chan struct{})
			// run tests
			if err = r.Run(alertsTimeoutInSeconds, stopCh); err != nil {
				log.Printf("Error running command on pod: %s", err.Error())
				return false
			}

			// get results
			results, err := r.RetrieveResults()
			if err != nil {
				log.Printf("Error retrieving results from pod: %s", err.Error())
				return false
			}

			// write results
			h.WriteResults(results)

			queryJSON := query{}
			if err = json.Unmarshal(results["alerts.json"], &queryJSON); err != nil {
				log.Printf("Error parsing JSON results from AlertManager: %s", err.Error())
				return false
			}

			clusterProvider, err := providers.ClusterProvider()
			if err != nil {
				log.Printf("Error getting cluster provider: %s", err.Error())
			}

			return !findCriticalAlerts(queryJSON.Data.Results, viper.GetString(config.Provider), clusterProvider.Environment())
		}, 5*time.Minute, 30*time.Second).Should(BeTrue(), "never able to find zero alerts")

	}, float64(alertsTimeoutInSeconds+30))
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
			log.Printf("Active alert: %s, Severity: %s (known to be consistently critical, ignoring)", result.Metric.AlertName, result.Metric.Severity)
		} else {
			log.Printf("Active alert: %s, Severity: %s", result.Metric.AlertName, result.Metric.Severity)
		}
	}

	return foundCritical
}

type query struct {
	Data data `json:"data"`
}

type data struct {
	Results []result `json:"result"`
}

type result struct {
	Metric metric `json:"metric"`
}

type metric struct {
	AlertName string `json:"alertname"`
	Severity  string `json:"severity"`
}
