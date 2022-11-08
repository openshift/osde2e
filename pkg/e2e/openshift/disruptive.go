// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
)

var disruptiveTestName = "[Suite: openshift][disruptive]"

func init() {
	alert.RegisterGinkgoAlert(disruptiveTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

// Disruptive tests require SSH access to nodes.
var _ = ginkgo.Describe(disruptiveTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	e2eTimeoutInSeconds := 3600
	util.GinkgoIt("should run until completion", func(ctx context.Context) {
		// configure tests
		cfg := DefaultE2EConfig
		cfg.Suite = "openshift/disruptive"
		cmd := cfg.Cmd()
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")

		// setup runner
		r := h.Runner(cmd)

		r.Name = "openshift-disruptive"

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
