package osd

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var nodeLabelsTestName string = "[Suite: service-definition] [OSD] NodeLabels"

func init() {
	alert.RegisterGinkgoAlert(nodeLabelsTestName, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(nodeLabelsTestName, func() {
	ginkgo.Context("Modifying nodeLabels is not allowed", func() {
		// setup helper
		h := helper.New()
		ginkgo.It("node-label cannot be added", func() {
			// Set it to a wildcard dedicated-admin
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-cluster")

			nodes, err := h.Kube().CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(nodes.Items)).Should(BeNumerically(">", 0))

			node := nodes.Items[0]

			node.Labels["osde2e"] = "touched by osde2e"

			_, err = h.Kube().CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{})
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
