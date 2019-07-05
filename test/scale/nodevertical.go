// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("Scaling", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	ginkgo.It("should be tested with NodeVertical", func() {
		// setup runner
		scaleCfg := scaleRunnerConfig{
			Name:         "node-vertical",
			PlaybookPath: "workloads/nodevertical.yml",
		}
		r := scaleCfg.Runner(h)

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
