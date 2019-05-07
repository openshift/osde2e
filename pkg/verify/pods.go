package verify

import (
	"fmt"

	"github.com/onsi/ginkgo"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("Pods", func() {
	defer ginkgo.GinkgoRecover()
	_, cluster := NewCluster()

	ginkgo.It("should be mostly running", func() {
		requiredRatio := float64(95)

		list, err := cluster.Kube().CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			ginkgo.Fail("Couldn't list Pods: " + err.Error())
		} else if list == nil {
			ginkgo.Fail("list should not be nil")
		} else {
			var running float64
			for _, pod := range list.Items {
				phase := pod.Status.Phase
				if phase == v1.PodRunning || phase == v1.PodSucceeded {
					running++
				}
			}

			ratio := (running / float64(len(list.Items))) * 100
			if ratio < requiredRatio {
				msg := fmt.Sprintf("only %f%% of clusters were ready, %f%% required ", ratio, requiredRatio)
				ginkgo.Fail(msg)
			}
		}
	})

	ginkgo.It("should not be Failed", func() {
		list, err := cluster.Kube().CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{
			FieldSelector: fmt.Sprintf("status.phase=%s", v1.PodFailed),
		})
		if err != nil {
			ginkgo.Fail("Couldn't list Pods: " + err.Error())
		} else if list == nil {
			ginkgo.Fail("list should not be nil")
		} else if len(list.Items) != 0 {
			msg := fmt.Sprintf("no Pods should have failed, '%d' are 'Failed'", len(list.Items))
			ginkgo.Fail(msg)
		}
	})
})
