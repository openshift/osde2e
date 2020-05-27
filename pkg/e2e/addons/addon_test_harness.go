package addons

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var addonTestTemplate *template.Template

func init() {
	var err error

	addonTestTemplate, err = templates.LoadTemplate("/assets/addons/addon-runner.template")

	if err != nil {
		panic(fmt.Sprintf("error while loading addon test runner: %v", err))
	}
}

var _ = ginkgo.Describe("[Suite: addons] Addon Test Harness", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	addonTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// We don't know what a test harness may need so let's give them everything.
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		addonTestHarnesses := strings.Split(viper.GetString(config.Addons.TestHarnesses), ",")
		for _, harness := range addonTestHarnesses {
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

			// ensure job has not failed
			job, err := h.Kube().BatchV1().Jobs(r.Namespace).Get("addon-tests", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(job.Status.Failed).Should(BeNumerically("==", 0))
		}
	}, float64(addonTimeoutInSeconds+30))
})
