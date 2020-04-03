package osd

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeDaemonSet(name, sa string, nodeLabels map[string]string) appsv1.DaemonSet {
	matchLabels := make(map[string]string)
	matchLabels["name"] = name
	ds := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   name,
					Labels: matchLabels,
				},
				Spec: v1.PodSpec{
					NodeSelector:       nodeLabels,
					ServiceAccountName: sa,
					Containers: []v1.Container{
						{
							Name:  "test",
							Image: "registry.access.redhat.com/ubi8/ubi-minimal",
						},
					},
				},
			},
		},
	}

	return ds
}

var _ = ginkgo.Describe("[Suite: service-definition] [OSD] DaemonSets", func() {
	ginkgo.Context("DaemonSets are not allowed", func() {
		// setup helper
		h := helper.New()
		nodeLabels := make(map[string]string)

		ginkgo.It("empty node-label daemonset should get created", func() {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			// Test creating a basic daemonset
			ds := makeDaemonSet("empty-node-labels", h.GetNamespacedServiceAccount(), nodeLabels)
			_, err := h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Create(&ds)
			Expect(err).NotTo(HaveOccurred())
		}, float64(config.Instance.Tests.PollingTimeout))

		ginkgo.It("worker node daemonset should get created", func() {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			// Test creating a worker daemonset
			nodeLabels["role"] = "worker"
			ds := makeDaemonSet("worker-node-labels", h.GetNamespacedServiceAccount(), nodeLabels)
			_, err := h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Create(&ds)
			Expect(err).NotTo(HaveOccurred())
		}, float64(config.Instance.Tests.PollingTimeout))

		ginkgo.It("infra node daemonset should get created", func() {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			// Test creating an infra daemonset
			nodeLabels["role"] = "infra"
			ds := makeDaemonSet("infra-node-labels", h.GetNamespacedServiceAccount(), nodeLabels)
			_, err := h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Create(&ds)
			Expect(err).NotTo(HaveOccurred())
		}, float64(config.Instance.Tests.PollingTimeout))
	})
})
