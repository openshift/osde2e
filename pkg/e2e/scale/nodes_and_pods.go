package scale

import (
	"strconv"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kubev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

const (
	numNodesForNodeVertical = 3
)

var _ = ginkgo.Describe("[Suite: scale-nodes-and-pods] Scaling", func() {
	defer ginkgo.GinkgoRecover()
	ginkgo.BeforeEach(func() {
		alert.RegisterGinkgoAlert(ginkgo.CurrentGinkgoTestDescription().TestText, "SD-CICD", "Michael Wilson", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	})
	h := helper.New()

	nodeVerticalTimeoutInSeconds := 3600
	ginkgo.It("should be tested with NodeVertical", func() {
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
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
	ginkgo.It("should be tested with PodVertical", func() {
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
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
