package state

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/e2e/verify"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// cmd to collect prometheus data
	clusterStateInformingName = "[Suite: e2e] Cluster state"
)

var _ = ginkgo.Describe(clusterStateInformingName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	// How long to wait for the prometheus test job to finish
	prometheusTimeout := 15 * time.Minute
	// How long to wait for Prometheus pods to be running
	prometheusPodStartedDuration := 5 * time.Minute

	util.GinkgoIt("should include Prometheus data", func(ctx context.Context) {
		// setup runner
		// this command is has specific code to capture and suppress an exit code of
		// 1 as tar 1.26 will exit 1 if files change while the tar is running, as is
		// common for a running prometheus instance
		cmd := "oc cp -c prometheus openshift-monitoring/prometheus-k8s-0:/prometheus /tmp/prometheus && tar -cvzf " + runner.DefaultRunner.OutputDir + "/prometheus.tar.gz -C /tmp/prometheus .; err=$? ; if (( $err != 0 )) ; then exit $err ; fi"

		// ensure prometheus pods are up before trying to extract data
		poderr := wait.PollImmediate(2*time.Second, prometheusPodStartedDuration, func() (bool, error) {
			podCount := 0
			list, listerr := verify.FilterPods(ctx, "openshift-monitoring", "app.kubernetes.io/name=prometheus", h)
			if listerr != nil {
				return false, listerr
			}
			names, podNum := verify.GetPodNames(list, h)
			if podNum > 0 {
				for _, value := range names {
					if strings.HasPrefix(value, "prometheus-k8s-") {
						podCount += 1
					}
				}
				if podCount >= 2 {
					return true, nil
				}
			}
			return false, errors.New("prometheus pods are not in running state")
		})
		Expect(poderr).NotTo(HaveOccurred())

		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")

		r := h.Runner(cmd)
		r.Name = "collect-prometheus"

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(int(prometheusTimeout.Seconds()), stopCh)
		Expect(err).NotTo(HaveOccurred())

		// get results
		results, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())

		// write results
		h.WriteResults(results)
	}, prometheusTimeout.Seconds()+prometheusPodStartedDuration.Seconds())
})
