package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/markbates/pkger"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/log"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
)

var (
	// cmd to run get alerts from alertmanager
	alertsCmdTpl *template.Template
)

func init() {
	var (
		fileReader http.File
		data       []byte
		err        error
	)

	if fileReader, err = pkger.Open("/assets/state/alerts.template"); err != nil {
		panic(fmt.Sprintf("unable to open alerts template: %v", err))
	}

	if data, err = ioutil.ReadAll(fileReader); err != nil {
		panic(fmt.Sprintf("unable to read alerts template: %v", err))
	}

	alertsCmdTpl = template.Must(template.New("alerts-cmd").Parse(string(data)))
}

var _ = ginkgo.Describe("[Suite: e2e] Cluster state", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()
	h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")

	alertsTimeoutInSeconds := 900
	ginkgo.It("should have no alerts", func() {
		// setup runner
		r := h.RunnerWithNoCommand()

		alertsCommand, err := h.ConvertTemplateToString(alertsCmdTpl, struct {
			OutputDir string
		}{
			OutputDir: runner.DefaultRunner.OutputDir,
		})
		Expect(err).NotTo(HaveOccurred(), "failure creating templated command")

		r.Name = "alerts"
		r.Cmd = alertsCommand

		// run tests
		stopCh := make(chan struct{})
		err = r.Run(alertsTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred(), "failure running command on pod")

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred(), "failure retrieving results from pod")

		// write results
		h.WriteResults(results)

		queryJSON := query{}
		err = json.Unmarshal(results["alerts.json"], &queryJSON)
		Expect(err).NotTo(HaveOccurred(), "failure parsing JSON results from alert manager")

		foundCritical := false
		for _, result := range queryJSON.Data.Results {
			if result.Metric.Severity == "critical" {
				foundCritical = true
			}

			log.Infof("Active alert: %s, Severity: %s", result.Metric.AlertName, result.Metric.Severity)
		}
		Expect(foundCritical).To(BeFalse(), "found a critical alert")

	}, float64(alertsTimeoutInSeconds+30))
})

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
