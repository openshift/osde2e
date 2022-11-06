package scale

import (
	"context"
	"strconv"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kubev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
)

const (
	numNodesForNodeVertical = 3
)

var nodesPodsTestName string = "[Suite: scale-nodes-and-pods] Scaling"

func init() {
	alert.RegisterGinkgoAlert(nodesPodsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(nodesPodsTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	nodeVerticalTimeoutInSeconds := 3600
	util.GinkgoIt("should be tested with NodeVertical", func(ctx context.Context) {
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
		// setup runner
		scaleCfg := scaleRunnerConfig{
			Name:         "node-vertical",
			PlaybookPath: "workloads/nodevertical.yml",
		}
		r := scaleCfg.Runner(h)

		// only test on 3 nodes
		r.PodSpec.Containers[0].Env = append(r.PodSpec.Containers[0].Env, kubev1.EnvVar{
			Name:  "NODEVERTICAL_NODE_COUNT",
			Value: strconv.Itoa(numNodesForNodeVertical),
		}, kubev1.EnvVar{
			Name:  "NODEVERTICAL_MAXPODS",
			Value: strconv.Itoa(numNodesForNodeVertical * 250),
		}, kubev1.EnvVar{
			Name:  "EXPECTED_NODEVERTICAL_DURATION",
			Value: strconv.Itoa(nodeVerticalTimeoutInSeconds),
		})
		// run tests
		stopCh := make(chan struct{})
		err := r.Run(nodeVerticalTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())
	}, float64(nodeVerticalTimeoutInSeconds))

	podVerticalTimeoutInSeconds := 3600
	util.GinkgoIt("should be tested with PodVertical", func(ctx context.Context) {
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
		// setup runner
		scaleCfg := scaleRunnerConfig{
			Name:         "pod-vertical",
			PlaybookPath: "workloads/podvertical.yml",
		}
		r := scaleCfg.Runner(h)

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(podVerticalTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())
	}, float64(podVerticalTimeoutInSeconds))
})
