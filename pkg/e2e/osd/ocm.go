package osd

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/spf13/viper"
)

var ocmTestName string = "[Suite: informing] [OSD] OCM"

func init() {
	alert.RegisterGinkgoAlert(ocmTestName, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(ocmTestName, func() {
	ginkgo.Context("Metrics", func() {
		clusterID := viper.GetString(config.Cluster.ID)
		ginkgo.It("do exist and are not empty", func() {
			provider, err := providers.ClusterProvider()
			Expect(err).NotTo(HaveOccurred())

			metrics, err := provider.Metrics(clusterID)

			Expect(err).NotTo(HaveOccurred())
			Expect(metrics.CriticalAlertsFiring()).NotTo(BeNil())
			Expect(metrics.OperatorsConditionFailing()).NotTo(BeNil())
			Expect(metrics.Nodes().Compute()).NotTo(BeZero())
			Expect(metrics.Nodes().Infra()).NotTo(BeZero())
			Expect(metrics.Nodes().Master()).NotTo(BeZero())
			Expect(metrics.ComputeNodesSockets().Empty()).NotTo(BeFalse())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
