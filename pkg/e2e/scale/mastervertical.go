package scale

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kubev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/pkg/common/helper"
)

var _ = ginkgo.Describe("[Suite: scale-mastervertical] Scaling", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	masterVerticalTimeoutInSeconds := 7200
	ginkgo.It("should be tested with MasterVertical", func() {
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		// setup runner
		scaleCfg := scaleRunnerConfig{
			Name:         "master-vertical",
			PlaybookPath: "workloads/mastervertical.yml",
		}
		r := scaleCfg.Runner(h)

		// only test on 3 nodes
		r.PodSpec.Containers[0].Env = append(r.PodSpec.Containers[0].Env, kubev1.EnvVar{
			Name:  "MASTERVERTICAL_PROJECTS",
			Value: "100",
		})
		// run tests
		stopCh := make(chan struct{})
		err := r.Run(masterVerticalTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())
	}, float64(masterVerticalTimeoutInSeconds))
})
