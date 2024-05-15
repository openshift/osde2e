// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/util"
)

// DefaultE2EConfig is the base configuration for E2E runs.
var DefaultE2EConfig = E2EConfig{
	OutputDir: "/test-run-results",
	Tarball:   false,
	Suite:     "kubernetes/conformance",
	Flags: []string{
		"--include-success",
		"--junit-dir=/test-run-results",
	},
	ServiceAccountDir: "/var/run/secrets/kubernetes.io/serviceaccount",
}

var _ = ginkgo.Describe("[Suite: conformance][openshift]", ginkgo.Ordered, label.OCPNightlyBlocking, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 7200
	ginkgo.It("should run until completion", func(ctx context.Context) {
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
		// configure tests
		cfg := DefaultE2EConfig
		if viper.GetString(config.Tests.OCPTestSuite) != "" {
			cfg.Suite = "openshift/conformance/parallel " + viper.GetString(config.Tests.OCPTestSuite)
		} else {
			cfg.Suite = "openshift/conformance/parallel suite"
		}

		// setup runner
		r := h.RunnerWithNoCommand()
		latestImageStream, err := r.GetLatestImageStreamTag()
		Expect(err).NotTo(HaveOccurred(), "Could not get latest imagestream tag")
		testcmd := cfg.GenerateOcpTestCmdBlock()
		cmd := h.GetRunnerCommandString("tests/ocp-tests-runner.template", e2eTimeoutInSeconds, latestImageStream, latestImageStream, util.RandomStr(5), "openshift-conformance", cfg.ServiceAccountDir, testcmd, "cluster-admin")
		r = h.SetRunnerCommand(cmd, r)

		// run tests
		stopCh := make(chan struct{})
		err = r.Run(e2eTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveTestResults()

		// write results, including non-xml log files
		h.WriteResults(results)

		Expect(err).NotTo(HaveOccurred(), "Error reading xml results, test may have exited abruptly. Check conformance logs for errors")
	})
})
