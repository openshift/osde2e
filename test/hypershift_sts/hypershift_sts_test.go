package hypershift_sts

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

const (
	nodePoolLabel = "hypershift.openshift.io/nodePool"
	awsProvider   = "aws"
)

var nodePoolGVR = schema.GroupVersionResource{
	Group:    "hypershift.openshift.io",
	Version:  "v1beta1",
	Resource: "nodepools",
}

var _ = Describe("HyperShift STS Permissions", Ordered, Label("HyperShiftSTS"), func() {
	var (
		logger           = GinkgoLogr
		dynamicClient    dynamic.Interface
		client           *resources.Resources
		clusterNamespace string
		testNodePoolName string
		canAccessAPI     bool
		currentNamespace string
	)

	BeforeAll(func(ctx context.Context) {
		// Environment checks
		if os.Getenv("KUBECONFIG") == "" {
			Skip("No KUBECONFIG available")
		}

		cloudProvider := os.Getenv("CLOUD_PROVIDER_ID")
		if cloudProvider == "" {
			cloudProvider = "aws" // Default assumption for HyperShift
		}
		if cloudProvider != awsProvider {
			Skip(fmt.Sprintf("Tests only supported on AWS, got %s", cloudProvider))
		}

		// Initialize Kubernetes clients
		config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		Expect(err).NotTo(HaveOccurred(), "Failed to build kubeconfig")

		dynamicClient, err = dynamic.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred(), "Failed to create dynamic client")

		client, err = resources.New(config)
		Expect(err).NotTo(HaveOccurred(), "Failed to create resources client")

		// Get current namespace for pod operations
		if ns := os.Getenv("NAMESPACE"); ns != "" {
			currentNamespace = ns
		} else {
			currentNamespace = "default"
		}

		// Determine cluster namespace
		clusterNamespace, err = getClusterNamespace()
		if err != nil {
			Skip(fmt.Sprintf("Cannot determine cluster namespace: %v", err))
		}

		canAccessAPI = canAccessNodePoolAPI(dynamicClient, clusterNamespace)
		testNodePoolName = fmt.Sprintf("sts-test-%d", time.Now().Unix()%100000)

		logger.Info("Test setup complete",
			"canAccessAPI", canAccessAPI,
			"testNodePoolName", testNodePoolName,
			"clusterNamespace", clusterNamespace,
			"currentNamespace", currentNamespace)
	})

	It("validates existing nodes have proper AWS integration", func(ctx context.Context) {
		logger.Info("Starting existing nodes validation")
		nodePoolNodes, err := findNodePoolNodes(ctx, client)
		failIfSTSError(err, "node discovery", "ec2:DescribeInstances")
		Expect(err).NotTo(HaveOccurred(), "Failed to find NodePool nodes")
		Expect(len(nodePoolNodes)).To(BeNumerically(">", 0), "No NodePool nodes found")

		logger.Info("Validating AWS integration", "nodeCount", len(nodePoolNodes))
		for _, node := range nodePoolNodes {
			validateAWSIntegration(node)
		}
	})

	Describe("NodePool Node Validation", func() {
		It("schedules workloads on NodePool nodes", func(ctx context.Context) {
			logger.Info("Testing workload scheduling")
			nodePoolNodes, err := findReadyNodePoolNodes(ctx, client)
			Expect(err).NotTo(HaveOccurred(), "Failed to find ready NodePool nodes")
			Expect(len(nodePoolNodes)).To(BeNumerically(">", 0), "No ready NodePool nodes available")

			var targetNode *corev1.Node
			for i := range nodePoolNodes {
				if isNodeReady(nodePoolNodes[i]) {
					targetNode = &nodePoolNodes[i]
					break
				}
			}
			if targetNode == nil {
				targetNode = &nodePoolNodes[0]
			}

			logger.Info("Selected target node", "nodeName", targetNode.Name)

			pod := createPod(currentNamespace, &podConfig{
				namePrefix: "nodepool-test-",
				nodeName:   targetNode.Name,
				image:      "registry.access.redhat.com/ubi8/ubi-minimal:latest",
				command:    []string{"/bin/sh", "-c", "echo 'NodePool test successful' && sleep 5"},
				cpuRequest: "100m",
				memRequest: "128Mi",
			})

			err = client.Create(ctx, pod)
			failIfSTSError(err, "workload creation", "ec2:RunInstances")
			Expect(err).NotTo(HaveOccurred(), "Failed to create test pod")
			defer func() {
				if deleteErr := client.Delete(ctx, pod); deleteErr != nil {
					logger.Error(deleteErr, "Failed to delete test pod", "podName", pod.Name)
				}
			}()

			err = wait.For(conditions.New(client).PodPhaseMatch(pod, corev1.PodSucceeded),
				wait.WithTimeout(3*time.Minute), wait.WithInterval(5*time.Second))
			failIfSTSError(err, "workload scheduling", "ec2:RunInstances")
			Expect(err).NotTo(HaveOccurred(), "Workload failed to complete")
		})

		It("handles invalid configurations gracefully", func(ctx context.Context) {
			logger.Info("Testing invalid configuration handling")
			pod := createPod(currentNamespace, &podConfig{
				namePrefix:   "invalid-selector-test-",
				nodeSelector: map[string]string{nodePoolLabel: "non-existent-nodepool"},
				image:        "registry.access.redhat.com/ubi8/ubi-minimal:latest",
				command:      []string{"/bin/sh", "-c", "sleep 30"},
				cpuRequest:   "50m",
				memRequest:   "64Mi",
			})

			err := client.Create(ctx, pod)
			Expect(err).NotTo(HaveOccurred(), "Should create pod with invalid selector")
			defer func() {
				if deleteErr := client.Delete(ctx, pod); deleteErr != nil {
					logger.Error(deleteErr, "Failed to delete test pod", "podName", pod.Name)
				}
			}()

			err = wait.For(func(ctx context.Context) (bool, error) {
				podList := &corev1.PodList{}
				err := client.WithNamespace(pod.Namespace).List(ctx, podList)
				if err != nil {
					return false, err
				}

				for _, p := range podList.Items {
					if p.Name == pod.Name {
						return p.Status.Phase == corev1.PodPending, nil
					}
				}
				return false, nil
			}, wait.WithTimeout(1*time.Minute), wait.WithInterval(5*time.Second))
			Expect(err).NotTo(HaveOccurred(), "Pod should remain in pending state")
		})

		It("handles resource-intensive workloads", func(ctx context.Context) {
			logger.Info("Testing resource-intensive workloads")
			nodePoolNodes, err := findReadyNodePoolNodes(ctx, client)
			Expect(err).NotTo(HaveOccurred(), "Failed to find ready NodePool nodes")
			Expect(len(nodePoolNodes)).To(BeNumerically(">", 0), "No ready NodePool nodes available")

			var targetNode *corev1.Node
			for i := range nodePoolNodes {
				if isNodeReady(nodePoolNodes[i]) {
					targetNode = &nodePoolNodes[i]
					break
				}
			}
			if targetNode == nil {
				targetNode = &nodePoolNodes[0]
			}

			pod := createPod(currentNamespace, &podConfig{
				namePrefix: "resource-intensive-test-",
				nodeName:   targetNode.Name,
				image:      "registry.access.redhat.com/ubi8/ubi-minimal:latest",
				command:    []string{"/bin/sh", "-c", "echo 'Resource test' && sleep 10"},
				cpuRequest: "500m",
				memRequest: "512Mi",
			})

			err = client.Create(ctx, pod)
			Expect(err).NotTo(HaveOccurred(), "Failed to create resource-intensive pod")
			defer func() {
				if deleteErr := client.Delete(ctx, pod); deleteErr != nil {
					logger.Error(deleteErr, "Failed to delete test pod", "podName", pod.Name)
				}
			}()

			err = wait.For(conditions.New(client).PodPhaseMatch(pod, corev1.PodSucceeded),
				wait.WithTimeout(5*time.Minute), wait.WithInterval(5*time.Second))
			Expect(err).NotTo(HaveOccurred(), "Resource-intensive workload failed")
		})
	})

	Describe("NodePool API Management", func() {
		BeforeAll(func() {
			if !canAccessAPI {
				Skip("NodePool API not accessible from guest cluster")
			}

			logger.Info("Creating NodePool for API tests", "nodePoolName", testNodePoolName)
			nodePoolSpec := buildNodePoolSpec(testNodePoolName, clusterNamespace)
			nodePoolObj := &unstructured.Unstructured{Object: nodePoolSpec}

			_, err := dynamicClient.Resource(nodePoolGVR).Namespace(clusterNamespace).
				Create(context.Background(), nodePoolObj, metav1.CreateOptions{})

			failIfSTSError(err, "NodePool creation", "ec2:RunInstances,ec2:CreateTags")
			Expect(err).NotTo(HaveOccurred(), "NodePool creation failed")
		})

		AfterAll(func() {
			logger.Info("Cleaning up NodePool", "nodePoolName", testNodePoolName)
			err := dynamicClient.Resource(nodePoolGVR).Namespace(clusterNamespace).
				Delete(context.Background(), testNodePoolName, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				logger.Error(err, "Failed to cleanup test NodePool", "nodePoolName", testNodePoolName)
			}

			wait.For(func(ctx context.Context) (bool, error) {
				_, err := dynamicClient.Resource(nodePoolGVR).Namespace(clusterNamespace).
					Get(ctx, testNodePoolName, metav1.GetOptions{})
				return apierrors.IsNotFound(err), nil
			}, wait.WithTimeout(3*time.Minute), wait.WithInterval(10*time.Second))
		})

		It("creates NodePool successfully", func(ctx context.Context) {
			logger.Info("Verifying NodePool creation", "nodePoolName", testNodePoolName)
			_, err := dynamicClient.Resource(nodePoolGVR).Namespace(clusterNamespace).
				Get(ctx, testNodePoolName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "NodePool should exist after creation")
		})

		It("provisions nodes from new NodePool", func(ctx context.Context) {
			logger.Info("Waiting for NodePool to provision nodes", "nodePoolName", testNodePoolName)
			err := waitForNodePoolReady(ctx, dynamicClient, clusterNamespace, testNodePoolName, 10*time.Minute)
			failIfSTSError(err, "node provisioning", "ec2:RunInstances")
			Expect(err).NotTo(HaveOccurred(), "NodePool failed to provision nodes")

			provisionedNodes, err := findNodesForNodePool(ctx, client, testNodePoolName)
			Expect(err).NotTo(HaveOccurred(), "Failed to find provisioned nodes")
			Expect(len(provisionedNodes)).To(BeNumerically(">", 0), "No nodes were provisioned")

			logger.Info("Validating provisioned nodes", "nodeCount", len(provisionedNodes))
			for _, node := range provisionedNodes {
				validateAWSIntegration(node)
			}
		})

		It("rejects duplicate NodePool creation", func(ctx context.Context) {
			logger.Info("Testing duplicate NodePool creation", "nodePoolName", testNodePoolName)
			duplicateSpec := buildNodePoolSpec(testNodePoolName, clusterNamespace)
			duplicateObj := &unstructured.Unstructured{Object: duplicateSpec}

			_, err := dynamicClient.Resource(nodePoolGVR).Namespace(clusterNamespace).
				Create(ctx, duplicateObj, metav1.CreateOptions{})

			Expect(err).To(HaveOccurred(), "Should fail creating duplicate NodePool")
			Expect(apierrors.IsAlreadyExists(err)).To(BeTrue(), "Should return AlreadyExists error")
		})

		It("rejects operations on non-existent NodePools", func(ctx context.Context) {
			logger.Info("Testing non-existent NodePool access")
			_, err := dynamicClient.Resource(nodePoolGVR).Namespace(clusterNamespace).
				Get(ctx, "non-existent-nodepool", metav1.GetOptions{})

			Expect(err).To(HaveOccurred(), "Should fail accessing non-existent NodePool")
			Expect(apierrors.IsNotFound(err)).To(BeTrue(), "Should return NotFound error")
		})
	})
})

