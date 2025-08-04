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

var _ = ginkgo.Describe("[Suite: e2e] NodePool STS Permissions", ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	var clusterNamespace string
	var testNodePoolName string
	var initialNodeCount int

	nodePoolGVR := schema.GroupVersionResource{
		Group:    "hypershift.openshift.io",
		Version:  "v1beta1",
		Resource: "nodepools",
	}

	ginkgo.BeforeAll(func() {
		if !viper.GetBool(config.Hypershift) {
			ginkgo.Skip("NodePool tests are only supported on HyperShift clusters")
		}

		h = helper.New()
		client = h.AsUser("")

		var nodeList corev1.NodeList
		expect.NoError(client.List(context.Background(), &nodeList))

		if len(nodeList.Items) == 0 {
			ginkgo.Skip("No nodes found - cannot run NodePool tests")
		}

		initialNodeCount = len(nodeList.Items)

		for _, node := range nodeList.Items {
			if label, exists := node.Labels["hypershift.openshift.io/nodePool"]; exists {
				parts := strings.Split(label, "-workers-")
				if len(parts) >= 1 {
					clusterNamespace = parts[0]
					break
				}
			}
		}

		if clusterNamespace == "" {
			ginkgo.Skip("Could not determine cluster namespace from node labels")
		}

		testNodePoolName = fmt.Sprintf("test-%d", time.Now().Unix()%100000)
	})

	ginkgo.AfterAll(func(ctx context.Context) {
		if testNodePoolName != "" {
			ginkgo.By("Cleaning up test NodePool")
			err := h.Dynamic().Resource(nodePoolGVR).Namespace(clusterNamespace).
				Delete(ctx, testNodePoolName, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				ginkgo.GinkgoLogr.Error(err, "Failed to cleanup test NodePool", "name", testNodePoolName)
			}

			ginkgo.By("Cleaning up test pods")
			podList := &corev1.PodList{}
			err = client.WithNamespace(h.CurrentProject()).List(ctx, podList)
			if err == nil {
				for _, pod := range podList.Items {
					if strings.HasPrefix(pod.Name, "nodepool-test-") {
						client.Delete(ctx, &pod)
					}
				}
			}
		}
	})

	ginkgo.It("should successfully create NodePool", func(ctx context.Context) {
		ginkgo.By("Getting existing NodePool configuration")

		existingNodePools, err := h.Dynamic().Resource(nodePoolGVR).Namespace(clusterNamespace).List(ctx, metav1.ListOptions{})
		expect.NoError(err, "Failed to list existing NodePools")
		Expect(len(existingNodePools.Items)).To(BeNumerically(">", 0), "No existing NodePools found to reference")

		ginkgo.By("Creating test NodePool to validate STS permissions")

		refNodePool := existingNodePools.Items[0]
		var subnet string
		if spec, found, err := unstructured.NestedMap(refNodePool.Object, "spec"); found && err == nil {
			if platform, found, err := unstructured.NestedMap(spec, "platform"); found && err == nil {
				if aws, found, err := unstructured.NestedMap(platform, "aws"); found && err == nil {
					if s, found, err := unstructured.NestedString(aws, "subnet"); found && err == nil {
						subnet = s
					}
				}
			}
		}

		nodePoolSpec := map[string]interface{}{
			"apiVersion": "hypershift.openshift.io/v1beta1",
			"kind":       "NodePool",
			"metadata": map[string]interface{}{
				"name":      testNodePoolName,
				"namespace": clusterNamespace,
			},
			"spec": map[string]interface{}{
				"clusterName": clusterNamespace,
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

		if subnet != "" {
			spec := nodePoolSpec["spec"].(map[string]interface{})
			platform := spec["platform"].(map[string]interface{})
			aws := platform["aws"].(map[string]interface{})
			aws["subnet"] = subnet
		}

		nodePoolObj := &unstructured.Unstructured{Object: nodePoolSpec}
		_, err = h.Dynamic().Resource(nodePoolGVR).Namespace(clusterNamespace).Create(ctx, nodePoolObj, metav1.CreateOptions{})
		expect.NoError(err, "NodePool creation failed - STS permissions (ec2:RunInstances, ec2:CreateTags) missing")
	})

	ginkgo.It("should provision nodes with correct labels", func(ctx context.Context) {
		ginkgo.By("Waiting for new nodes to be provisioned")

		var newNodes []corev1.Node
		err := wait.For(func(ctx context.Context) (bool, error) {
			var nodeList corev1.NodeList
			err := client.List(ctx, &nodeList)
			if err != nil {
				return false, err
			}

			newNodes = nil
			if len(nodeList.Items) > initialNodeCount {
				for _, node := range nodeList.Items {
					if label, exists := node.Labels["hypershift.openshift.io/nodePool"]; exists {
						if strings.Contains(label, testNodePoolName) && isNodeReady(node) {
							newNodes = append(newNodes, node)
						}
					}
				}
			}
			return len(newNodes) > 0, nil
		}, wait.WithTimeout(20*time.Minute), wait.WithInterval(30*time.Second))

		expect.NoError(err, "NodePool failed to provision nodes - STS permissions (ec2:RunInstances) may be missing")
		Expect(len(newNodes)).To(BeNumerically(">", 0), "No new nodes found")

		ginkgo.By("Validating node has proper AWS integration")

		for _, node := range newNodes {
			nodePoolLabel, exists := node.Labels["hypershift.openshift.io/nodePool"]
			Expect(exists).To(BeTrue(), "Node %s missing NodePool label", node.Name)
			Expect(nodePoolLabel).To(ContainSubstring(testNodePoolName),
				"Node %s has incorrect NodePool label", node.Name)

			Expect(node.Spec.ProviderID).To(HavePrefix("aws://"),
				"Node %s should have AWS provider ID - ec2:DescribeInstances permission may be missing", node.Name)

			hasInternalIP := false
			for _, addr := range node.Status.Addresses {
				if addr.Type == corev1.NodeInternalIP {
					hasInternalIP = true
					Expect(addr.Address).To(MatchRegexp(`^10\.`),
						"Node %s should have VPC internal IP", node.Name)
					break
				}
			}
			Expect(hasInternalIP).To(BeTrue(), "Node %s should have internal IP", node.Name)
		}
	})

	ginkgo.It("should schedule workloads on new NodePool nodes", func(ctx context.Context) {
		ginkgo.By("Creating test workload targeted at NodePool")

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "nodepool-test-",
				Namespace:    h.CurrentProject(),
			},
			Spec: corev1.PodSpec{
				NodeSelector: map[string]string{
					"hypershift.openshift.io/nodePool": fmt.Sprintf("%s-%s", clusterNamespace, testNodePoolName),
				},
				Containers: []corev1.Container{{
					Name:    "test",
					Image:   "registry.access.redhat.com/ubi8/ubi-minimal",
					Command: []string{"/bin/sh", "-c", "echo 'NodePool workload test successful' && sleep 5"},
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

		expect.NoError(client.Create(ctx, pod), "Failed to create test pod")

		ginkgo.By("Waiting for workload to complete successfully")

		err := wait.For(conditions.New(client).PodPhaseMatch(pod, corev1.PodSucceeded), wait.WithTimeout(5*time.Minute))
		expect.NoError(err, "Workload scheduling failed on NodePool")

		expect.NoError(client.Delete(ctx, pod), "Failed to delete test pod")
	})

	ginkgo.It("should reject duplicate NodePool names", func(ctx context.Context) {
		ginkgo.By("Testing duplicate NodePool creation")

		duplicateNodePoolSpec := map[string]interface{}{
			"apiVersion": "hypershift.openshift.io/v1beta1",
			"kind":       "NodePool",
			"metadata": map[string]interface{}{
				"name":      testNodePoolName,
				"namespace": clusterNamespace,
			},
			"spec": map[string]interface{}{
				"clusterName": clusterNamespace,
				"replicas":    1,
				"platform": map[string]interface{}{
					"aws": map[string]interface{}{
						"instanceType": "m5.large",
					},
				},
			},
		}

		duplicateNodePoolObj := &unstructured.Unstructured{Object: duplicateNodePoolSpec}
		_, err := h.Dynamic().Resource(nodePoolGVR).Namespace(clusterNamespace).Create(ctx, duplicateNodePoolObj, metav1.CreateOptions{})
		Expect(err).To(HaveOccurred(), "Should fail when creating NodePool with duplicate name")
	})

	ginkgo.It("should reject operations on non-existent NodePool", func(ctx context.Context) {
		ginkgo.By("Testing access to non-existent NodePool")

		testNodePool := &unstructured.Unstructured{}
		testNodePool.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   nodePoolGVR.Group,
			Version: nodePoolGVR.Version,
			Kind:    "NodePool",
		})

		_, err := h.Dynamic().Resource(nodePoolGVR).Namespace(clusterNamespace).
			Get(ctx, "non-existent-nodepool", metav1.GetOptions{})

		Expect(err).To(HaveOccurred(), "Getting non-existent NodePool should fail")
		Expect(apierrors.IsNotFound(err)).To(BeTrue(), "Should return NotFound error")
	})
})

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}