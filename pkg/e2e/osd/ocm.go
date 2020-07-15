package osd

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/spf13/viper"
)

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: informing] [OSD] OCM",
		TeamOwner:        "SD-CICD",
		PrimaryContact:   "Jeffrey Sica",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 4,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name, func() {
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
			Expect(metrics.ComputeNodesCPU().Total().Empty()).NotTo(BeFalse())
			Expect(metrics.ComputeNodesSockets().Empty()).NotTo(BeFalse())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
