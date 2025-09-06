package verify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

const (
	nodePoolLabel = "hypershift.openshift.io/nodePool"
)

var nodePoolGVR = schema.GroupVersionResource{
	Group:    "hypershift.openshift.io",
	Version:  "v1beta1",
	Resource: "nodepools",
}

var _ = ginkgo.Describe("[Suite: e2e] NodePool STS Permissions", ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var (
		h                *helper.H
		client           *resources.Resources
		clusterNamespace string
		testNodePoolName string
		createdNodePool  bool
	)

	ginkgo.BeforeAll(func() {
		cloudProvider := viper.GetString(config.CloudProvider.CloudProviderID)
		if cloudProvider != "aws" {
			ginkgo.Skip(fmt.Sprintf("Tests only supported on AWS, got %s", cloudProvider))
		}

		h = helper.New()
		client = h.AsUser("")

		var err error
		clusterNamespace, err = getClusterNamespace(h)
		if err != nil {
			ginkgo.Skip(fmt.Sprintf("Cannot determine cluster namespace: %v", err))
		}

		testNodePoolName = fmt.Sprintf("sts-test-%d", time.Now().Unix()%100000)
	})

	ginkgo.AfterAll(func() {
		if createdNodePool && testNodePoolName != "" && clusterNamespace != "" {
			cleanupNodePool(context.Background(), h, clusterNamespace, testNodePoolName)
		}
	})

	ginkgo.It("validates existing nodes have proper AWS integration", func(ctx context.Context) {
		nodePoolNodes, err := findNodePoolNodes(ctx, client)
		if isSTSError(err) {
			failWithSTSError(err, "node discovery", "ec2:DescribeInstances")
		}
		expect.NoError(err, "Failed to find NodePool nodes")
		Expect(len(nodePoolNodes)).To(BeNumerically(">", 0), "No NodePool nodes found")

		for _, node := range nodePoolNodes {
			validateAWSIntegration(node)
		}
	})

	ginkgo.It("creates new NodePool successfully", func(ctx context.Context) {
		if clusterNamespace == "" {
			ginkgo.Skip("Cannot determine cluster namespace")
		}

		if !canAccessNodePoolAPI(h, clusterNamespace) {
			ginkgo.Skip("NodePool API not accessible from guest cluster")
		}

		nodePoolSpec := buildNodePoolSpec(testNodePoolName, clusterNamespace)
		nodePoolObj := &unstructured.Unstructured{Object: nodePoolSpec}

		_, err := h.Dynamic().Resource(nodePoolGVR).Namespace(clusterNamespace).
			Create(ctx, nodePoolObj, metav1.CreateOptions{})

		if isSTSError(err) {
			failWithSTSError(err, "NodePool creation", "ec2:RunInstances,ec2:CreateTags")
		}

		expect.NoError(err, "NodePool creation failed")
		createdNodePool = true
	})

	ginkgo.It("provisions nodes from new NodePool", func(ctx context.Context) {
		if !createdNodePool {
			ginkgo.Skip("NodePool creation was skipped")
		}

		err := waitForNodePoolReady(ctx, h, clusterNamespace, testNodePoolName, 10*time.Minute)
		if isSTSError(err) {
			failWithSTSError(err, "node provisioning", "ec2:RunInstances")
		}
		expect.NoError(err, "NodePool failed to provision nodes")

		provisionedNodes, err := findNodesForNodePool(ctx, client, testNodePoolName)
		expect.NoError(err, "Failed to find provisioned nodes")
		Expect(len(provisionedNodes)).To(BeNumerically(">", 0), "No nodes were provisioned")

		for _, node := range provisionedNodes {
			validateAWSIntegration(node)
		}
	})

	ginkgo.It("schedules workloads on NodePool nodes", func(ctx context.Context) {
		nodePoolNodes, err := findReadyNodePoolNodes(ctx, client)
		expect.NoError(err, "Failed to find ready NodePool nodes")
		Expect(len(nodePoolNodes)).To(BeNumerically(">", 0), "No ready NodePool nodes available")

		targetNode := &nodePoolNodes[0]
		pod := createTestPod(h.CurrentProject(), targetNode.Name)

		err = client.Create(ctx, pod)
		if isSTSError(err) {
			failWithSTSError(err, "workload creation", "ec2:RunInstances")
		}
		expect.NoError(err, "Failed to create test pod")
		defer client.Delete(ctx, pod)

		err = wait.For(conditions.New(client).PodPhaseMatch(pod, corev1.PodSucceeded),
			wait.WithTimeout(3*time.Minute), wait.WithInterval(5*time.Second))
		if isSTSError(err) {
			failWithSTSError(err, "workload scheduling", "ec2:RunInstances")
		}
		expect.NoError(err, "Workload failed to complete")
	})

	ginkgo.It("handles invalid configurations gracefully", func(ctx context.Context) {
		pod := createPodWithSelector(h.CurrentProject(), map[string]string{
			nodePoolLabel: "non-existent-nodepool",
		})

		err := client.Create(ctx, pod)
		expect.NoError(err, "Should create pod with invalid selector")
		defer client.Delete(ctx, pod)

		time.Sleep(30 * time.Second)

		podList := &corev1.PodList{}
		err = client.WithNamespace(pod.Namespace).List(ctx, podList)
		expect.NoError(err, "Failed to list pods")

		var foundPod *corev1.Pod
		for _, p := range podList.Items {
			if p.Name == pod.Name {
				foundPod = &p
				break
			}
		}

		Expect(foundPod).ToNot(BeNil(), "Pod should exist")
		Expect(foundPod.Status.Phase).To(Equal(corev1.PodPending), "Pod should remain pending")
	})

	ginkgo.It("handles resource-intensive workloads", func(ctx context.Context) {
		nodePoolNodes, err := findReadyNodePoolNodes(ctx, client)
		expect.NoError(err, "Failed to find ready NodePool nodes")
		Expect(len(nodePoolNodes)).To(BeNumerically(">", 0), "No ready NodePool nodes available")

		targetNode := &nodePoolNodes[0]
		pod := createResourceIntensivePod(h.CurrentProject(), targetNode.Name)

		err = client.Create(ctx, pod)
		expect.NoError(err, "Failed to create resource-intensive pod")
		defer client.Delete(ctx, pod)

		err = wait.For(conditions.New(client).PodPhaseMatch(pod, corev1.PodSucceeded),
			wait.WithTimeout(5*time.Minute), wait.WithInterval(5*time.Second))
		expect.NoError(err, "Resource-intensive workload failed")
	})

	ginkgo.It("rejects operations on non-existent NodePools", func(ctx context.Context) {
		if clusterNamespace == "" || !canAccessNodePoolAPI(h, clusterNamespace) {
			ginkgo.Skip("NodePool API not accessible")
		}

		_, err := h.Dynamic().Resource(nodePoolGVR).Namespace(clusterNamespace).
			Get(ctx, "non-existent-nodepool", metav1.GetOptions{})

		Expect(err).To(HaveOccurred(), "Should fail accessing non-existent NodePool")
		Expect(apierrors.IsNotFound(err)).To(BeTrue(), "Should return NotFound error")
	})

	ginkgo.It("rejects duplicate NodePool creation", func(ctx context.Context) {
		if clusterNamespace == "" || !canAccessNodePoolAPI(h, clusterNamespace) {
			ginkgo.Skip("NodePool API not accessible")
		}

		if !createdNodePool {
			ginkgo.Skip("No NodePool created to test duplication")
		}

		duplicateSpec := buildNodePoolSpec(testNodePoolName, clusterNamespace)
		duplicateObj := &unstructured.Unstructured{Object: duplicateSpec}

		_, err := h.Dynamic().Resource(nodePoolGVR).Namespace(clusterNamespace).
			Create(ctx, duplicateObj, metav1.CreateOptions{})

		Expect(err).To(HaveOccurred(), "Should fail creating duplicate NodePool")
		Expect(apierrors.IsAlreadyExists(err)).To(BeTrue(), "Should return AlreadyExists error")
	})
})

