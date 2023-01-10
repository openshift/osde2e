package osd

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/pkg/common/alert"

	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

var privilegedTestname string = "[Suite: service-definition] [OSD] Privileged Containers"

func init() {
	alert.RegisterGinkgoAlert(privilegedTestname, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(privilegedTestname, ginkgo.Ordered, label.ServiceDefinition, label.HyperShift, func() {
	var h *helper.H

	ginkgo.BeforeAll(func() {
		h = helper.New()
	})

	makePod := func() v1.Pod {
		privileged := true

		return v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "osde2e-",
				Namespace:    h.CurrentProject(),
			},
			Spec: v1.PodSpec{
				ServiceAccountName: h.GetNamespacedServiceAccount(),
				Containers: []v1.Container{
					{
						Name:  "test",
						Image: "registry.access.redhat.com/ubi8/ubi-minimal",
						SecurityContext: &v1.SecurityContext{
							Privileged: &privileged,
						},
						Command: []string{"/bin/true"},
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
			},
		}
	}

	ginkgo.It("are not available by default", func(ctx context.Context) {
		client := h.AsServiceAccount(fmt.Sprintf("system:serviceaccount:%s:dedicated-admin-project", h.CurrentProject()))
		pod := makePod()
		err := client.Create(ctx, &pod)
		expect.Forbidden(err)
	})

	ginkgo.It("are only available for cluster-admin users", func(ctx context.Context) {
		client := h.AsServiceAccount(fmt.Sprintf("system:serviceaccount:%s:cluster-admin", h.CurrentProject()))
		pod := makePod()
		err := client.Create(ctx, &pod)
		expect.NoError(err)
		err = wait.For(conditions.New(client).PodPhaseMatch(&pod, v1.PodSucceeded))
		expect.NoError(err)
		err = client.Delete(ctx, &pod)
		expect.NoError(err)
	})
})
