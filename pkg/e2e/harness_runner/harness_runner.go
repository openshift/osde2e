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
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"
	serviceAccount    = "system:serviceaccount:%s:cluster-admin"
	TimeoutInSeconds  = viper.GetFloat64(config.Tests.PollingTimeout)
	harnesses         = strings.Split(viper.GetString(config.Tests.TestHarnesses), ",")
	h                 *helper.H
	HarnessEntries    []ginkgo.TableEntry
	harness           string
	r                 *runner.Runner
	jobName           string
	suffix            string
)

var _ = ginkgo.Describe("Test Harness", ginkgo.Ordered, ginkgo.ContinueOnFailure, label.TestHarness, func() {
	for _, harness := range harnesses {
		HarnessEntries = append(HarnessEntries, ginkgo.Entry("should run "+harness+" successfully", harness))
	}

	ginkgo.BeforeEach(func(ctx context.Context) {
		// Run harness in a new project
		viper.Set(config.Project, "")
		h = helper.New()
		h.SetServiceAccount(ctx, serviceAccount)
		suffix = util.RandomStr(5)
	})

	ginkgo.AfterEach(func(ctx context.Context) {
		// get results
		results, err := r.RetrieveTestResults()
		Expect(err).NotTo(HaveOccurred(), "Could not read results")

		// write results
		h.WriteResults(results)
		h.Cleanup(ctx)
	})

	ginkgo.DescribeTable("Executing Harness",
		func(ctx context.Context, harness string) {
			ginkgo.By("======= RUNNING HARNESS: " + harness + " =======")
			log.Printf("======= RUNNING HARNESS: %s =======", harness)
			harnessImageIndex := strings.LastIndex(harness, "/")
			harnessImage := harness[harnessImageIndex+1:]
			jobName := fmt.Sprintf("%s-%s", harnessImage, suffix)
			r = h.RunnerWithTemplateCommand(TimeoutInSeconds, harness, suffix, jobName, serviceAccountDir)

			// run tests
			stopCh := make(chan struct{})
			err := r.Run(int(TimeoutInSeconds), stopCh)
			Expect(err).NotTo(HaveOccurred(), "Could not run pod")

			// ensure job has not failed
			_, err = h.Kube().BatchV1().Jobs(r.Namespace).Get(ctx, jobName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "Harness job pods failed")

			ginkgo.By("======= FINISHED HARNESS: " + harness + " =======")
		},
		HarnessEntries)
})
