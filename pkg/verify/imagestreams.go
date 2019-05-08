package verify

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("ImageStreams", func() {
	defer ginkgo.GinkgoRecover()
	_, cluster := NewCluster()

	ginkgo.It("should exist in the cluster", func() {
		list, err := cluster.Image().ImageV1().ImageStreams(metav1.NamespaceAll).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list ImageStreams")
		Expect(list).NotTo(BeNil())

		numImages := len(list.Items)
		minImages := 50
		Expect(numImages).Should(BeNumerically(">", minImages), "need more images")
	})
})
