package verify

import (
	"context"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
)

var (
	podsTestName        string = "[Suite: e2e] Pods"
	e2eTimeoutInSeconds int    = 3600
)

func init() {
	alert.RegisterGinkgoAlert(podsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(podsTestName, func() {
	h := helper.New()

	util.GinkgoIt("should not be Failed", func(ctx context.Context) {
		list, err := h.Kube().CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
		filteredList := &v1.PodList{}

		for _, pod := range list.Items {
			if pod.Status.Phase == v1.PodFailed {
				if len(pod.GetOwnerReferences()) > 0 {
					if pod.GetOwnerReferences()[0].Kind == "Job" {
						// Ignore failed jobs
						continue
					}
					if pod.GetOwnerReferences()[0].Kind == "ConfigMap" && strings.HasPrefix(pod.GetOwnerReferences()[0].Name, "revision-status-") {
						// Ignore failed pods owned by revision status config maps
						continue
					}
				}

				filteredList.Items = append(filteredList.Items, pod)
			}
		}
		Expect(err).NotTo(HaveOccurred(), "couldn't list Pods")
		Expect(filteredList).NotTo(BeNil())
		Expect(filteredList.Items).Should(HaveLen(0), "'%d' Pods are 'Failed'", len(filteredList.Items))
	}, float64(e2eTimeoutInSeconds))
})
