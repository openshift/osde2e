package executioner_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	"github.com/openshift/osde2e/pkg/e2e/executioner"
	"k8s.io/klog/v2/textlogger"
)

var _ = Describe("Executioner", func() {
	BeforeEach(func() {
		if os.Getenv("KUBECONFIG") == "" {
			Skip("No KUBECONFIG available")
		}
	})

	It("should successfully execute", func(ctx SpecContext) {
		outputDir, err := os.MkdirTemp("", "osde2e-executioner-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(os.RemoveAll(outputDir)).To(Succeed())
		})

		logger := textlogger.NewLogger(textlogger.NewConfig())
		cfg := &executioner.Config{
			Image:               "quay.io/app-sre/route-monitor-operator-test-harness",
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

		exe, err := executioner.New(logger, cfg)
		Expect(err).NotTo(HaveOccurred())

		err = exe.Execute(ctx)
		Expect(err).NotTo(HaveOccurred())
	})
})
