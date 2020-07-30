package helper

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var addonTestTemplate *template.Template
var err error

// RunAddonTests will attempt to run the configured addon tests for the current job
// It allows you to specify a job name prefix and arguments to a test harness container
func (h *H) RunAddonTests(name string, args []string) {
	addonTimeoutInSeconds := 3600
	addonTestTemplate, err = templates.LoadTemplate("/assets/addons/addon-runner.template")

	if err != nil {
		panic(fmt.Sprintf("error while loading addon test runner: %v", err))
	}

	// We don't know what a test harness may need so let's give them everything.
	h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
	addonTestHarnesses := strings.Split(viper.GetString(config.Addons.TestHarnesses), ",")
	for key, harness := range addonTestHarnesses {
		// configure tests
		// setup runner
		r := h.RunnerWithNoCommand()

		latestImageStream, err := r.GetLatestImageStreamTag()
		jobName := fmt.Sprintf("%s-%d", name, key)
		Expect(err).NotTo(HaveOccurred())
		values := struct {
			JobName              string
			Arguments            string
			Timeout              int
			Image                string
			OutputDir            string
			ServiceAccount       string
			PushResultsContainer string
		}{
			JobName:              jobName,
			Timeout:              addonTimeoutInSeconds,
			Image:                harness,
			OutputDir:            runner.DefaultRunner.OutputDir,
			ServiceAccount:       h.GetNamespacedServiceAccount(),
			PushResultsContainer: latestImageStream,
		}

		if len(args) > 0 {
			values.Arguments = fmt.Sprintf("[\"%s\"]", strings.Join(args, "\", \""))
		}
		addonTestCommand, err := h.ConvertTemplateToString(addonTestTemplate, values)
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
		job, err := h.Kube().BatchV1().Jobs(r.Namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(job.Status.Failed).Should(BeNumerically("==", 0))
	}
}
