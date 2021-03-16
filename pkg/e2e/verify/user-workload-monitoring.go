package verify

import (
	"context"
	"log"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	kv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

var userWorkloadMonitoringTestName string = "[Suite: informing] [OSD] User Workload Monitoring"

func init() {
	alert.RegisterGinkgoAlert(userWorkloadMonitoringTestName, "SD-SREP", "Max Whittingham", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(userWorkloadMonitoringTestName, func() {
	h := helper.New()
	const (
		prometheusName = "prometheus-example-app"
	)

	userName := util.RandomStr(5) + "@customdomain"
	identities := []string{"otherIDP:testing_string"}
	groups := []string{"dedicated-admins"}
	uwmtestns := util.RandomStr(5)

	ginkgo.Context("User Workload Monitoring", func() {
		ginkgo.It("has the required prerequisites for testing", func() {
			//Create a new user that will have dedicated-admin privileges, add the user to the dedicated-admins group
			_, err := createUser(userName, identities, groups, h)
			Expect(err).NotTo(HaveOccurred(), "Could not create user workload monitoring testing user")
			//Add the user to the dedicated-admin group
			_, err = addUserToGroup(userName, groups[0], h)
			Expect(err).NotTo(HaveOccurred(), "Could not grant dedicated-admin permissions to user workload monitoring user")
			//Create a namespace to run the tests in
			_, err = createNamespace(uwmtestns, h)
			Expect(err).NotTo(HaveOccurred(), "Could not create user workload monitoring testing namespace")
			//Launch a pod & service as the targets of the ServiceMonitor and PrometheusRules objects
			pod := newPod(prometheusName, uwmtestns, "quay.io/brancz/prometheus-example-app:v0.2.0")
			err = createPodUwm(pod, uwmtestns, h)
			Expect(err).NotTo(HaveOccurred(), "Could not create user workload monitoring testing pod")
			svc := GenerateService(8080, prometheusName, uwmtestns, prometheusName)
			err = createService(svc, h)
			Expect(err).NotTo(HaveOccurred(), "Could not create user workload monitoring testing service")
		})

		ginkgo.It("has create access to the user-workload-monitoring-config configmap", func() {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: userName,
				Groups:   []string{"system:authenticated", "system:authenticated:oauth", "dedicated-admins"},
			})
			// Need a test to verify create/edit access to the user-workload-monitoring-config configmap
			uwme2ecm := newUwmCm("user-workload-monitoring-config", "openshift-user-workload-monitoring", "foo:bar")
			existingcm, err := h.Kube().CoreV1().ConfigMaps("openshift-user-workload-monitoring").Create(context.TODO(), uwme2ecm, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "could not create user-workload-monitoring-config configmap")
			existingcm.Data["config.yaml"] = "2foo:2bar"
			_, err = h.Kube().CoreV1().ConfigMaps("openshift-user-workload-monitoring").Update(context.TODO(), existingcm, metav1.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred(), "could not edit user-workload-monitoring-config configmap")
			err = deleteUwmCM(h)
			Expect(err).NotTo(HaveOccurred(), "could not delete user-workload-monitoring-config configmap")
		})

		//Verify prometheus-operator pod && promethus-user-workload*/thanos-ruler-user-workload* pods are active
		ginkgo.It("has the required prometheus and thanos pods", func() {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: userName,
				Groups:   []string{"system:authenticated", "system:authenticated:oauth", "dedicated-admins"},
			})
			uwmpods, err := h.Kube().CoreV1().Pods("openshift-user-workload-monitoring").List(context.TODO(), metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred(), "Did not find any user-workload-monitoring pods")
			Expect(uwmpods).NotTo(BeNil())
			//Regex prefix matching expected pods
			for _, someuwmpod := range uwmpods.Items {
				Expect(someuwmpod.Name).To(MatchRegexp("^prometheus-|^thanos"))
			}
		})
		//Verify a dedicated admin can create ServiceMonitor objects
		ginkgo.It("has access to create SerivceMonitor objects", func() {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: userName,
				Groups:   []string{"system:authenticated", "system:authenticated:oauth", "dedicated-admins"},
			})
			//Create ServiceMonitor
			uwme2esm := newServiceMonitor(prometheusName, uwmtestns)
			_, err := h.Prometheusop().MonitoringV1().ServiceMonitors(uwmtestns).Create(context.TODO(), uwme2esm, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Could not create ServiceMonitor")
		})

		//Verify a dedicated admin can create PrometheusRule objects
		ginkgo.It("has access to create PrometheusRule objects", func() {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: userName,
				Groups:   []string{"system:authenticated", "system:authenticated:oauth", "dedicated-admins"},
			})
			uwme2etestrule := newPrometheusRule(prometheusName, uwmtestns)
			_, err := h.Prometheusop().MonitoringV1().PrometheusRules(uwmtestns).Create(context.TODO(), uwme2etestrule, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Could not create PrometheusRules")
		})
	})

})

