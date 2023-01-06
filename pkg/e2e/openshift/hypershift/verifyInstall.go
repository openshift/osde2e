// Package hypershift runs test units for hypershift clusters
package hypershift

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

var (
	suiteName = "HyperShift"
	client    *resources.Resources
)

func init() {
	alert.RegisterGinkgoAlert(suiteName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

// Checks the installation of the hypershift worker nodes in CCS AWS account
var _ = ginkgo.Describe(suiteName, ginkgo.Ordered, label.HyperShift, func() {
	h := helper.New()
	client = h.AsUser("")

	ginkgo.It("worker nodes are available in aws ccs account", func(ctx context.Context) {
		ginkgo.By("getting the list of worker nodes from the cluster")
		workerNodes, err := getWorkerNodesInCluster(ctx)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("error getting worker nodes in cluster: %v", err))
		Expect(len(*workerNodes)).To(BeNumerically(">", 0), "no worker nodes found in cluster")

		ginkgo.By("checking if the worker nodes are present in aws")
		err = checkWorkerNodesInAWS(workerNodes)
		Expect(err).NotTo(HaveOccurred(), "Error checking if worker nodes are present in AWS")
	})
})

// checkWorkerNodesInCluster returns a list of nodes in the cluster
func getWorkerNodesInCluster(ctx context.Context) (*[]string, error) {
	// Get the list of nodes in the cluster
	var nodes v1.NodeList
	err := client.List(ctx, &nodes)
	if err != nil {
		return nil, fmt.Errorf("error getting nodes: %v", err)
	}

	// Add nodes to slice of worker nodes
	var workerNodes []string
	for _, node := range nodes.Items {
		workerNodes = append(workerNodes, node.Name)
	}

	return &workerNodes, nil
}

func checkWorkerNodesInAWS(workerNodes *[]string) error {
	for _, node := range *workerNodes {
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
