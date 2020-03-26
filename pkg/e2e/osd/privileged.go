package osd

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makePod(name, sa string, privileged bool) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			ServiceAccountName: sa,
			Containers: []v1.Container{
				{
					Name:  "test",
					Image: "registry.access.redhat.com/ubi8/ubi-minimal",
					SecurityContext: &v1.SecurityContext{
						Privileged: &privileged,
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("[Suite: service-definition] [OSD] Privileged Containers", func() {
	ginkgo.Context("Privileged containers are not allowed", func() {
		// setup helper
		h := helper.New()

		ginkgo.It("privileged container should not get created", func() {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			// Test creating a privileged pod and expect a failure
			pod := makePod("privileged-pod", h.GetNamespacedServiceAccount(), true)

			_, err := h.Kube().CoreV1().Pods(h.CurrentProject()).Create(&pod)
			Expect(err).To(HaveOccurred())

			// Test creating an unprivileged pod and expect success
			pod = makePod("unprivileged-pod", h.GetNamespacedServiceAccount(), false)
			_, err = h.Kube().CoreV1().Pods(h.CurrentProject()).Create(&pod)
			Expect(err).NotTo(HaveOccurred())

		}, float64(config.Instance.Tests.PollingTimeout))
	})
})