func createService(svc *kv1.Service, h *helper.H) error {
	uwm, err := h.Kube().CoreV1().Services(svc.Namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
	log.Printf("Result of the create command: (%v)", err)
	if err != nil {
		log.Printf("Could not issue create command")
		return err
	}

	// Wait for the pod to create.
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		if _, err := h.Kube().CoreV1().Services(uwm.Namespace).Get(context.TODO(), uwm.Name, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})
	return err
}

func createPodUwm(pod *kv1.Pod, namespace string, h *helper.H) error {
	uwm, err := h.Kube().CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	log.Printf("Result of the create command: (%v)", err)
	if err != nil {
		log.Printf("Could not issue create command")
		return err
	}

	// Wait for the pod to create.
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		if _, err := h.Kube().CoreV1().Pods(namespace).Get(context.TODO(), uwm.Name, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})
	return err
}

func newPod(name, namespace, imageName string) *kv1.Pod {
	return &kv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kv1.PodSpec{
			Containers: []kv1.Container{
				{
					Name:  "test",
					Image: imageName,
				},
			},
		},
	}
}

func newServiceMonitor(name, namespace string) *v1.ServiceMonitor {
	return &v1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceMonitorSpec{
			JobLabel:          "",
			TargetLabels:      []string{},
			PodTargetLabels:   []string{},
			Endpoints:         []v1.Endpoint{},
			Selector:          metav1.LabelSelector{},
			NamespaceSelector: v1.NamespaceSelector{},
			SampleLimit:       0,
			TargetLimit:       0,
		},
	}
}

func GenerateService(port int32, serviceName, serviceNamespace string, prometheusName string) *kv1.Service {
	service := &kv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      prometheusName,
			Namespace: serviceNamespace,
			Labels:    map[string]string{prometheusName: serviceName},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		Spec: kv1.ServiceSpec{
			Ports: []kv1.ServicePort{
				{
					Port:       8080,
					Protocol:   kv1.ProtocolTCP,
					TargetPort: intstr.IntOrString{IntVal: 8080},
					Name:       "web",
				},
			},
			Selector: map[string]string{prometheusName: serviceName},
		},
	}
	return service
}

func newPrometheusRule(name, namespace string) *v1.PrometheusRule {
	uwme2eprometheusrule := &v1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PrometheusRuleSpec{
			Groups: []v1.RuleGroup{
				{
					Name: "example",
					Rules: []v1.Rule{
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

func newUwmCm(name, namespace, cmConfigYaml string) *kv1.ConfigMap {
	configMapData := make(map[string]string)
	configMapData["config.yaml"] = cmConfigYaml
	uwmcm := &kv1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Data:       configMapData,
	}
	return uwmcm
}

func deleteUwmCM(h *helper.H) error {
	return h.Kube().CoreV1().ConfigMaps("openshift-user-workload-monitoring").Delete(context.TODO(), "user-workload-monitoring-config", metav1.DeleteOptions{})
}
