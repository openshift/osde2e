package state

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
)

const (
	// cmd to collect prometheus data
	promCollectCmd            = "oc exec -n openshift-monitoring prometheus-k8s-0 -c prometheus -- /bin/sh -c \"cp -ruf /prometheus /tmp/ && tar cvzO -C /tmp/prometheus . \""
	clusterStateInformingName = "[Suite: informing] Cluster state"
)

var _ = ginkgo.Describe(clusterStateInformingName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	prometheusTimeoutInSeconds := 900
	ginkgo.It("should include Prometheus data", func() {
		// setup runner
		// this command is has specific code to capture and suppress an exit code of
		// 1 as tar 1.26 will exit 1 if files change while the tar is running, as is
		// common for a running prometheus instance
		cmd := promCollectCmd + " > " + runner.DefaultRunner.OutputDir + "/prometheus.tar.gz; err=$? ; if (( $err != 0 )) ; then exit $err ; fi"

		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		r := h.Runner(cmd)
		r.Name = "collect-prometheus"

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(prometheusTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())

		// write results
		h.WriteResults(results)
	}, float64(prometheusTimeoutInSeconds+30))
})