func findNodePoolNodes(ctx context.Context, client *resources.Resources) ([]corev1.Node, error) {
	var nodeList corev1.NodeList
	if err := client.List(ctx, &nodeList); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var nodePoolNodes []corev1.Node
	for _, node := range nodeList.Items {
		if _, hasLabel := node.Labels[nodePoolLabel]; hasLabel {
			nodePoolNodes = append(nodePoolNodes, node)
		}
	}
	return nodePoolNodes, nil
}

func findReadyNodePoolNodes(ctx context.Context, client *resources.Resources) ([]corev1.Node, error) {
	nodes, err := findNodePoolNodes(ctx, client)
	if err != nil {
		return nil, err
	}

	var readyNodes []corev1.Node
	for _, node := range nodes {
		if isNodeReady(node) {
			readyNodes = append(readyNodes, node)
		}
	}
	return readyNodes, nil
}

func findNodesForNodePool(ctx context.Context, client *resources.Resources, nodePoolName string) ([]corev1.Node, error) {
	var nodeList corev1.NodeList
	if err := client.List(ctx, &nodeList); err != nil {
		return nil, err
	}

	var matchingNodes []corev1.Node
	for _, node := range nodeList.Items {
		if labelValue, exists := node.Labels[nodePoolLabel]; exists {
			if strings.Contains(labelValue, nodePoolName) {
				matchingNodes = append(matchingNodes, node)
			}
		}
	}
	return matchingNodes, nil
}

