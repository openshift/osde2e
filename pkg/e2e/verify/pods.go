package verify

import (
	"context"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

var podsTestName = "[Suite: e2e] Pods"

func init() {
	alert.RegisterGinkgoAlert(podsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(podsTestName, ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.It("should not be Failed", func(ctx context.Context) {
		list := &v1.PodList{}
		filteredList := &v1.PodList{}
		expect.NoError(client.WithNamespace(metav1.NamespaceAll).List(ctx, list))
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
		Expect(filteredList.Items).Should(HaveLen(0), "'%d' Pods are 'Failed'", len(filteredList.Items))
	})
})
