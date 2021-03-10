// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var imageRegistryTestName string = "[Suite: openshift][image-registry]"
var imageEcosystemTestName string = "[Suite: openshift][image-ecosystem]"

func init() {
	alert.RegisterGinkgoAlert(imageRegistryTestName, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	alert.RegisterGinkgoAlert(imageEcosystemTestName, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(imageRegistryTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
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
		results, err := r.RetrieveTestResults()

		// write results
		h.WriteResults(results)

		// evaluate results
		Expect(err).NotTo(HaveOccurred())
	}, float64(e2eTimeoutInSeconds+30))
})

var _ = ginkgo.Describe(imageEcosystemTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		// configure tests
		cfg := DefaultE2EConfig
		cfg.Suite = "openshift/image-ecosystem"
		cmd := cfg.Cmd()

		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-image-ecosystem"

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
