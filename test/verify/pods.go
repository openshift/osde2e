package verify

import (
	"fmt"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("Pods", func() {
	h := helper.New()

	ginkgo.It("should be Running or Succeeded", func() {
		err := h.PollForHealthyPods(60, 30)
		Expect(err).NotTo(HaveOccurred(), err)
	}, 300)

	ginkgo.It("should not be Failed", func() {
		list, err := h.Kube().CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{
			FieldSelector: fmt.Sprintf("status.phase=%s", v1.PodFailed),
		})
		Expect(err).NotTo(HaveOccurred(), "couldn't list Pods")
		Expect(list).NotTo(BeNil())
		Expect(list.Items).Should(HaveLen(0), "'%d' Pods are 'Failed'", len(list.Items))
	}, 300)
})
