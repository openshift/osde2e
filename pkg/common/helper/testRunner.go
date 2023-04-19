package helper

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunTests will attempt to run the configured tests for the current job.
// It allows you to specify a job name prefix and arguments to a test harness container.
// It returns the names of test harnesses that failed (empty slice if none failed).
func (h *H) RunTests(ctx context.Context, name string, timeout int, harnesses, args []string) (failed []string) {
	testTemplate, err := templates.LoadTemplate("tests/tests-runner.template")
	if err != nil {
		panic(fmt.Sprintf("error while loading test runner: %v", err))
	}

	// We don't know what a test harness may need so let's give them everything.
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
			EnvironmentVariables []struct {
				Name  string
				Value string
			}
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
			EnvironmentVariables: []struct {
				Name  string
				Value string
			}{
				{
					Name:  "OCM_CLUSTER_ID",
					Value: viper.GetString(config.Cluster.ID),
				},
			},
		}

		if len(args) > 0 {
			values.Arguments = fmt.Sprintf("[\"%s\"]", strings.Join(args, "\", \""))
		}
		testCommand, err := h.ConvertTemplateToString(testTemplate, values)
		Expect(err).NotTo(HaveOccurred())

		r.Name = "test-harness"
		r.Cmd = testCommand

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
