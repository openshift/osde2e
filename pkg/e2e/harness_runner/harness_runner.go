package harness_runner

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/prow"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"
	TimeoutInSeconds  = viper.GetFloat64(config.Tests.PollingTimeout)
	harnesses         = strings.Split(viper.GetString(config.Tests.TestHarnesses), ",")
	h                 *helper.H
	failed            []string
	err               error
	HarnessEntries    []ginkgo.TableEntry
)

var _ = ginkgo.Describe("Test Harness", ginkgo.Ordered, label.TestHarness, func() {
	for _, harness := range harnesses {
		HarnessEntries = append(HarnessEntries, ginkgo.Entry("should run "+harness+" successfully", harness))
	}
	ginkgo.DescribeTable("Executing Harness",
		func(harness string) {
			ginkgo.By("======= RUNNING HARNESS: " + harness + " =======")
			log.Printf("======= RUNNING HARNESS: %s =======", harness)
			viper.Set(config.Project, "")
			//Run harness in new project
			h = helper.New()
			h.SetServiceAccount(context.TODO(), "system:serviceaccount:%s:cluster-admin")
			harnessImageIndex := strings.LastIndex(harness, "/")
			harnessImage := harness[harnessImageIndex+1:]
			suffix := util.RandomStr(5)
			jobName := fmt.Sprintf("%s-%s", harnessImage, suffix)
			r := h.RunnerWithNoCommand()
			testCommand, err := getCommandString(h, harness, r, suffix, jobName)
			Expect(err).NotTo(HaveOccurred(), "Could not create test pod template")
			r.Cmd = testCommand

			// run tests
			stopCh := make(chan struct{})
			err = r.Run(int(TimeoutInSeconds), stopCh)
			Expect(err).NotTo(HaveOccurred(), "Could not run pod")

			// get results
			results, err := r.RetrieveTestResults()
			Expect(err).NotTo(HaveOccurred(), "Could not read results")

			// write results
			h.WriteResults(results)

			// ensure job has not failed
			job, err := h.Kube().BatchV1().Jobs(r.Namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "Harness job pods failed")
			if !Expect(job.Status.Failed).Should(BeNumerically("==", 0)) {
				failed = append(failed, harness)
			}

			h.Cleanup(context.TODO())
			ginkgo.By("======= FINISHED HARNESS: " + harness + " =======")
		},
		HarnessEntries)

	if len(failed) > 0 {
		message := fmt.Sprintf("Tests failed: %v", failed)
		if url, ok := prow.JobURL(); ok {
			message += "\n" + url
		}
		if err := alert.SendSlackMessage(viper.GetString(config.Tests.SlackChannel), message); err != nil {
			log.Printf("Failed sending slack alert for test failure: %v", err)
		}
	}
})

// Generates templated command string to provide to test harness container
func getCommandString(h *helper.H, harness string, r *runner.Runner, suffix string, jobName string) (string, error) {
	// setup runner
	latestImageStream, err := r.GetLatestImageStreamTag()
	Expect(err).NotTo(HaveOccurred())

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
		Timeout:              int(TimeoutInSeconds),
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
	testTemplate, err := templates.LoadTemplate("tests/tests-runner.template")
	return h.ConvertTemplateToString(testTemplate, values)

}
