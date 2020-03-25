package state

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
)

var (
	// cmd to run must-gather
	mustGatherCmd = "oc adm must-gather --dest-dir=" + runner.DefaultRunner.OutputDir
)

var _ = ginkgo.Describe("[Suite: e2e] Cluster state", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	mustGatherTimeoutInSeconds := 900
	ginkgo.It("should be captured with must-gather", func() {
		// setup runner
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		r := h.Runner(mustGatherCmd)
		r.Name = "must-gather"
		r.Tarball = true

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(mustGatherTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())

		// write results
		h.WriteResults(results)
	}, float64(mustGatherTimeoutInSeconds+30))
})
