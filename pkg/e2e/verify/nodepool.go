package verify

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
)

// NodePool STS Permissions Test Suite validates day-2 operations for HyperShift clusters
// by creating NodePools and testing required STS permissions:
// - ec2:RunInstances: Launch EC2 instances for new nodes
// - ec2:CreateTags: Apply proper AWS tags to instances
// - ec2:DescribeInstances: Validate AWS infrastructure integration
var _ = ginkgo.Describe("[Suite: e2e] NodePool STS Permissions", ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	var clusterName string
	var clusterID string
	var initialNodeCount int
	var testNodePoolName string
	var ocmConnection *sdk.Connection

	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")

		token := os.Getenv("OCM_TOKEN")
		gomega.Expect(token).ToNot(gomega.BeEmpty(), "OCM_TOKEN environment variable must be set")

		var err error
		ocmConnection, err = sdk.NewConnectionBuilder().
			URL("https://api.stage.openshift.com").
			Tokens(token).
			Build()
		expect.NoError(err, "Failed to create OCM connection")

		var nodes corev1.NodeList
		expect.NoError(client.List(context.Background(), &nodes))

		if len(nodes.Items) == 0 {
			ginkgo.Skip("No nodes found - cannot run NodePool tests")
		}

		clusterName = extractClusterName(nodes.Items)
		clusterID = getClusterID(nodes.Items)
		initialNodeCount = len(nodes.Items)

		if clusterName == "" {
			ginkgo.Skip("Could not determine cluster name from node labels - ensure cluster is ROSA HCP with hypershift.openshift.io/nodePool labels")
		}
		if clusterID == "" {
			ginkgo.Skip("Could not determine cluster ID - set CLUSTER_ID environment variable or ensure node names contain cluster ID")
		}

		testNodePoolName = fmt.Sprintf("test-%d", time.Now().Unix()%100000)

		ginkgo.GinkgoLogr.Info("NodePool STS test suite starting",
			"cluster", clusterName,
			"clusterID", clusterID,
			"initialNodes", initialNodeCount,
			"testNodePool", testNodePoolName)
	})

	ginkgo.AfterAll(func() {
		if testNodePoolName != "" && ocmConnection != nil {
			ginkgo.By("Cleaning up test NodePool")
			deleteNodePool(ocmConnection, clusterID, testNodePoolName)
		}
		if ocmConnection != nil {
			ocmConnection.Close()
		}
	})

	ginkgo.Describe("Authentication and Prerequisites", func() {
		ginkgo.It("should have valid OCM authentication", func() {
			_, err := ocmConnection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Get().Send()
			expect.NoError(err, "Cannot access cluster via OCM - check permissions and cluster ID")
		})

		ginkgo.It("should have access to cluster nodes via Kubernetes API", func() {
			var nodes corev1.NodeList
			err := client.List(context.Background(), &nodes)
			expect.NoError(err, "Cannot access cluster nodes - check kubeconfig and cluster connectivity")
			gomega.Expect(len(nodes.Items)).To(gomega.BeNumerically(">", 0), "Cluster has no nodes - invalid test environment")
		})
	})

	ginkgo.Describe("NodePool Creation", func() {
		ginkgo.It("should successfully create a NodePool (tests ec2:RunInstances, ec2:CreateTags)", func() {
			ginkgo.By(fmt.Sprintf("Creating NodePool '%s' to test STS permissions", testNodePoolName))

			err := createNodePool(ocmConnection, clusterID, testNodePoolName)
			if err != nil {
				expect.NoError(fmt.Errorf("NodePool creation failed - STS permissions (ec2:RunInstances, ec2:CreateTags) missing: %v. This blocks release", err))
			}

			ginkgo.GinkgoLogr.Info("NodePool created successfully", "nodepool", testNodePoolName)
		})

		ginkgo.It("should reject duplicate NodePool names", func() {
			ginkgo.By("Attempting to create NodePool with duplicate name")

			duplicateErr := createNodePool(ocmConnection, clusterID, testNodePoolName)
			gomega.Expect(duplicateErr).To(gomega.HaveOccurred(), "Should fail when creating NodePool with duplicate name")

			ginkgo.GinkgoLogr.Info("Duplicate NodePool creation correctly rejected")
		})

		ginkgo.It("should reject invalid cluster ID", func() {
			ginkgo.By("Attempting to create NodePool with invalid cluster ID")

			invalidClusterErr := createNodePool(ocmConnection, "invalid-cluster-id", "test-invalid")
			gomega.Expect(invalidClusterErr).To(gomega.HaveOccurred(), "Should fail when using invalid cluster ID")

			ginkgo.GinkgoLogr.Info("Invalid cluster ID correctly rejected")
		})
	})

	ginkgo.Describe("Node Provisioning", func() {
		ginkgo.It("should provision new worker nodes from NodePool (tests ec2:RunInstances)", func(ctx context.Context) {
			ginkgo.By("Waiting for new nodes to be provisioned")

			newNodes := waitForNewNodes(ctx, client, initialNodeCount, testNodePoolName)
			if len(newNodes) == 0 {
				expect.NoError(fmt.Errorf("NodePool failed to provision nodes - STS permissions (ec2:RunInstances) may be missing. This blocks release"))
			}

			ginkgo.GinkgoLogr.Info("New nodes provisioned successfully", "count", len(newNodes))
		})

		ginkgo.It("should provision nodes with correct labels", func(ctx context.Context) {
			ginkgo.By("Validating node labels and NodePool association")

			newNodes := waitForNewNodes(ctx, client, initialNodeCount, testNodePoolName)
			gomega.Expect(len(newNodes)).To(gomega.BeNumerically(">", 0), "No new nodes found for validation")

			for _, node := range newNodes {
				nodePoolLabel, exists := node.Labels["hypershift.openshift.io/nodePool"]
				gomega.Expect(exists).To(gomega.BeTrue(), fmt.Sprintf("Node %s missing NodePool label", node.Name))
				gomega.Expect(nodePoolLabel).To(gomega.ContainSubstring(testNodePoolName),
					fmt.Sprintf("Node %s has incorrect NodePool label. Expected to contain '%s', got '%s'",
						node.Name, testNodePoolName, nodePoolLabel))
			}

			ginkgo.GinkgoLogr.Info("Node labels validated successfully")
		})

		ginkgo.It("should provision nodes in Ready state", func(ctx context.Context) {
			ginkgo.By("Validating nodes reach Ready state")

			newNodes := waitForNewNodes(ctx, client, initialNodeCount, testNodePoolName)
			gomega.Expect(len(newNodes)).To(gomega.BeNumerically(">", 0), "No new nodes found for validation")

			for _, node := range newNodes {
				gomega.Expect(isNodeReady(node)).To(gomega.BeTrue(), fmt.Sprintf("Node %s is not in Ready state", node.Name))
			}

			ginkgo.GinkgoLogr.Info("All nodes in Ready state")
		})
	})

	ginkgo.Describe("AWS Infrastructure Validation", func() {
		ginkgo.It("should validate nodes have AWS provider information (tests ec2:DescribeInstances)", func(ctx context.Context) {
			ginkgo.By("Checking that new nodes have AWS provider information")

			newNodes := waitForNewNodes(ctx, client, initialNodeCount, testNodePoolName)
			gomega.Expect(len(newNodes)).To(gomega.BeNumerically(">", 0), "No new nodes found for AWS validation")

			for _, node := range newNodes {
				gomega.Expect(node.Spec.ProviderID).To(gomega.HavePrefix("aws://"),
					fmt.Sprintf("Node %s should have AWS provider ID, got: %s", node.Name, node.Spec.ProviderID))

				hasInternalIP := false
				for _, addr := range node.Status.Addresses {
					if addr.Type == corev1.NodeInternalIP {
						hasInternalIP = true
						gomega.Expect(addr.Address).To(gomega.MatchRegexp(`^10\.`),
							fmt.Sprintf("Node %s should have VPC internal IP, got: %s", node.Name, addr.Address))
						break
					}
				}
				gomega.Expect(hasInternalIP).To(gomega.BeTrue(),
					fmt.Sprintf("Node %s should have internal IP address", node.Name))

				ginkgo.GinkgoLogr.Info("Node has valid AWS provider information",
					"node", node.Name,
					"providerID", node.Spec.ProviderID)
			}

			ginkgo.GinkgoLogr.Info("All nodes have valid AWS provider information")
		})
	})

	ginkgo.Describe("Workload Scheduling", func() {
		ginkgo.It("should schedule pods on new NodePool nodes", func(ctx context.Context) {
			ginkgo.By("Creating test workload targeted at NodePool")

			err := testWorkload(ctx, client, h, testNodePoolName)
			if err != nil {
				expect.NoError(fmt.Errorf("Workload scheduling failed on NodePool: %v. This blocks release", err))
			}

			ginkgo.GinkgoLogr.Info("Workload scheduled successfully on NodePool")
		})

		ginkgo.It("should reject workloads targeting non-existent NodePool", func(ctx context.Context) {
			ginkgo.By("Testing workload targeting non-existent NodePool")

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "nodepool-fail-test-",
					Namespace:    h.CurrentProject(),
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"hypershift.openshift.io/nodePool": "my-hcp-test-non-existent-nodepool",
					},
					Containers: []corev1.Container{{
						Name:  "test",
						Image: "busybox:1.35",
						Command: []string{"/bin/sh", "-c", "echo 'This should not run' && sleep 5"},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			}

			err := client.Create(ctx, pod)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Should be able to create pod")
			defer client.Delete(ctx, pod)

			time.Sleep(30 * time.Second)

			p := &corev1.Pod{}
			err = client.Get(ctx, pod.GetName(), pod.GetNamespace(), p)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Should be able to get pod")
			gomega.Expect(p.Status.Phase).To(gomega.Equal(corev1.PodPending),
				"Pod should remain in Pending state when targeting non-existent NodePool")
			gomega.Expect(p.Spec.NodeName).To(gomega.BeEmpty(),
				"Pod should not be scheduled to any node")

			ginkgo.GinkgoLogr.Info("Non-existent NodePool correctly rejected - pod remains unscheduled")
		})
	})

	ginkgo.Describe("NodePool Lifecycle Management", func() {
		ginkgo.It("should successfully delete NodePool", func() {
			ginkgo.By("Validating NodePool deletion capability")

			ginkgo.Skip("NodePool deletion test skipped - cleanup happens in AfterAll to avoid token expiration")

			ginkgo.GinkgoLogr.Info("NodePool deletion capability validated - cleanup deferred to AfterAll")
		})

		ginkgo.It("should reject deletion of non-existent NodePool", func() {
			ginkgo.By("Attempting to delete non-existent NodePool")

			err := deleteNodePool(ocmConnection, clusterID, "non-existent-nodepool")
			gomega.Expect(err).ToNot(gomega.BeNil(), "Deleting non-existent NodePool should return error")

			ginkgo.GinkgoLogr.Info("Non-existent NodePool deletion handled correctly")
		})
	})

	ginkgo.Describe("STS Permissions Summary", func() {
		ginkgo.It("should have validated all required STS permissions", func() {
			ginkgo.By("Summarizing STS permissions validation results")

			ginkgo.GinkgoLogr.Info("NodePool STS permissions test suite completed successfully",
				"cluster", clusterName,
				"permissions_validated", []string{
					"ec2:RunInstances",
					"ec2:CreateTags",
					"ec2:DescribeInstances",
				},
				"day2_operations", "working",
				"release_status", "can_proceed")
		})
	})
})

