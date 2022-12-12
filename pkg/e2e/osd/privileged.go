package osd

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/util"
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

var privilegedTestname string = "[Suite: service-definition] [OSD] Privileged Containers"

func init() {
	alert.RegisterGinkgoAlert(privilegedTestname, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(privilegedTestname, label.ServiceDefinition, func() {
	ginkgo.Context("Privileged containers are not allowed", func() {
		// setup helper
		h := helper.New()

		util.GinkgoIt("privileged container should not get created", func(ctx context.Context) {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")

			// Test creating a privileged pod and expect a failure
			pod := makePod("privileged-pod", h.GetNamespacedServiceAccount(), true)

			_, err := h.Kube().CoreV1().Pods(h.CurrentProject()).Create(ctx, &pod, metav1.CreateOptions{})
			Expect(err).To(HaveOccurred())

			// Test creating an unprivileged pod and expect success
			pod = makePod("unprivileged-pod", h.GetNamespacedServiceAccount(), false)
			_, err = h.Kube().CoreV1().Pods(h.CurrentProject()).Create(ctx, &pod, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
