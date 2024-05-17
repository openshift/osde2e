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
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/util"
)

// DefaultE2EConfig is the base configuration for E2E runs.
var DefaultE2EConfig = E2EConfig{
	OutputDir: "/test-run-results",
	Tarball:   false,
	Suite:     "kubernetes/conformance",
	Flags: []string{
		"--include-success",
		"--junit-dir=" + runner.DefaultRunner.OutputDir,
	},
	ServiceAccountDir: "/var/run/secrets/kubernetes.io/serviceaccount",
}

var (
	conformanceK8sTestName       = "[Suite: conformance][k8s]"
	conformanceOpenshiftTestName = "[Suite: conformance][openshift]"
)

var _ = ginkgo.Describe(conformanceK8sTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 7200
	ginkgo.It("should run until completion", func(ctx context.Context) {
		// configure tests
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")

		cfg := DefaultE2EConfig
		cmd := cfg.GenerateOcpTestCmdBlock()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "k8s-conformance"

		// run tests
		stopCh := make(chan struct{})

		err := r.Run(e2eTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveTestResults()

		// write results
		h.WriteResults(results)

		// evaluate results
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = ginkgo.Describe(conformanceOpenshiftTestName, ginkgo.Ordered, label.OCPNightlyBlocking, func() {
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
		suffix := util.RandomStr(5)
		r.Name = "osde2e-main-" + suffix
		latestImageStream, err := r.GetLatestImageStreamTag()
		Expect(err).NotTo(HaveOccurred(), "Could not get latest imagestream tag")
		testcmd := cfg.GenerateOcpTestCmdBlock()
		cmd := h.GetRunnerCommandString("tests/ocp-tests-runner.template",
			e2eTimeoutInSeconds,
			latestImageStream,
			latestImageStream,
			suffix,
			"openshift-conformance",
			"/var/run/secrets/kubernetes.io/serviceaccount",
			testcmd,
			"cluster-admin")
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
