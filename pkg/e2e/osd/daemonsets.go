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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var daemonSetsTestName string = "[Suite: service-definition] [OSD] DaemonSets"

func init() {
	alert.RegisterGinkgoAlert(daemonSetsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(daemonSetsTestName, label.ServiceDefinition, func() {
	ginkgo.Context("DaemonSets are not allowed", func() {
		// setup helper
		h := helper.New()
		nodeLabels := make(map[string]string)

		util.GinkgoIt("empty node-label daemonset should get created", func(ctx context.Context) {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")

			// Test creating a basic daemonset
			ds := makeDaemonSet("empty-node-labels", h.GetNamespacedServiceAccount(), nodeLabels)
			_, err := h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Create(ctx, &ds, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		util.GinkgoIt("worker node daemonset should get created", func(ctx context.Context) {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")

			// Test creating a worker daemonset
			nodeLabels["role"] = "worker"
			ds := makeDaemonSet("worker-node-labels", h.GetNamespacedServiceAccount(), nodeLabels)
			_, err := h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Create(ctx, &ds, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		util.GinkgoIt("infra node daemonset should get created", func(ctx context.Context) {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")

			// Test creating an infra daemonset
			nodeLabels["role"] = "infra"
			ds := makeDaemonSet("infra-node-labels", h.GetNamespacedServiceAccount(), nodeLabels)
			_, err := h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Create(ctx, &ds, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})

// Test Helper Functions
func makeDaemonSet(name, sa string, nodeLabels map[string]string) appsv1.DaemonSet {
	matchLabels := make(map[string]string)
	matchLabels["name"] = name
	ds := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", name, util.RandomStr(5)),
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
