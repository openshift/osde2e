package harness_runner

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joshdk/go-junit"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	projectv1 "github.com/openshift/api/project/v1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/openshift/osde2e/pkg/common/util"
)

var (
	serviceAccountDir            = "/var/run/secrets/kubernetes.io/serviceaccount"
	serviceAccount               = "system:serviceaccount:%s:cluster-admin"
	serviceAccountNamespacedName = "cluster-admin"
	timeoutInSeconds             int
	h                            *helper.H
	HarnessEntries               []ginkgo.TableEntry
	r                            *runner.Runner
	suffix                       string
	subProject                   *projectv1.Project
	err                          error
)

var _ = ginkgo.Describe("Test harness", ginkgo.Ordered, ginkgo.ContinueOnFailure, label.TestHarness, func() {
	harnesses := viper.GetStringSlice(config.Tests.TestHarnesses)
	if viper.IsSet(config.Tests.HarnessTimeout) {
		timeoutInSeconds = viper.GetInt(config.Tests.HarnessTimeout)
	} else {
		timeoutInSeconds = viper.GetInt(config.Tests.PollingTimeout)
	}
	fmt.Println("Harnesses to run: ", harnesses)
	for _, harness := range harnesses {
		HarnessEntries = append(HarnessEntries, ginkgo.Entry(harness+" should pass", harness))
	}

	ginkgo.BeforeAll(func(ctx context.Context) {
		h = helper.New()
	})
	ginkgo.BeforeEach(func(ctx context.Context) {
		ginkgo.By("Setting up new namespace")
		suffix = "h-" + util.RandomStr(5)
		subProject, err = h.SetupNewProject(ctx, suffix)
		Expect(err).NotTo(HaveOccurred(), "Could not set up harness namespace")
		err = h.SetPassthruSecretInProject(ctx, subProject)
		Expect(err).NotTo(HaveOccurred(), "Could not set up passthru secrets")
	})

	ginkgo.DescribeTable("execution",
		func(ctx context.Context, harness string) {
			log.Printf("======= RUNNING HARNESS: %s =======", harness)
			harnessImageIndex := strings.LastIndex(harness, "/")
			harnessImage := strings.Split(harness[harnessImageIndex+1:], ":")[0]
			jobName := fmt.Sprintf("%s-%s", harnessImage, suffix)

			// Create templated runner pod
			ginkgo.By("Creating test runner pod")
			h.SetServiceAccount(ctx, serviceAccount)
			r = h.RunnerWithNoCommand()
			h.SetRunnerProject(subProject.Name, r)
			latestImageStream, err := r.GetLatestImageStreamTag()
			Expect(err).NotTo(HaveOccurred(), "Could not get latest imagestream tag")
			cmd := getCommandString(timeoutInSeconds, latestImageStream, harness, suffix, jobName, serviceAccountDir)
			r = h.SetRunnerCommand(cmd, r)
			// TODO: Refactor the logic to determine whether the pod has finished or not
			//	Would be nice to see the test suite handle exiting and osde2e can pick up
			//	status of pod to decide pass/fail. Would then remove need to set individual
			//	timeouts and just have one large suite timeout for ginkgo which osde2e defines
			//	today
			ginkgo.By("Running harness pod")
			stopCh := make(chan struct{})
			err = r.Run(timeoutInSeconds, stopCh)
			Expect(err).NotTo(HaveOccurred(), "Could not run pod")

			// Retrieve and write results
			ginkgo.By("Retrieving results from test pod")
			results, err := r.RetrieveResults()
			Expect(err).NotTo(HaveOccurred(), "Could not read results")
			ginkgo.By("Writing results")
			h.WriteResults(results)
			if config.Tests.LogBucket != "" {
				err = h.UploadResultsToS3(results, harnessImage+time.Now().Format(time.DateOnly+"_"+time.TimeOnly))
				if err != nil {
					ginkgo.GinkgoLogr.Error(err, fmt.Sprintf("reporting error"))
				}
			}
			// Adding harness report failures to top level junit report
			for _, data := range results {
				suites, _ := junit.Ingest(data)
				for _, suite := range suites {
					for _, testcase := range suite.Tests {
						Expect(testcase.Error).To(BeNil(), "Assertion failed: "+testcase.Name)
					}
				}
			}
		},
		HarnessEntries)

	ginkgo.AfterEach(func(ctx context.Context) {
		ginkgo.By("Deleting harness namespace")
		err := h.DeleteProject(ctx, subProject.Name)
		if err != nil {
			ginkgo.GinkgoLogr.Error(err, fmt.Sprintf("error deleting project %q", subProject.Name))
		}
	})
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
		EnvironmentVariablesFromSecret []struct {
			SecretName string
			SecretKey  string
		}
	}{
		Name:                 jobName,
		JobName:              jobName,
		Timeout:              timeout,
		Image:                harness,
		OutputDir:            runner.DefaultRunner.OutputDir,
		ServiceAccount:       serviceAccountNamespacedName,
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
		EnvironmentVariablesFromSecret: []struct {
			SecretName string
			SecretKey  string
		}{
			{
				SecretName: "ci-secrets",
				SecretKey:  "OCM_TOKEN",
			},
		},
	}
	testTemplate, err := templates.LoadTemplate("tests/tests-runner.template")
	Expect(err).NotTo(HaveOccurred(), "Could not load pod template")
	cmd, err := h.ConvertTemplateToString(testTemplate, values)
	Expect(err).NotTo(HaveOccurred(), "Could not convert pod template")
	return cmd
}
