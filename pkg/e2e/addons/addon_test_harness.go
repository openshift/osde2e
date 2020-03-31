package addons

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/markbates/pkger"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
)

var addonTestTemplate *template.Template

func init() {
	var (
		fileReader http.File
		data       []byte
		err        error
	)

	if fileReader, err = pkger.Open("/assets/addons/addon-runner.template"); err != nil {
		panic(fmt.Sprintf("unable to open addon runner template: %v", err))
	}

	if data, err = ioutil.ReadAll(fileReader); err != nil {
		panic(fmt.Sprintf("unable to read addon runner template: %v", err))
	}

	addonTestTemplate = template.Must(template.New("addon-test-runner").Parse(string(data)))

}

var _ = ginkgo.Describe("[Suite: addons] Addon Test Harness", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	addonTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// We don't know what a test harness may need so let's give them everything.
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		for _, harness := range config.Instance.Addons.TestHarnesses {
			// configure tests
			// setup runner
			r := h.RunnerWithNoCommand()

			latestImageStream, err := r.GetLatestImageStreamTag()
			Expect(err).NotTo(HaveOccurred())
			addonTestCommand, err := h.ConvertTemplateToString(addonTestTemplate, struct {
				Timeout              int
				Image                string
				OutputDir            string
				ServiceAccount       string
				PushResultsContainer string
			}{
				Timeout:              addonTimeoutInSeconds,
				Image:                harness,
				OutputDir:            runner.DefaultRunner.OutputDir,
				ServiceAccount:       h.GetNamespacedServiceAccount(),
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
		}
	}, float64(addonTimeoutInSeconds+30))
})
