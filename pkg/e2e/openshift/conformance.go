// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
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

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: conformance]",
		TeamOwner:        "SD-CICD",
		PrimaryContact:   "Jeffrey Sica",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 1,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name+"[k8s]", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// configure tests
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")

		cfg := DefaultE2EConfig
		cmd := cfg.Cmd()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "k8s-conformance"

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

var _ = ginkgo.Describe(testAlert.Name+"[openshift]", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 7200
	ginkgo.It("should run until completion", func() {
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		// configure tests
		cfg := DefaultE2EConfig
		cfg.Suite = "openshift/conformance"
		cmd := cfg.Cmd()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-conformance"

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