func validateAWSIntegration(node corev1.Node) {
	Expect(node.Spec.ProviderID).To(HavePrefix("aws://"),
		fmt.Sprintf("Node %s missing AWS provider ID", node.Name))

	hasInternalIP := false
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP && addr.Address != "" {
			hasInternalIP = true
			break
		}
	}
	Expect(hasInternalIP).To(BeTrue(), fmt.Sprintf("Node %s missing InternalIP", node.Name))

	Expect(node.Labels["node.kubernetes.io/instance-type"]).ToNot(BeEmpty(),
		fmt.Sprintf("Node %s missing instance type", node.Name))
}

func getClusterNamespace(h *helper.H) (string, error) {
	if name := viper.GetString(config.Cluster.Name); name != "" {
		return "clusters-" + name, nil
	}

	current := h.CurrentProject()
	if strings.HasPrefix(current, "clusters-") {
		return current, nil
	}

	return "", fmt.Errorf("could not determine cluster namespace")
}

func canAccessNodePoolAPI(h *helper.H, namespace string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := h.Dynamic().Resource(nodePoolGVR).Namespace(namespace).List(ctx, metav1.ListOptions{Limit: 1})
	return err == nil
}

func buildNodePoolSpec(name, namespace string) map[string]interface{} {
	clusterName := strings.TrimPrefix(namespace, "clusters-")

	return map[string]interface{}{
		"apiVersion": "hypershift.openshift.io/v1beta1",
		"kind":       "NodePool",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"clusterName": clusterName,
			"replicas":    1,
			"management": map[string]interface{}{
				"autoRepair":  true,
				"upgradeType": "Replace",
			},
			"platform": map[string]interface{}{
				"aws": map[string]interface{}{
					"instanceType": "m5.large",
				},
			},
		},
	}
}

func waitForNodePoolReady(ctx context.Context, h *helper.H, namespace, nodePoolName string, timeout time.Duration) error {
	return wait.For(func(ctx context.Context) (bool, error) {
		np, err := h.Dynamic().Resource(nodePoolGVR).Namespace(namespace).
			Get(ctx, nodePoolName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		ready, found, err := unstructured.NestedInt64(np.Object, "status", "readyReplicas")
		if err != nil || !found {
			return false, nil
		}

		return ready >= 1, nil
	}, wait.WithTimeout(timeout), wait.WithInterval(30*time.Second))
}

func cleanupNodePool(ctx context.Context, h *helper.H, namespace, nodePoolName string) {
	err := h.Dynamic().Resource(nodePoolGVR).Namespace(namespace).
		Delete(ctx, nodePoolName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		ginkgo.GinkgoLogr.Error(err, "Failed to cleanup test NodePool", "name", nodePoolName)
	}

	wait.For(func(ctx context.Context) (bool, error) {
		_, err := h.Dynamic().Resource(nodePoolGVR).Namespace(namespace).
			Get(ctx, nodePoolName, metav1.GetOptions{})
		return apierrors.IsNotFound(err), nil
	}, wait.WithTimeout(3*time.Minute), wait.WithInterval(10*time.Second))
}

func createTestPod(namespace, nodeName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "nodepool-test-",
			Namespace:    namespace,
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
			Containers: []corev1.Container{{
				Name:    "test",
				Image:   "registry.access.redhat.com/ubi8/ubi-minimal:latest",
				Command: []string{"/bin/sh", "-c", "echo 'NodePool test successful' && sleep 5"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
				},
			}},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
}

func createPodWithSelector(namespace string, nodeSelector map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "invalid-selector-test-",
			Namespace:    namespace,
		},
		Spec: corev1.PodSpec{
			NodeSelector: nodeSelector,
			Containers: []corev1.Container{{
				Name:    "test",
				Image:   "registry.access.redhat.com/ubi8/ubi-minimal:latest",
				Command: []string{"/bin/sh", "-c", "sleep 30"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("50m"),
						corev1.ResourceMemory: resource.MustParse("64Mi"),
					},
				},
			}},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
}

func createResourceIntensivePod(namespace, nodeName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "resource-intensive-test-",
			Namespace:    namespace,
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
			Containers: []corev1.Container{{
				Name:    "resource-test",
				Image:   "registry.access.redhat.com/ubi8/ubi-minimal:latest",
				Command: []string{"/bin/sh", "-c", "echo 'Resource test' && sleep 10"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("500m"),
						corev1.ResourceMemory: resource.MustParse("512Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			}},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
}

func isSTSError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	patterns := []string{
		"accessdenied", "unauthorizedoperation", "forbidden",
		"invalid iam role", "sts permissions", "not authorized",
		"assumerolewithwebidentity", "invalididentitytoken",
		"authfailure", "signaturedoesnotmatch",
	}
	for _, pattern := range patterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

func failWithSTSError(err error, operation, permissions string) {
	ginkgo.Fail(fmt.Sprintf(
		"STS_PERMISSION_ERROR: %s failed due to missing AWS permissions (%s): %v. "+
		"This blocks release until STS policies are updated.",
		operation, permissions, err))
}

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}