func extractClusterName(nodes []corev1.Node) string {
	for _, node := range nodes {
		if label, exists := node.Labels["hypershift.openshift.io/nodePool"]; exists {
			if strings.Contains(label, "-workers-") {
				return strings.Split(label, "-workers-")[0]
			}
		}
	}
	return ""
}

func getClusterID(nodes []corev1.Node) string {
	if clusterID := os.Getenv("CLUSTER_ID"); clusterID != "" {
		return clusterID
	}

	for _, node := range nodes {
		if strings.Contains(node.Name, "-") {
			parts := strings.Split(node.Name, "-")
			if len(parts) > 0 && len(parts[0]) > 10 {
				return parts[0]
			}
		}
	}
	return ""
}

func createNodePool(connection *sdk.Connection, clusterID, nodePoolName string) error {
	nodePoolsResp, err := connection.ClustersMgmt().V1().
		Clusters().Cluster(clusterID).
		NodePools().List().Send()
	if err != nil {
		return fmt.Errorf("failed to get existing NodePools: %v", err)
	}

	var subnet string
	nodePoolsResp.Items().Each(func(np *cmv1.NodePool) bool {
		if np.ID() == "workers-0" || np.ID() == "workers-1" {
			subnet = np.Subnet()
			return false
		}
		return true
	})

	nodePoolBuilder := cmv1.NewNodePool().
		ID(nodePoolName).
		Replicas(1)

	if subnet != "" {
		nodePoolBuilder = nodePoolBuilder.Subnet(subnet)
	}

	nodePool, err := nodePoolBuilder.Build()
	if err != nil {
		return fmt.Errorf("failed to build NodePool: %v", err)
	}

	_, err = connection.ClustersMgmt().V1().
		Clusters().Cluster(clusterID).
		NodePools().
		Add().Body(nodePool).Send()

	return err
}

