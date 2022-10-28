package operators

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"

	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var deploymentValidationOperatorTestName string = "[Suite: informing] [OSD] Deployment Validation Operator (dvo)"

func init() {
	alert.RegisterGinkgoAlert(deploymentValidationOperatorTestName, "SD-SREP", "Ron Green", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.FDescribe(deploymentValidationOperatorTestName, func() {
	const (
		operatorNamespace      = "openshift-deployment-validation-operator"
		operatorName           = "deployment-validation-operator"
		operatorDeploymentName = "deployment-validation-operator"
		operatorServiceName    = "deployment-validation-operator-metrics"
		operatorCsvDisplayName = "Deployment Validation Operator"
		testNamespace          = "osde2e-dvo-test"
		dvoString              = "\"msg\":\"Set memory requests and limits for your container based on its requirements. Refer to https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits for details.\",\"request.namespace\":\"osde2e-dvo-test\""
		operatorLockFile       = "deployment-validation-operator-lock"

		defaultDesiredReplicas int32 = 1
	)

	var clusterRoles = []string{
		"deployment-validation-operator-og-admin",
		"deployment-validation-operator-og-edit",
		"deployment-validation-operator-og-view",
	}

	h := helper.New()
	nodeLabels := make(map[string]string)

	checkDeployment(h, operatorNamespace, operatorDeploymentName, defaultDesiredReplicas)
	checkService(h, operatorNamespace, operatorServiceName, 8383)
	checkPod(h, operatorNamespace, operatorDeploymentName, 2, 3)
	checkClusterRoles(h, clusterRoles, false)

	util.GinkgoIt("Create and test deployment for DVO functionality", func() {

		//Create and check test deployment
		h.CreateProject("dvo-test")
		h.SetProjectByName("osde2e-dvo-test")

		// Set it to a wildcard dedicated-admin
		h.CreateServiceAccounts()
		h.SetServiceAccount("system:serviceaccount:osde2e-dvo-test:cluster-admin")

		// Test creating a basic deployment
		ds := makeDeployment("dvo-test-case", h.GetNamespacedServiceAccount(), nodeLabels)
		_, err := h.Kube().AppsV1().Deployments(h.CurrentProject()).Create(context.TODO(), &ds, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

	}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

	// Check the logs of DVO to assert the right test is flagging for test deployment
	checkPodLogs(h, operatorNamespace, testNamespace, operatorDeploymentName, operatorName, dvoString, 10)

	// Delete DVO Test Deployment
	deleteDVO(helper.New(), testNamespace)

})

// Function to create a standard deployment
func makeDeployment(name, sa string, nodeLabels map[string]string) appsv1.Deployment {
	matchLabels := make(map[string]string)
	matchLabels["name"] = name
	dep := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", name, util.RandomStr(5)),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   name,
					Labels: matchLabels,
				},
				Spec: v1.PodSpec{
					NodeSelector:       nodeLabels,
					ServiceAccountName: sa,
					Containers: []v1.Container{
						{
							Name:  "test",
							Image: "registry.k8s.io/echoserver:1.4",
						},
					},
				},
			},
		},
	}

	return dep
}

// Check Pod Logs to see if DVO pod is reporting correct metrics
func checkPodLogs(h *helper.H, namespace string, testNamespace string, name string, containerName string, dvoString string, gracePeriod int) {

	fmt.Println("Enterned Check Pod Logs")

	podLogOptions := v1.PodLogOptions{
		Container: containerName,
	}

	dvoCheck := false

	ginkgo.Context("pods", func() {
		util.GinkgoIt(fmt.Sprintf("Check logs in test namespace %s", testNamespace), func() {
			// wait for graceperiod
			fmt.Println("Waiting for grace period")
			// Wait for graceperiod
			time.Sleep(time.Duration(gracePeriod) * time.Second)
			// Retrieve pods
			pods, err := h.Kube().CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "name=" + name})
			Expect(err).ToNot(HaveOccurred(), "failed fetching pods")

			// Grab logs of pods
			fmt.Println("Grabbing Logs for pod")

			for _, pod := range pods.Items {
				logs := h.Kube().CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOptions)
				podLogs, err := logs.DoRaw(context.TODO())
				if err != nil {
					break
				}
				podString := string(podLogs)

				if strings.Contains(podString, dvoString) {
					dvoCheck = true
				} else {
					dvoCheck = false
					Expect(dvoCheck).NotTo(HaveOccurred())
				}
			}

		}, float64(gracePeriod)+viper.GetFloat64(config.Tests.PollingTimeout))
	})
}

// Delete DVO Test Deployment
func deleteDVO(h *helper.H, subNamespace string) {

	ginkgo.Context("Delete NS", func() {
		util.GinkgoIt("Delete NS used for testing DVO", func() {

			var err error
			err = helper.DeleteNamespace(subNamespace, true, h)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to delete project %s", subNamespace))
			log.Printf("Removed project %s", subNamespace)

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}
