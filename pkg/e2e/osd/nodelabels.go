package osd

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: service-definition] [OSD] NodeLabels",
		TeamOwner:        "SD-CICD",
		PrimaryContact:   "Jeffrey Sica",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 4,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name, func() {
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
