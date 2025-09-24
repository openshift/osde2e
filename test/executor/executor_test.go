package executor_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	"github.com/openshift/osde2e/pkg/common/executor"
	"k8s.io/klog/v2/textlogger"
)

var _ = Describe("Executor", func() {
	BeforeEach(func() {
		if os.Getenv("KUBECONFIG") == "" {
			Skip("No KUBECONFIG available")
		}
	})

	It("should successfully execute", func(ctx SpecContext) {
		outputDir, err := os.MkdirTemp("", "osde2e-executor-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(os.RemoveAll(outputDir)).To(Succeed())
		})

		logger := textlogger.NewLogger(textlogger.NewConfig())
		cfg := &executor.Config{
			OutputDir:           outputDir,
			Environment:         ocm.Stage,
			ClusterID:           "test-cluster",
			CloudProviderID:     "aws",
			CloudProviderRegion: "us-east-1",
			PassthruSecrets: map[string]string{
				"TESTING": "true",
			},
			Timeout: 5 * time.Minute,
		}

		exe, err := executor.New(logger, cfg)
		Expect(err).NotTo(HaveOccurred())

		results, err := exe.Execute(ctx, "quay.io/app-sre/route-monitor-operator-test-harness")
		Expect(err).NotTo(HaveOccurred())
		Expect(results).NotTo(BeNil())

		// Validate basic TestResults structure
		Expect(results.TotalTests).To(BeNumerically(">=", 0))
		Expect(results.PassedTests).To(BeNumerically(">=", 0))
		Expect(results.FailedTests).To(BeNumerically(">=", 0))
		Expect(results.SkippedTests).To(BeNumerically(">=", 0))
		Expect(results.ErrorTests).To(BeNumerically(">=", 0))
		Expect(results.Suites).NotTo(BeNil())
	})
})
