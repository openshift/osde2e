// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/util"
)

// DefaultE2EConfig is the base configuration for E2E runs.
var DefaultE2EConfig = E2EConfig{
	OutputDir: "/test-run-results",
	TestCmd:   "run",
	Tarball:   false,
	Suite:     "kubernetes/conformance",
	Flags: []string{
		"--include-success",
		"--junit-dir=" + runner.DefaultRunner.OutputDir,
	},
	Name:      "kubernetes-conformance",
	CA:        "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
	TokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
}

var (
	conformanceK8sTestName       string = "[Suite: conformance][k8s]"
	conformanceOpenshiftTestName string = "[Suite: conformance][openshift]"
)

func init() {
	alert.RegisterGinkgoAlert(conformanceK8sTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	alert.RegisterGinkgoAlert(conformanceOpenshiftTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(conformanceK8sTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 7200
	util.GinkgoIt("should run until completion", func(ctx context.Context) {
		// configure tests
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")

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
		results, err := r.RetrieveTestResults()

		// write results
		h.WriteResults(results)

		// evaluate results
		Expect(err).NotTo(HaveOccurred())
	}, float64(e2eTimeoutInSeconds+30))
})

var _ = ginkgo.Describe(conformanceOpenshiftTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 7200
	util.GinkgoIt("should run until completion", func(ctx context.Context) {
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
		// configure tests
		cfg := DefaultE2EConfig
		cfg.Suite = "openshift/conformance/parallel suite"
		cfg.Name = "openshift-conformance"
		cmd := cfg.Cmd()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-conformance"

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
	}, float64(e2eTimeoutInSeconds+30))
})