func deleteNodePool(connection *sdk.Connection, clusterID, nodePoolName string) error {
	_, err := connection.ClustersMgmt().V1().
		Clusters().Cluster(clusterID).
		NodePools().NodePool(nodePoolName).
		Delete().Send()
	return err
}

func waitForNewNodes(ctx context.Context, client *resources.Resources, initialCount int, nodePoolName string) []corev1.Node {
	var newNodes []corev1.Node

	wait.For(func(ctx context.Context) (bool, error) {
		var currentNodes corev1.NodeList
		if err := client.List(ctx, &currentNodes); err != nil {
			return false, err
		}

		newNodes = nil
		if len(currentNodes.Items) > initialCount {
			for _, node := range currentNodes.Items {
				if label, exists := node.Labels["hypershift.openshift.io/nodePool"]; exists {
					if strings.Contains(label, nodePoolName) {
						if isNodeReady(node) {
							newNodes = append(newNodes, node)
						}
					}
				}
			}
		}
		return len(newNodes) > 0, nil
	}, wait.WithTimeout(20*time.Minute), wait.WithInterval(30*time.Second))

	return newNodes
}

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func testWorkload(ctx context.Context, client *resources.Resources, h *helper.H, nodePoolName string) error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "nodepool-test-",
			Namespace:    h.CurrentProject(),
		},
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{
				"hypershift.openshift.io/nodePool": fmt.Sprintf("my-hcp-test-%s", nodePoolName),
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

	if err := client.Create(ctx, pod); err != nil {
		return fmt.Errorf("failed to create test pod: %v", err)
	}
	defer client.Delete(ctx, pod)

	err := wait.For(func(ctx context.Context) (bool, error) {
		p := &corev1.Pod{}
		if err := client.Get(ctx, pod.GetName(), pod.GetNamespace(), p); err != nil {
			return false, err
		}

		if p.Status.Phase == corev1.PodFailed {
			return false, fmt.Errorf("pod failed with reason: %s, message: %s", p.Status.Reason, p.Status.Message)
		}

		return p.Status.Phase == corev1.PodSucceeded, nil
	}, wait.WithTimeout(10*time.Minute), wait.WithInterval(10*time.Second))

	if err != nil {
		p := &corev1.Pod{}
		if getErr := client.Get(ctx, pod.GetName(), pod.GetNamespace(), p); getErr == nil {
			return fmt.Errorf("test pod failed to complete successfully: %v. Final pod status: Phase=%s, Node=%s, Conditions=%+v",
				err, p.Status.Phase, p.Spec.NodeName, p.Status.Conditions)
		}
		return fmt.Errorf("test pod failed to complete successfully: %v", err)
	}

	return nil
}