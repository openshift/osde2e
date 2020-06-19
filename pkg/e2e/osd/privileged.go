package osd

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makePod(name, sa string, privileged bool) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", name, util.RandomStr(5)),
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

			_, err := h.Kube().CoreV1().Pods(h.CurrentProject()).Create(context.TODO(), &pod, metav1.CreateOptions{})
			Expect(err).To(HaveOccurred())

			// Test creating an unprivileged pod and expect success
			pod = makePod("unprivileged-pod", h.GetNamespacedServiceAccount(), false)
			_, err = h.Kube().CoreV1().Pods(h.CurrentProject()).Create(context.TODO(), &pod, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
