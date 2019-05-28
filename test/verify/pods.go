package verify

import (
	"fmt"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("Pods", func() {
	h := helper.New()

	ginkgo.It("should be Running or Succeeded", func() {
		requiredRatio := float64(100)

		list, err := h.Kube().CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list Pods")
		Expect(list).NotTo(BeNil())

		var notReady []v1.Pod
		for _, pod := range list.Items {
			phase := pod.Status.Phase
			if phase != v1.PodRunning && phase != v1.PodSucceeded {
				notReady = append(notReady, pod)
			}
		}

		total := float64(len(list.Items))
		ready := total - float64(len(notReady))
		ratio := (ready / total) * 100
		Expect(ratio).Should(Equal(requiredRatio),
			"only %f%% of Pods ready, need %f%%. Not ready: %s", ratio, requiredRatio, listPodPhases(notReady))
	})

	ginkgo.It("should not be Failed", func() {
		list, err := h.Kube().CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{
			FieldSelector: fmt.Sprintf("status.phase=%s", v1.PodFailed),
		})
		Expect(err).NotTo(HaveOccurred(), "couldn't list Pods")
		Expect(list).NotTo(BeNil())
		Expect(list.Items).Should(HaveLen(0), "'%d' Pods are 'Failed'", len(list.Items))
	})
})

func listPodPhases(pods []v1.Pod) (out string) {
	for i, pod := range pods {
		if i != 0 {
			out += ", "
		}
		out += fmt.Sprintf("%s/%s (Phase: %s)", pod.Namespace, pod.Name, pod.Status.Phase)
	}
	return
}
