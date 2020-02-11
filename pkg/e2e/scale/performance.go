package scale

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kubev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/pkg/common/helper"
)

var _ = ginkgo.Describe("[Suite: scale-performance] Scaling", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	httpTimeoutInSeconds := 7200
	ginkgo.It("should be tested with HTTP", func() {
		// setup runner
		scaleCfg := scaleRunnerConfig{
			Name:         "http",
			PlaybookPath: "workloads/http.yml",
		}
		r := scaleCfg.Runner(h)

		r.PodSpec.Containers[0].Env = append(r.PodSpec.Containers[0].Env, kubev1.EnvVar{
			Name:  "WORKLOAD_JOB_NODE_SELECTOR",
			Value: "false",
		})

		// run tests
		stopCh := make(chan struct{})
		err := r.Run(httpTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())
	}, float64(httpTimeoutInSeconds))

	// TODO: Enable once the network test is fixed. Currently running into UPERF_SSHD_PORT not defined
	//networkTimeoutInSeconds := 7200
	//ginkgo.It("should be tested with Network", func() {
	//	// setup runner
	//	scaleCfg := scaleRunnerConfig{
	//		Name:         "network",
	//		PlaybookPath: "workloads/network.yml",
	//	}
	//	r := scaleCfg.Runner(h)

	//	r.PodSpec.Containers[0].Env = append(r.PodSpec.Containers[0].Env, kubev1.EnvVar{
	//		Name:  "WORKLOAD_JOB_NODE_SELECTOR",
	//		Value: "false",
	//	})

	//	// run tests
	//	stopCh := make(chan struct{})
	//	err := r.Run(networkTimeoutInSeconds, stopCh)
	//	Expect(err).NotTo(HaveOccurred())
	//}, float64(networkTimeoutInSeconds))

	// TODO: Reenable this once we can figure out how to get it working. It looks like this takes longer than 2.5 hours,
	//       so this may require being split into multiple tests
	//prometheusTimeoutInSeconds := 7200
	//ginkgo.It("should be tested with Prometheus", func() {
	//	// setup runner
	//	scaleCfg := scaleRunnerConfig{
	//		Name:         "prometheus",
	//		PlaybookPath: "workloads/prometheus.yml",
	//	}
	//	r := scaleCfg.Runner(h)

	//	r.PodSpec.Containers[0].Env = append(r.PodSpec.Containers[0].Env, kubev1.EnvVar{
	//		Name:  "WORKLOAD_JOB_NODE_SELECTOR",
	//		Value: "false",
	//	})

	//	// run tests
	//	stopCh := make(chan struct{})
	//	err := r.Run(prometheusTimeoutInSeconds, stopCh)
	//	Expect(err).NotTo(HaveOccurred())
	//}, float64(prometheusTimeoutInSeconds))

	// TODO: Enable once the fio test is fixed. Currently failing with 'azure_auth' is undefined
	//fioTimeoutInSeconds := 3600
	//ginkgo.It("should be tested with fio", func() {
	//	// setup runner
	//	scaleCfg := scaleRunnerConfig{
	//		Name:         "fio",
	//		PlaybookPath: "workloads/fio.yml",
	//	}
	//	r := scaleCfg.Runner(h)

	//	r.PodSpec.Containers[0].Env = append(r.PodSpec.Containers[0].Env, kubev1.EnvVar{
	//		Name:  "WORKLOAD_JOB_NODE_SELECTOR",
	//		Value: "false",
	//	})

	//	// run tests
	//	stopCh := make(chan struct{})
	//	err := r.Run(fioTimeoutInSeconds, stopCh)
	//	Expect(err).NotTo(HaveOccurred())
	//}, float64(fioTimeoutInSeconds))
})
