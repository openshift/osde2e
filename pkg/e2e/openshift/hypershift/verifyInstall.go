// Package hypershift runs test units for hypershift clusters
package hypershift

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/util"
)

var (
	hypershiftInstallVerify = "[Suite: hypershift]"
	workerNodes             []string
)

func init() {
	alert.RegisterGinkgoAlert(hypershiftInstallVerify, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

// Checks the installation of the hypershift worker nodes in CCS AWS account
var _ = ginkgo.Describe(hypershiftInstallVerify, func() {
	defer ginkgo.GinkgoRecover()
	ginkgo.Context("Verify Hypershift worker node in CCS Account", func() {
		util.GinkgoIt("Worker nodes are available in AWS", func(ctx context.Context) {
			ginkgo.By("Getting the list of worker nodes from the cluster")
			err := getWorkerNodesInCluster(ctx)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("error getting worker nodes in cluster: %v", err))
			Expect(len(workerNodes)).To(BeNumerically(">", 0), "No worker nodes found in the cluster")

			ginkgo.By("Checking if the worker nodes are present in AWS")
			err = checkWorkerNodesInAWS()
			Expect(err).NotTo(HaveOccurred(), "Error checking if worker nodes are present in AWS")

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})

// checkWorkerNodesInCluster returns a list of nodes in the cluster
func getWorkerNodesInCluster(ctx context.Context) error {
	restConfig, _, err := cluster.ClusterConfig(viper.GetString(config.Cluster.ID))
	if err != nil {
		return fmt.Errorf("error getting cluster config: %v", err)
	}

	// Create a clientset to interact with the cluster
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("error creating clientset: %v", err)
	}

	// Get the list of nodes in the cluster
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting nodes: %v", err)
	}

	// Add nodes to slice of worker nodes
	for _, node := range nodes.Items {
		workerNodes = append(workerNodes, node.Name)
	}

	return nil
}

func checkWorkerNodesInAWS() error {
	for _, node := range workerNodes {
		exists, err := aws.CcsAwsSession.CheckIfEC2ExistBasedOnNodeName(node)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("worker node %s not found in AWS", node)
		}
	}

	return nil
}
