package operators

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var (
	osdMetricsExporterTestPrefix = "[Suite: operators] [OSD] OSD Metrics Exporter"
	osdMetricsExporterBasicTest  = osdMetricsExporterTestPrefix + " Basic Test"
)

func init() {
	alert.RegisterGinkgoAlert(osdMetricsExporterBasicTest, "SD_SREP", "@sre-platform-team-v1alpha1", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
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
