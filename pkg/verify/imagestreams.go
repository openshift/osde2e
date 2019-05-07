package verify

import (
	"fmt"

	"github.com/onsi/ginkgo"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("ImageStreams", func() {
	defer ginkgo.GinkgoRecover()
	_, cluster := NewCluster()

	ginkgo.It("should exist in the cluster", func() {
		list, err := cluster.Image().ImageV1().ImageStreams(metav1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			ginkgo.Fail("Couldn't list clusters: " + err.Error())
		} else if list == nil {
			ginkgo.Fail("list should not be nil")
		}

		minImages := 50
		if len(list.Items) < minImages {
			msg := fmt.Sprintf("wanted at least '%d' images but have only '%d'", minImages, len(list.Items))
			ginkgo.Fail(msg)
		}
	})
})
