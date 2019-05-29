// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("OpenShift E2E", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	ginkgo.It("should run until completion", func() {
		// setup runner
		r := h.Runner()

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())

		// write results
		h.WriteResults(results)
	})
})
