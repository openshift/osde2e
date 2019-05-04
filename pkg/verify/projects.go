package verify

import (
	"fmt"

	"github.com/onsi/ginkgo"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/config"
)

var _ = ginkgo.Describe("Projects", func() {
	defer ginkgo.GinkgoRecover()

	var cluster *Cluster
	ginkgo.BeforeEach(func() {
		cluster = newCluster(config.Cfg.Kubeconfig)
		cluster.Setup()
	})

	ginkgo.AfterEach(func() {
		cluster.Cleanup()
	})

	ginkgo.It("Empty Project should be created", func() {
		if _, err := cluster.Project().ProjectV1().Projects().Get(cluster.proj, metav1.GetOptions{}); err != nil {
			msg := fmt.Sprintf("project should have been created: %v", err)
			ginkgo.Fail(msg)
		}
	})
})
