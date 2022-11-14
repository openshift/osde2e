package operators

import (
	"context"
	"fmt"
	"time"

	// "reflect" this is needed when PR https://github.com/openshift/route-monitor-operator/pull/94 is merged

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega" // go-staticcheck ST1001  should not use dot imports

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"

	"k8s.io/apimachinery/pkg/util/wait"

	kubev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var routeMonitorOperatorTestName string = "[Suite: informing] [OSD] Route Monitor Operator (rmo)"

func init() {
	alert.RegisterGinkgoAlert(routeMonitorOperatorTestName, "SD-SREP", "@sre-platform-team-orange", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(routeMonitorOperatorTestName, func() {
	const (
		operatorNamespace      = "openshift-route-monitor-operator"
		operatorName           = "route-monitor-operator"
		operatorDeploymentName = "route-monitor-operator-controller-manager"
		operatorCsvDisplayName = "Route Monitor Operator"
		// operatorLockFile  = "route-monitor-operator-lock"

		defaultDesiredReplicas int32 = 1
	)

	clusterRoles := []string{
		"route-monitor-operator-admin",
		"route-monitor-operator-edit",
		"route-monitor-operator-view",
	}

	h := helper.New()

	checkClusterServiceVersion(h, operatorNamespace, operatorCsvDisplayName)
	// checkConfigMapLockfile(h,operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorDeploymentName, defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles, false)

	// should I create a new helper here? seems everyone else does but I am not sure
	checkUpgrade(helper.New(),
		operatorNamespace,
		operatorName,
		operatorName,
		"route-monitor-operator-registry")
	verifyExistingRouteMonitorsAreValid(h)
	testRouteMonitorCreationWorks(h)
})

func verifyExistingRouteMonitorsAreValid(h *helper.H) {
	ginkgo.Context("rmo Route Monitor Operator regression for console", func() {
		util.GinkgoIt("has all of the required resouces", func(ctx context.Context) {
			const (
				consoleNamespace = "openshift-route-monitor-operator"
				consoleName      = "console"
			)
			var err error
			// RouteMonitor is expected to be there as a selectorsyncset
			// see https://github.com/openshift/managed-cluster-config/blob/master/deploy/osd-route-monitor-operator/100-openshift-console.console.RouteMonitor.yaml
			_, err = h.Prometheusop().MonitoringV1().ServiceMonitors(consoleNamespace).Get(ctx, consoleName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "Could not get console serviceMonitor")
			_, err = h.Prometheusop().MonitoringV1().PrometheusRules(consoleNamespace).Get(ctx, consoleName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "Could not get console prometheusRule")
		})
	})
}

func testRouteMonitorCreationWorks(h *helper.H) {
	ginkgo.Context("rmo Route Monitor Operator integration test", func() {
		// How long to wait for service monitors to be created
		pollingDuration := 10 * time.Minute
		util.GinkgoIt("Creates and deletes a RouteMonitor to see if it works accordingly", func(ctx context.Context) {
			var (
				routeMonitorNamespace = h.CurrentProject()
				err                   error
			)
			const routeMonitorName = "routemonitor-e2e-test"

			ginkgo.By("Creating a pod, service and route to monitor with a ServiceMonitor and PrometheusRule")

			pod := helper.SamplePod(routeMonitorName, routeMonitorNamespace, "quay.io/jitesoft/nginx:mainline")
			err = helper.CreatePod(ctx, pod, routeMonitorNamespace, h)
			Expect(err).NotTo(HaveOccurred(), "Couldn't create a testing pod")
			phase := h.WaitForPodPhase(ctx, pod, kubev1.PodRunning, 6, time.Second*15)
			Expect(phase).To(Equal(kubev1.PodRunning), fmt.Sprintf("pod %s in ns %s is not running, current pod state is %v", routeMonitorName, routeMonitorNamespace, phase))

			svc := helper.SampleService(8080, 80, routeMonitorName, routeMonitorNamespace, routeMonitorName)
			err = helper.CreateService(ctx, svc, h)
			Expect(err).NotTo(HaveOccurred(), "Couldn't create a testing service")

			appRoute := helper.SampleRoute(routeMonitorName, routeMonitorNamespace)
			err = helper.CreateRoute(ctx, appRoute, routeMonitorNamespace, h)
			Expect(err).NotTo(HaveOccurred(), "Couldn't create application route")

			ginkgo.By("Creating a sample RouteMonitor to monitor the service")
			rmo := helper.SampleRouteMonitor(routeMonitorName, routeMonitorNamespace, h)
			err = helper.CreateRouteMonitor(ctx, rmo, routeMonitorNamespace, h)
			Expect(err).NotTo(HaveOccurred(), "Couldn't create application route monitor")

			err = wait.PollImmediate(15*time.Second, pollingDuration, func() (bool, error) {
				_, err = h.Prometheusop().MonitoringV1().ServiceMonitors(routeMonitorNamespace).Get(ctx, routeMonitorName, metav1.GetOptions{})
				if !k8serrors.IsNotFound(err) {
					return false, err
				}
				_, err = h.Prometheusop().MonitoringV1().PrometheusRules(routeMonitorNamespace).Get(ctx, routeMonitorName, metav1.GetOptions{})
				if !k8serrors.IsNotFound(err) {
					return false, err
				}
				return true, nil
			})

			Expect(err).NotTo(HaveOccurred(), "dependant resources weren't created via RouteMonitor")

			// //will be re-added when https://github.com/openshift/route-monitor-operator/pull/94 is ready in production (IT IS NOW, BUT TESTING IS REQUIRED BEFORE UNCOMMENT)
			// modifiedRmo, err := helper.GetRouteMonitor(routeMonitorName, routeMonitorNamespace, h)
			// Expect(err).NotTo(HaveOccurred(), "Couldn't get application route monitor")

			// modifiedRmo.Spec.Slo.TargetAvailabilityPercent = "99.9995"
			// err = helper.UpdateRouteMonitor(modifiedRmo, routeMonitorNamespace, h)
			// Expect(err).NotTo(HaveOccurred(), "Couldn't update application route monitor")

			// modifiedPromRule, err := h.Prometheusop().MonitoringV1().PrometheusRules(routeMonitorNamespace).Get(ctx, routeMonitorName, metav1.GetOptions{})
			// Expect(err).NotTo(HaveOccurred(), "Couldn't get sample prometheusRule, should have been generated from RouteMonitor")
			// areEqual := reflect.DeepEqual(promRule.Spec, modifiedPromRule.Spec)
			// Expect(areEqual).To(BeFalse(), "Modifying the RouteMonitor .spec.Slo.TargetAvailabilityPercent did not modify the PrometheusRule")

			ginkgo.By("Deleting the sample RouteMonitor")
			nsName := types.NamespacedName{Namespace: routeMonitorNamespace, Name: routeMonitorName}
			err = helper.DeleteRouteMonitor(ctx, nsName, true, h)
			Expect(err).NotTo(HaveOccurred(), "Couldn't delete application route monitor")

			_, err = h.Prometheusop().MonitoringV1().ServiceMonitors(routeMonitorNamespace).Get(ctx, routeMonitorName, metav1.GetOptions{})
			Expect(k8serrors.IsNotFound(err)).To(BeTrue(), "sample serviceMonitor still exists, deletion of RouteMonitor didn't clean it up")
			_, err = h.Prometheusop().MonitoringV1().PrometheusRules(routeMonitorNamespace).Get(ctx, routeMonitorName, metav1.GetOptions{})
			Expect(k8serrors.IsNotFound(err)).To(BeTrue(), "sample prometheusRule still exists, deletion of RouteMonitor didn't clean it up")
		}, pollingDuration.Seconds())
	})
}
