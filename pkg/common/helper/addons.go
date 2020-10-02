package helper

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var addonTestTemplate *template.Template
var err error

// RunAddonTests will attempt to run the configured addon tests for the current job
// It allows you to specify a job name prefix and arguments to a test harness container
func (h *H) RunAddonTests(name string, timeout int, harnesses []string, args []string) {
	addonTestTemplate, err = templates.LoadTemplate("/assets/addons/addon-runner.template")

	if err != nil {
		panic(fmt.Sprintf("error while loading addon test runner: %v", err))
	}

	// We don't know what a test harness may need so let's give them everything.
	h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
	for _, harness := range harnesses {
		// configure tests
		// setup runner
		r := h.RunnerWithNoCommand()

		suffix := util.RandomStr(5)

		latestImageStream, err := r.GetLatestImageStreamTag()
		jobName := fmt.Sprintf("%s-%s", name, suffix)
		Expect(err).NotTo(HaveOccurred())
		serviceAccountDir := "/var/run/secrets/kubernetes.io/serviceaccount"
		values := struct {
			Name                 string
			JobName              string
			Arguments            string
			Timeout              int
			Image                string
			OutputDir            string
			ServiceAccount       string
			PushResultsContainer string
			Suffix               string
			Server               string
			CA                   string
			TokenFile            string
		}{
			Name:                 jobName,
			JobName:              jobName,
			Timeout:              timeout,
			Image:                harness,
			OutputDir:            runner.DefaultRunner.OutputDir,
			ServiceAccount:       h.GetNamespacedServiceAccount(),
			PushResultsContainer: latestImageStream,
			Suffix:               suffix,
			Server:               "https://kubernetes.default",
			CA:                   serviceAccountDir + "/ca.crt",
			TokenFile:            serviceAccountDir + "/token",
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
		err = r.Run(timeout, stopCh)
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
