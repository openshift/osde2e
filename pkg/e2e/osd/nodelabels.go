package osd

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var nodeLabelsTestName string = "[Suite: service-definition] [OSD] NodeLabels"

func init() {
	alert.RegisterGinkgoAlert(nodeLabelsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(nodeLabelsTestName, label.ServiceDefinition, func() {
	ginkgo.Context("Modifying nodeLabels is not allowed", func() {
		// setup helper
		h := helper.New()
		util.GinkgoIt("node-label cannot be added", func(ctx context.Context) {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-cluster")

			nodes, err := h.Kube().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(nodes.Items)).Should(BeNumerically(">", 0))

			node := nodes.Items[0]

			node.Labels["osde2e"] = "touched by osde2e"

			_, err = h.Kube().CoreV1().Nodes().Update(ctx, &node, metav1.UpdateOptions{})
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
