package helper

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GEneralizd harness runner: Essentially replicates addon harness runner, with generalized harness template
func (h *H) RunTestHarness(ctx context.Context, name string, timeout int, harnesses, args []string) (failed []string) {
	testHarnessTemplate, err := templates.LoadTemplate("harness/harness-runner.template")
	if err != nil {
		panic(fmt.Sprintf("error while loading test harness runner: %v", err))
	}

	h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
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
		testHarnessCommand, err := h.ConvertTemplateToString(testHarnessTemplate, values)
		Expect(err).NotTo(HaveOccurred())

		r.Name = "test-harness"
		r.Cmd = testHarnessCommand

		// run tests
		stopCh := make(chan struct{})
		err = r.Run(timeout, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveTestResults()

		// write results
		h.WriteResults(results)

		// evaluate results
		Expect(err).NotTo(HaveOccurred())

		// ensure job has not failed
		job, err := h.Kube().BatchV1().Jobs(r.Namespace).Get(ctx, jobName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		if !Expect(job.Status.Failed).Should(BeNumerically("==", 0)) {
			failed = append(failed, harness)
		}
	}
	return failed
}
