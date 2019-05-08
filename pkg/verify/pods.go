package verify

import (
	"fmt"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("Pods", func() {
	defer ginkgo.GinkgoRecover()
	_, cluster := NewCluster()

	ginkgo.It("should be mostly running", func() {
		requiredRatio := float64(95)

		list, err := cluster.Kube().CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list clusters")
		Expect(list).NotTo(BeNil())

		var running float64
		for _, pod := range list.Items {
			phase := pod.Status.Phase
			if phase == v1.PodRunning || phase == v1.PodSucceeded {
				running++
			}
		}

		ratio := (running / float64(len(list.Items))) * 100
		Expect(ratio).Should(BeNumerically(">", requiredRatio),
			"only %f%% of clusters ready, need %f%%", ratio, requiredRatio)
	})

	ginkgo.It("should not be Failed", func() {
		list, err := cluster.Kube().CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{
			FieldSelector: fmt.Sprintf("status.phase=%s", v1.PodFailed),
		})
		Expect(err).NotTo(HaveOccurred(), "couldn't list Pods")
		Expect(list).NotTo(BeNil())
		Expect(list.Items).Should(HaveLen(0), "'%d' Pods are 'Failed'", len(list.Items))
	})
})
