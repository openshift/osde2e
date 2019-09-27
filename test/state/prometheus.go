package state

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/helper"
	"github.com/openshift/osde2e/pkg/runner"
)

const (
	// cmd to collect prometheus data
	promCollectCmd = "oc exec -n openshift-monitoring prometheus-k8s-0 -- tar cvzf - -C /prometheus ."
)

var _ = ginkgo.Describe("Cluster state", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	ginkgo.It("should include Prometheus data", func() {
		// setup runner
		cmd := promCollectCmd + " >" + runner.DefaultRunner.OutputDir + "/prometheus.tar.gz"
		r := h.Runner(cmd)
		r.Name = "collect-prometheus"

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())

		// write results
		h.WriteResults(results)
	}, 900)
})
