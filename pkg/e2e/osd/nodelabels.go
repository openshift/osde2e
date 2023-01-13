package osd

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"

	v1 "k8s.io/api/core/v1"
)

var nodeLabelsTestName string = "[Suite: service-definition] [OSD] Node labels"

func init() {
	alert.RegisterGinkgoAlert(nodeLabelsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(nodeLabelsTestName, ginkgo.Ordered, label.ServiceDefinition, func() {
	var h *helper.H

	ginkgo.BeforeAll(func() {
		h = helper.New()
	})

	ginkgo.It("cannot be modified after node creation", func(ctx context.Context) {
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("Clusterrole 'dedicated-admin' is not deployed to ROSA hosted-cp clusters")
		}

		client := h.AsServiceAccount(fmt.Sprintf("system:serviceaccount:%s:dedicated-admin-cluster", h.CurrentProject()))

		var nodes v1.NodeList
		err := client.List(ctx, &nodes)
		expect.NoError(err)
		Expect(len(nodes.Items)).To(BeNumerically(">", 0), "no nodes found in cluster")

		for _, node := range nodes.Items {
			node.Labels["osde2e"] = "modified by osde2e"
			err = client.Update(ctx, &node)
			expect.Forbidden(err)
		}
	})
})
