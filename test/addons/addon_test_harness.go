package addons

import (
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/helper"
	"github.com/openshift/osde2e/pkg/runner"
)

var addonTestTemplate *template.Template

func init() {
	data, err := ioutil.ReadFile("testdata/addon-runner.template")
	if err != nil {
		panic(fmt.Sprintf("unable to read addon runner template: %v", err))
	}

	addonTestTemplate = template.Must(template.New("addon-test-runner").Parse(string(data)))

}

var _ = ginkgo.Describe("Addon Test Harness", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	for _, harness := range config.Instance.Addons.TestHarnesses {
		if harness != "" {
			addonTimeoutInSeconds := 3600
			ginkgo.It("should run until completion", func() {
				// configure tests
				// setup runner
				r := h.RunnerWithNoCommand()

				latestImageStream, err := r.GetLatestImageStreamTag()
				Expect(err).NotTo(HaveOccurred())
				addonTestCommand, err := h.ConvertTemplateToString(addonTestTemplate, struct {
					Timeout              int
					Image                string
					OutputDir            string
					PushResultsContainer string
				}{
					Timeout:              addonTimeoutInSeconds,
					Image:                harness,
					OutputDir:            runner.DefaultRunner.OutputDir,
					PushResultsContainer: latestImageStream,
				})
				Expect(err).NotTo(HaveOccurred())

				r.Name = "addon-tests"
				r.Cmd = addonTestCommand

				// run tests
				stopCh := make(chan struct{})
				err = r.Run(addonTimeoutInSeconds, stopCh)
				Expect(err).NotTo(HaveOccurred())

				// get results
				results, err := r.RetrieveResults()
				Expect(err).NotTo(HaveOccurred())

				// write results
				h.WriteResults(results)
			}, float64(addonTimeoutInSeconds+30))
		}
	}
})
