// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/helper"
	"github.com/openshift/osde2e/pkg/runner"
)

// DefaultE2EConfig is the base configuration for E2E runs.
var DefaultE2EConfig = E2EConfig{
	TestCmd: "run",
	Suite:   "kubernetes/conformance",
	Flags: []string{
		"--include-success",
		"--junit-dir=" + runner.DefaultRunner.OutputDir,
	},
}

var _ = ginkgo.Describe("[Suite: conformance] OpenShift E2E", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// configure tests
		cfg := DefaultE2EConfig
		cmd := cfg.Cmd()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-tests"

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(e2eTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())

		// write results
		h.WriteResults(results)
	}, float64(e2eTimeoutInSeconds+30))
})
