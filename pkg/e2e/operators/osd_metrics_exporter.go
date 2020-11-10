package operators

import (
	"context"
	"fmt"
	"net/http"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	osdMetricsExporterTestPrefix   = "[Suite: operators] [OSD] OSD Metrics Exporter"
	osdMetricsExporterBasicTest    = osdMetricsExporterTestPrefix + " Basic Test"
	osdMetricsExporterEndpointTest = osdMetricsExporterTestPrefix + " Endpoint Test"
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
	)
	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkDeployment(h, operatorNamespace, operatorName, 1)
	checkClusterRoles(h, clusterRoles, true)
	checkClusterRoleBindings(h, clusterRoleBindings, true)
	checkUpgrade(helper.New(), operatorNamespace, operatorName, operatorName, "osd-metrics-exporter-registry")
})

var _ = ginkgo.FDescribe(osdMetricsExporterEndpointTest, func() {
	var (
		operatorNamespace = "openshift-osd-metrics"
		operatorName      = "osd-metrics-exporter"
		servicePort       = 8383
	)
	h := helper.New()
	checkService(h, operatorNamespace, operatorName, servicePort)
})

func checkService(h *helper.H, namespace string, name string, port int) {
	pollTimeout := viper.GetFloat64(config.Tests.PollingTimeout)
	serviceEndpoint := fmt.Sprintf("http://%s.%s:%d/metrics", name, namespace, port)
	ginkgo.Context("service", func() {
		ginkgo.It(
			"should exist",
			func() {
				service, err := h.Kube().CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(service).NotTo(gomega.BeNil())
			},
			pollTimeout,
		)
		ginkgo.It(
			"should return response",
			func() {
				response, err := http.Get(serviceEndpoint)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(response).To(gomega.HaveHTTPStatus(http.StatusOK))
			},
			pollTimeout,
		)
	})
}
