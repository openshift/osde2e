package verify

import (
	"context"
	"log"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/util"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

var userWorkloadMonitoringTestName string = "[Suite: informing] [OSD] User Workload Monitoring"

func init() {
	alert.RegisterGinkgoAlert(userWorkloadMonitoringTestName, "SD-SREP", "Max Whittingham", "hcm-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(userWorkloadMonitoringTestName, ginkgo.Ordered, label.Informing, func() {
	var h *helper.H
	ginkgo.BeforeAll(func() {
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("dedicated-admins group currently does not exist by default")
		}
		h = helper.New()
	})

	const (
		prometheusName = "prometheus-example-app"
	)

	userName := util.RandomStr(5) + "@customdomain"
	identities := []string{"otherIDP:testing_string"}
	groups := []string{"dedicated-admins"}
	uwmtestns := util.RandomStr(5)

	// How long to wait for expected resources to be created on-cluster
	uwmPollingDuration := 2 * time.Minute

	ginkgo.Context("User Workload Monitoring", func() {
		ginkgo.It("has the required prerequisites for testing", func(ctx context.Context) {
			// Create a new user that will have dedicated-admin privileges, add the user to the dedicated-admins group
			_, err := helper.CreateUser(ctx, userName, identities, groups, h)
			Expect(err).NotTo(HaveOccurred(), "Could not create user workload monitoring testing user")
			// Add the user to the dedicated-admin group
			_, err = helper.AddUserToGroup(ctx, userName, groups[0], h)
			Expect(err).NotTo(HaveOccurred(), "Could not grant dedicated-admin permissions to user workload monitoring user")
			// Create a namespace to run the tests in
			_, err = helper.CreateNamespace(ctx, uwmtestns, h)
			Expect(err).NotTo(HaveOccurred(), "Could not create user workload monitoring testing namespace")
			// Launch a pod & service as the targets of the ServiceMonitor and PrometheusRules objects
			pod := helper.SamplePod(prometheusName, uwmtestns, "quay.io/brancz/prometheus-example-app:v0.2.0")
			err = helper.CreatePod(ctx, pod, uwmtestns, h)
			Expect(err).NotTo(HaveOccurred(), "Could not create user workload monitoring testing pod")

			svc := helper.SampleService(8080, 8080, prometheusName, uwmtestns, prometheusName)
			err = helper.CreateService(ctx, svc, h)
			Expect(err).NotTo(HaveOccurred(), "Could not create user workload monitoring testing service")
		})

		ginkgo.It("has create access to the user-workload-monitoring-config configmap", func(ctx context.Context) {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: userName,
				Groups:   []string{"system:authenticated", "system:authenticated:oauth", "dedicated-admins"},
			})
			// Need a test to verify create/edit access to the user-workload-monitoring-config configmap
			uwme2ecm := newUwmCm("user-workload-monitoring-config", "openshift-user-workload-monitoring", "foo:bar")

			existingcm, err := h.Kube().CoreV1().ConfigMaps("openshift-user-workload-monitoring").Get(ctx, "user-workload-monitoring-config", metav1.GetOptions{})
			if err != nil {
				existingcm, err = h.Kube().CoreV1().ConfigMaps("openshift-user-workload-monitoring").Create(ctx, uwme2ecm, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "could not create user-workload-monitoring-config configmap")
				existingcm.Data["config.yaml"] = "2foo:2bar"
				_, err = h.Kube().CoreV1().ConfigMaps("openshift-user-workload-monitoring").Update(ctx, existingcm, metav1.UpdateOptions{})
				Expect(err).NotTo(HaveOccurred(), "could not edit user-workload-monitoring-config configmap")
				err = deleteUwmCM(ctx, h)
				Expect(err).NotTo(HaveOccurred(), "could not delete user-workload-monitoring-config configmap")

			}
			Expect(existingcm).NotTo(BeNil(), "Configmap user-workload-monitoring-config was created")
		})

		// Verify prometheus-operator pod && promethus-user-workload*/thanos-ruler-user-workload* pods are active
		ginkgo.It("has the required prometheus and thanos pods", func(ctx context.Context) {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: userName,
				Groups:   []string{"system:authenticated", "system:authenticated:oauth", "dedicated-admins"},
			})
			uwmpods, err := h.Kube().CoreV1().Pods("openshift-user-workload-monitoring").List(ctx, metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred(), "Did not find any user-workload-monitoring pods")
			Expect(uwmpods).NotTo(BeNil())
			// Regex prefix matching expected pods
			for _, someuwmpod := range uwmpods.Items {
				Expect(someuwmpod.Name).To(MatchRegexp("^prometheus-|^thanos"))
			}
		})

		// Verify a dedicated admin can create ServiceMonitor objects
		ginkgo.It("has access to create SerivceMonitor objects", func(ctx context.Context) {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: userName,
				Groups:   []string{"system:authenticated", "system:authenticated:oauth", "dedicated-admins"},
			})
			// Create ServiceMonitor
			uwme2esm := newServiceMonitor(prometheusName, uwmtestns)
			err := wait.PollUntilContextTimeout(ctx, time.Second*15, uwmPollingDuration, true, func(ctx context.Context) (bool, error) {
				_, err := h.Prometheusop().MonitoringV1().ServiceMonitors(uwmtestns).Create(ctx, uwme2esm, metav1.CreateOptions{})
				if err != nil {
					log.Printf("failed creating service monitor: %v", err)
					return false, nil
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred(), "Could not create ServiceMonitor")
		})

		// Verify a dedicated admin can create PrometheusRule objects
		ginkgo.It("has access to create PrometheusRule objects", func(ctx context.Context) {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: userName,
				Groups:   []string{"system:authenticated", "system:authenticated:oauth", "dedicated-admins"},
			})
			uwme2etestrule := newPrometheusRule(prometheusName, uwmtestns)
			err := wait.PollUntilContextTimeout(ctx, time.Second*15, uwmPollingDuration, true, func(ctx context.Context) (bool, error) {
				_, err := h.Prometheusop().MonitoringV1().PrometheusRules(uwmtestns).Create(ctx, uwme2etestrule, metav1.CreateOptions{})
				if err != nil {
					log.Printf("failed creating prometheus rules: %v", err)
					return false, nil
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred(), "Could not create PrometheusRules")
		})
	})
})

func newServiceMonitor(name, namespace string) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			JobLabel:          "",
			TargetLabels:      []string{},
			PodTargetLabels:   []string{},
			Endpoints:         []monitoringv1.Endpoint{},
			Selector:          metav1.LabelSelector{},
			NamespaceSelector: monitoringv1.NamespaceSelector{},
			SampleLimit:       0,
			TargetLimit:       0,
		},
	}
}

func newPrometheusRule(name, namespace string) *monitoringv1.PrometheusRule {
	uwme2eprometheusrule := &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: "example",
					Rules: []monitoringv1.Rule{
						{
							Alert: "VersionAlert",
							Expr: intstr.IntOrString{
								StrVal: "version{job==\"prometheus-example-app\"} == 0",
							},
						},
					},
				},
			},
		},
	}
	return uwme2eprometheusrule
}

func newUwmCm(name, namespace, cmConfigYaml string) *corev1.ConfigMap {
	configMapData := make(map[string]string)
	configMapData["config.yaml"] = cmConfigYaml
	uwmcm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Data:       configMapData,
	}
	return uwmcm
}

func deleteUwmCM(ctx context.Context, h *helper.H) error {
	return h.Kube().CoreV1().ConfigMaps("openshift-user-workload-monitoring").Delete(ctx, "user-workload-monitoring-config", metav1.DeleteOptions{})
}
