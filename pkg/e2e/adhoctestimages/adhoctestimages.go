package adhoctestimages

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
	"github.com/openshift/osde2e/pkg/common/util"
)

var (
	serviceAccountDir            = "/var/run/secrets/kubernetes.io/serviceaccount"
	serviceAccount               = "system:serviceaccount:%s:cluster-admin"
	serviceAccountNamespacedName = "cluster-admin"
	timeoutInSeconds             int
	h                            *helper.H
	AdHocTestImageEntries        []ginkgo.TableEntry
	r                            *runner.Runner
	suffix                       string
	subProject                   *projectv1.Project
	err                          error
)

var _ = ginkgo.Describe("Ad Hoc Test Images", ginkgo.Ordered, ginkgo.ContinueOnFailure, label.AdHocTestImages, func() {
	adHocTestImages := viper.GetStringSlice(config.Tests.AdHocTestImages)
	if viper.IsSet(config.Tests.AdHocTestContainerTimeout) {
		timeoutInSeconds = viper.GetInt(config.Tests.AdHocTestContainerTimeout)
	} else {
		timeoutInSeconds = viper.GetInt(config.Tests.PollingTimeout)
	}
	for _, adHocTestImage := range adHocTestImages {
		AdHocTestImageEntries = append(AdHocTestImageEntries, ginkgo.Entry(adHocTestImage+" should pass", adHocTestImage))
	}

	ginkgo.BeforeAll(func(ctx context.Context) {
		h = helper.New()
		log.Println("Test images to run: ", adHocTestImages)
	})
	ginkgo.BeforeEach(func(ctx context.Context) {
		ginkgo.By("Setting up new namespace")
		suffix = "h-" + util.RandomStr(5)
		subProject, err = h.SetupNewProject(ctx, suffix)
		Expect(err).NotTo(HaveOccurred(), "Could not set up test namespace")
		err = h.SetPassthruSecretInProject(ctx, subProject)
		Expect(err).NotTo(HaveOccurred(), "Could not set up passthru secrets")
	})

	ginkgo.DescribeTable("execution",
		func(ctx context.Context, adHocTestImage string) {
			log.Printf("======= RUNNING AD_HOC_TEST_IMAGE: %s =======", adHocTestImage)
			adHocTestImageIndex := strings.LastIndex(adHocTestImage, "/")
			adHocTestImageName := strings.Split(adHocTestImage[adHocTestImageIndex+1:], ":")[0]
			adHocTestImageTestTemplate := "tests/tests-runner.template"
			jobName := fmt.Sprintf("%s-%s", adHocTestImageName, suffix)

			// Create templated runner pod
			ginkgo.By("Creating test runner pod")
			h.SetServiceAccount(ctx, serviceAccount)
			r = h.RunnerWithNoCommand()
			h.SetRunnerProject(subProject.Name, r)
			latestImageStream, err := r.GetLatestImageStreamTag()
			Expect(err).NotTo(HaveOccurred(), "Could not get latest imagestream tag")
			cmd := h.GetRunnerCommandString(adHocTestImageTestTemplate, timeoutInSeconds, latestImageStream, adHocTestImage, suffix, jobName, serviceAccountDir, "", serviceAccountNamespacedName)
			r = h.SetRunnerCommand(cmd, r)

			ginkgo.By("Running test pod")
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
				err = h.UploadResultsToS3(results, adHocTestImageName+time.Now().Format(time.DateOnly+"_"+time.TimeOnly))
				if err != nil {
					ginkgo.GinkgoLogr.Error(err, "reporting error")
				}
			}
			// Adding test report failures to top level junit report
			for _, data := range results {
				suites, _ := junit.Ingest(data)
				for _, suite := range suites {
					for _, testcase := range suite.Tests {
						Expect(testcase.Error).To(BeNil(), "Assertion failed: "+testcase.Name)
					}
				}
			}
		},
		AdHocTestImageEntries)

	ginkgo.AfterEach(func(ctx context.Context) {
		if !viper.GetBool(config.Cluster.SkipDestroyCluster) {
			ginkgo.By("Deleting test namespace")
			err := h.DeleteProject(ctx, subProject.Name)
			if err != nil {
				ginkgo.GinkgoLogr.Error(err, fmt.Sprintf("error deleting project %q", subProject.Name))
			}
		} else {
			log.Printf("For debugging, see test namespace: %s", subProject.Name)
		}
	})
})
