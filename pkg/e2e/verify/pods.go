package verify

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
)

var podsTestName string = "[Suite: e2e] Pods"

func init() {
	alert.RegisterGinkgoAlert(podsTestName, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(podsTestName, func() {
	h := helper.New()

	util.GinkgoIt("should be Running or Succeeded", func() {
		var (
			interval = 30 * time.Second
			timeout  = 10 * time.Minute

			requiredRatio float64 = 100
			curRatio      float64
			notReady      []v1.Pod
		)

		err := wait.Poll(interval, timeout, func() (done bool, err error) {
			if curRatio != 0 {
				log.Printf("Checking that all Pods are running or completed (currently %f%%)...", curRatio)
			}

			list, err := h.Kube().CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return false, err
			}
			Expect(list).NotTo(BeNil())

			notReady = nil
			for _, pod := range list.Items {
				phase := pod.Status.Phase
				if phase != v1.PodRunning && phase != v1.PodSucceeded {
					if len(pod.GetOwnerReferences()) > 0 && pod.GetOwnerReferences()[0].Kind != "Job" {
						notReady = append(notReady, pod)
					}
				}
			}

			total := len(list.Items)
			ready := float64(total - len(notReady))
			curRatio = (ready / float64(total)) * 100

			return len(notReady) == 0, nil
		})

		msg := "only %f%% of Pods ready, need %f%%. Not ready: %s"
		Expect(err).NotTo(HaveOccurred(), msg, curRatio, requiredRatio, listPodPhases(notReady))
		Expect(curRatio).Should(Equal(requiredRatio), msg, curRatio, requiredRatio, listPodPhases(notReady))
	}, 300)

	util.GinkgoIt("should not be Failed", func() {
		list, err := h.Kube().CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		filteredList := &v1.PodList{}

		for _, pod := range list.Items {
			if pod.Status.Phase == v1.PodFailed {
				if len(pod.GetOwnerReferences()) > 0 && pod.GetOwnerReferences()[0].Kind != "Job" {
					filteredList.Items = append(filteredList.Items, pod)
				}
			}
		}
		Expect(err).NotTo(HaveOccurred(), "couldn't list Pods")
		Expect(filteredList).NotTo(BeNil())
		Expect(filteredList.Items).Should(HaveLen(0), "'%d' Pods are 'Failed'", len(filteredList.Items))
	}, 300)
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
