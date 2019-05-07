package verify

import (
	"fmt"

	"github.com/onsi/ginkgo"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("Projects", func() {
	defer ginkgo.GinkgoRecover()
	_, cluster := NewCluster()

	ginkgo.It("Empty Project should be created", func() {
		if _, err := cluster.Project().ProjectV1().Projects().Get(cluster.proj, metav1.GetOptions{}); err != nil {
			msg := fmt.Sprintf("project should have been created: %v", err)
			ginkgo.Fail(msg)
		}
	})
})
