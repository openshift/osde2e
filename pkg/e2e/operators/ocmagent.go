package operators

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	ocmAgentTestPrefix = "[Suite: informing] [OSD] OCM Agent Operator"
	ocmAgentBasicTest  = ocmAgentTestPrefix + " Basic Test"
)

func init() {
	alert.RegisterGinkgoAlert(ocmAgentBasicTest, "SD_SREP", "@ocm-agent-operator", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(ocmAgentBasicTest, func() {
	var (
		operatorNamespace                = "openshift-ocm-agent-operator"
		operatorName                     = "ocm-agent-operator"
		operatorRegistry                 = "ocm-agent-operator-registry"
		ocmAgentTokenRefName             = "ocm-access-token"
		ocmAgentCofigRefName             = "ocm-agent-config"
		ocmAgentDeploymentName           = "ocm-agent"
		ocmAgentServiceMonitorName       = "ocm-agent-metrics"
		ocmAgentServiceName              = "ocm-agent"
		ocmAgentNetworkpolicyName        = "ocm-agent-allow-only-alertmanager"
		ocmAgentDeploymentReplicas int32 = 1

		clusterRoles = []string{
			"ocm-agent-operator",
		}
		clusterRoleBindings = []string{
			"ocm-agent-operator",
		}
		// servicePort = 8081
	)
	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkDeployment(h, operatorNamespace, operatorName, 1)
	checkClusterRoles(h, clusterRoles, true)
	checkClusterRoleBindings(h, clusterRoleBindings, true)
	checkDeployment(h, operatorNamespace, ocmAgentDeploymentName, ocmAgentDeploymentReplicas)
	// checkService(h, operatorNamespace, operatorName, servicePort)
	checkUpgrade(helper.New(), operatorNamespace, operatorName, operatorName, operatorRegistry)
	checkPod(h, operatorNamespace, ocmAgentDeploymentName, 300, 3)

	ginkgo.Context("Reconcile resources", func() {

		// Waiting period to wait for OAO resources to be appear once deleted
		pollingDuration := 600 * time.Second
		util.GinkgoIt("ocm-agent deployment should be restored when it gets deleted", func() {

			err := deleteDeployment(ocmAgentDeploymentName, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())

			err = wait.Poll(30*time.Second, 5*time.Minute, func() (bool, error) {
				obj, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group: "apps", Version: "v1", Resource: "deployments",
				}).Namespace(operatorNamespace).Get(context.TODO(), ocmAgentDeploymentName, metav1.GetOptions{})
				if err != nil {
					return false, fmt.Errorf("unable to retrieve ocm-agent Deployment")
				}

				if obj.GetName() != ocmAgentDeploymentName {
					return false, fmt.Errorf("ocm-agent deployment not found")
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, pollingDuration.Seconds())

		util.GinkgoIt("ocm-agent-config configmap should be restored when it gets deleted", func() {

			err := deleteConfigMap(ocmAgentCofigRefName, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())

			err = wait.Poll(30*time.Second, 5*time.Minute, func() (bool, error) {
				obj, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group: "", Version: "v1", Resource: "configmaps",
				}).Namespace(operatorNamespace).Get(context.TODO(), ocmAgentCofigRefName, metav1.GetOptions{})
				if err != nil {
					return false, fmt.Errorf("unable to retrieve ocm-agent-config configmap")
				}

				if obj.GetName() != ocmAgentCofigRefName {
					return false, fmt.Errorf("ocm-agent-config configmap not found")
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, pollingDuration.Seconds())

		util.GinkgoIt("ocm-agent-token secret should be restored when it gets deleted", func() {

			err := deleteSecret(ocmAgentTokenRefName, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())

			err = wait.Poll(30*time.Second, 5*time.Minute, func() (bool, error) {
				obj, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group: "", Version: "v1", Resource: "secrets",
				}).Namespace(operatorNamespace).Get(context.TODO(), ocmAgentTokenRefName, metav1.GetOptions{})
				if err != nil {
					return false, fmt.Errorf("unable to retrieve ocm-agent-token secret")
				}

				if obj.GetName() != ocmAgentTokenRefName {
					return false, fmt.Errorf("ocm-agent-token secret not found")
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, pollingDuration.Seconds())

		util.GinkgoIt("ocm-agent-metrics servicemonitor should be restored when it gets deleted", func() {

			err := deleteServiceMonitor(ocmAgentServiceMonitorName, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())

			err = wait.Poll(30*time.Second, 5*time.Minute, func() (bool, error) {
				obj, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group: "monitoring.coreos.com", Version: "v1", Resource: "servicemonitors",
				}).Namespace(operatorNamespace).Get(context.TODO(), ocmAgentServiceMonitorName, metav1.GetOptions{})
				if err != nil {
					return false, fmt.Errorf("unable to retrieve ocm-agent-metrics servicemonitor")
				}

				if obj.GetName() != ocmAgentServiceMonitorName {
					return false, fmt.Errorf("ocm-agent-metrics servicemonitor not found")
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, pollingDuration.Seconds())

		util.GinkgoIt("ocm-agent service should be restored when it gets deleted", func() {

			err := deleteService(ocmAgentServiceName, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())

			err = wait.Poll(30*time.Second, 5*time.Minute, func() (bool, error) {
				obj, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group: "", Version: "v1", Resource: "services",
				}).Namespace(operatorNamespace).Get(context.TODO(), ocmAgentServiceName, metav1.GetOptions{})
				if err != nil {
					return false, fmt.Errorf("unable to retrieve ocm-agent service")
				}

				if obj.GetName() != ocmAgentServiceName {
					return false, fmt.Errorf("ocm-agent service not found")
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, pollingDuration.Seconds())

		util.GinkgoIt("ocm-agent-allow-only-alertmanager  networkpolicy should restored when it gets deleted", func() {

			err := deleteNetworkPolicy(ocmAgentNetworkpolicyName, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())

			err = wait.Poll(30*time.Second, 5*time.Minute, func() (bool, error) {
				obj, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies",
				}).Namespace(operatorNamespace).Get(context.TODO(), ocmAgentNetworkpolicyName, metav1.GetOptions{})
				if err != nil {
					return false, fmt.Errorf("unable to retrieve ocm-agent-allow-only-alertmanager  networkpolicy")
				}

				if obj.GetName() != ocmAgentNetworkpolicyName {
					return false, fmt.Errorf("ocm-agent-allow-only-alertmanager  networkpolicy not found")
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, pollingDuration.Seconds())
	})
})

func deleteDeployment(resourceName string, namespace string, h *helper.H) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "apps", Version: "v1", Resource: "deployments",
	}).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
}

func deleteConfigMap(resourceName string, namespace string, h *helper.H) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "", Version: "v1", Resource: "configmaps",
	}).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
}

func deleteSecret(resourceName string, namespace string, h *helper.H) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "", Version: "v1", Resource: "secrets",
	}).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
}

func deleteService(resourceName string, namespace string, h *helper.H) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "", Version: "v1", Resource: "services",
	}).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
}

func deleteServiceMonitor(resourceName string, namespace string, h *helper.H) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "monitoring.coreos.com", Version: "v1", Resource: "servicemonitors",
	}).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
}

func deleteNetworkPolicy(resourceName string, namespace string, h *helper.H) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies",
	}).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
}
