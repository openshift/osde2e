package verify

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("Projects", func() {
	defer ginkgo.GinkgoRecover()
	_, cluster := NewCluster()

	ginkgo.It("Empty Project should be created", func() {
		_, err := cluster.Project().ProjectV1().Projects().Get(cluster.proj, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "project should have been created")
	})
})
