// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: openshift][image-",
		TeamOwner:        "SD-CICD",
		PrimaryContact:   "Jeffrey Sica",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 4,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name+"registry", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// configure tests
		cfg := DefaultE2EConfig
		cfg.Suite = "openshift/image-registry"
		cmd := cfg.Cmd()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-image-registry"

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

var _ = ginkgo.Describe(testAlert.Name+"ecosystem", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// configure tests
		cfg := DefaultE2EConfig
		cfg.Suite = "openshift/image-ecosystem"
		cmd := cfg.Cmd()

		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-image-ecosystem"

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
