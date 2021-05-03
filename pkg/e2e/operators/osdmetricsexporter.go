package operators

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	osdMetricsExporterTestPrefix   = "[Suite: operators] [OSD] OSD Metrics Exporter"
	osdMetricsExporterBasicTest    = osdMetricsExporterTestPrefix + " Basic Test"
)

func init() {
	alert.RegisterGinkgoAlert(osdMetricsExporterBasicTest, "SD_SREP", "Arjun Naik", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(osdMetricsExporterBasicTest, func() {
	var (
		operatorNamespace = "openshift-osd-metrics"
		operatorName      = "osd-metrics-exporter"
		clusterRoles      = []string{
			"osd-metrics-exporter",
		}
		clusterRoleBindings = []string{
			"osd-metrics-exporter",
		}
		servicePort = 8383
	)
	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkDeployment(h, operatorNamespace, operatorName, 1)
	checkClusterRoles(h, clusterRoles, true)
	checkClusterRoleBindings(h, clusterRoleBindings, true)
	checkService(h, operatorNamespace, operatorName, servicePort)
	checkUpgrade(helper.New(), operatorNamespace, operatorName, operatorName, "osd-metrics-exporter-registry")
})

func checkService(h *helper.H, namespace string, name string, port int) {
	pollTimeout := viper.GetFloat64(config.Tests.PollingTimeout)
	ginkgo.Context("service", func() {
		ginkgo.It(
			"should exist",
			func() {
				Eventually(func() bool {
					_, err := h.Kube().CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
					if err != nil {
						return false
					}
					return true
				}, "30m", "1m").Should(BeTrue())
			},
			pollTimeout,
		)
	})
}