type podConfig struct {
	namePrefix   string
	nodeName     string
	nodeSelector map[string]string
	image        string
	command      []string
	cpuRequest   string
	memRequest   string
}

func createPod(namespace string, config *podConfig) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: config.namePrefix,
			Namespace:    namespace,
		},
		Spec: corev1.PodSpec{
			NodeName:     config.nodeName,
			NodeSelector: config.nodeSelector,
			Containers: []corev1.Container{{
				Name:    "test",
				Image:   config.image,
				Command: config.command,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(config.cpuRequest),
						corev1.ResourceMemory: resource.MustParse(config.memRequest),
					},
				},
			}},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	return pod
}

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
	Expect(node.Spec.ProviderID).To(HavePrefix(fmt.Sprintf("%s://", awsProvider)),
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

func getClusterNamespace() (string, error) {
	// Try environment variable first
	if clusterName := os.Getenv("CLUSTER_NAME"); clusterName != "" {
		return fmt.Sprintf("clusters-%s", clusterName), nil
	}

	// Try cluster ID
	if clusterID := os.Getenv("CLUSTER_ID"); clusterID != "" {
		return fmt.Sprintf("clusters-%s", clusterID), nil
	}

	return "", fmt.Errorf("could not determine cluster namespace - set CLUSTER_NAME or CLUSTER_ID environment variable")
}

func canAccessNodePoolAPI(client dynamic.Interface, namespace string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Resource(nodePoolGVR).Namespace(namespace).List(ctx, metav1.ListOptions{Limit: 1})
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
				awsProvider: map[string]interface{}{
					"instanceType": "m5.large",
				},
			},
		},
	}
}

func waitForNodePoolReady(ctx context.Context, client dynamic.Interface, namespace, nodePoolName string, timeout time.Duration) error {
	return wait.For(func(ctx context.Context) (bool, error) {
		np, err := client.Resource(nodePoolGVR).Namespace(namespace).
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

func failIfSTSError(err error, operation, permissions string) {
	if err == nil {
		return
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
			Fail(fmt.Sprintf(
				"STS_PERMISSION_ERROR: %s failed due to missing AWS permissions (%s): %v. This blocks release until STS policies are updated.",
				operation, permissions, err))
		}
	}
}

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}