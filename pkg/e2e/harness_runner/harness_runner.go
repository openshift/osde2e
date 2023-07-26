package harness_runner

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/openshift/osde2e/pkg/common/util"
)

var (
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"
	serviceAccount    = "system:serviceaccount:%s:cluster-admin"
	timeoutInSeconds  int
	h                 *helper.H
	HarnessEntries    []ginkgo.TableEntry
	r                 *runner.Runner
	suffix            string
)

var _ = ginkgo.Describe("Test harness", ginkgo.Ordered, ginkgo.ContinueOnFailure, label.TestHarness, func() {
	harnesses := viper.GetStringSlice(config.Tests.TestHarnesses)
	if viper.IsSet(config.Tests.SuiteTimeout) {
		timeoutInSeconds = viper.GetInt(config.Tests.SuiteTimeout)
	} else {
		timeoutInSeconds = viper.GetInt(config.Tests.PollingTimeout)
	}
	fmt.Println("Harnesses to run: ", harnesses)
	for _, harness := range harnesses {
		HarnessEntries = append(HarnessEntries, ginkgo.Entry("should run "+harness+" successfully", harness))
	}

	ginkgo.BeforeEach(func(ctx context.Context) {
		ginkgo.By("Setting up new namespace")
		viper.Set(config.Project, "")
		h = helper.New()
		h.SetServiceAccount(ctx, serviceAccount)
		suffix = util.RandomStr(5)
	})

	ginkgo.AfterEach(func(ctx context.Context) {
		ginkgo.By("Deleting test namespace")
		h.Cleanup(ctx)
	})

	ginkgo.DescribeTable("execution",
		func(ctx context.Context, harness string) {
			log.Printf("======= RUNNING HARNESS: %s =======", harness)
			harnessImageIndex := strings.LastIndex(harness, "/")
			harnessImage := harness[harnessImageIndex+1:]
			jobName := fmt.Sprintf("%s-%s", harnessImage, suffix)

			// Create templated runner pod
			ginkgo.By("Creating test runner pod")
			r = h.RunnerWithNoCommand()
			latestImageStream, err := r.GetLatestImageStreamTag()
			Expect(err).NotTo(HaveOccurred(), "Could not get latest imagestream tag")
			r = h.Runner(getCommandString(timeoutInSeconds, latestImageStream, harness, suffix, jobName, serviceAccountDir))

			// run tests
			ginkgo.By("Running harness pod")
			stopCh := make(chan struct{})
			err = r.Run(timeoutInSeconds, stopCh)
			Expect(err).NotTo(HaveOccurred(), "Could not run pod")

			// Retrieve and write results
			ginkgo.By("Retrieving results from test pod")
			results, err := r.RetrieveTestResults()
			Expect(err).NotTo(HaveOccurred(), "Could not read results")
			ginkgo.By("Writing results")
			h.WriteResults(results)
		},
		HarnessEntries)
})

// Generates templated command string to provide to test harness container
func getCommandString(timeout int, latestImageStream string, harness string, suffix string, jobName string, serviceAccountDir string) string {
	ginkgo.GinkgoHelper()
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
	testTemplate, err := templates.LoadTemplate("tests/tests-runner.template")
	Expect(err).NotTo(HaveOccurred(), "Could not load pod template")
	cmd, err := h.ConvertTemplateToString(testTemplate, values)
	Expect(err).NotTo(HaveOccurred(), "Could not convert pod template")
	return cmd
}
