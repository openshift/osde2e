// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
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

// Checks the installation of the hypershift cluster
var _ = ginkgo.Describe(hypershiftInstallVerify, func() {
	ginkgo.Context("Verify install using oc", func() {
		util.GinkgoIt("Worker nodes are ready", func(ctx context.Context) {
			Expect(checkWorkerNodesInCluster()).ToNot(BeEmpty())
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})

	ginkgo.Context("Verify install using AWS", func() {
		util.GinkgoIt("Worker nodes are present in CCS account", func(ctx context.Context) {
			Expect(checkWorkerNodesInAWS()).To(BeTrue())
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})
})

// checkWorkerNodesInCluster returns a list of nodes in the cluster
func checkWorkerNodesInCluster() ([]string, error) {

	restConfig, _, err := cluster.ClusterConfig(viper.GetString(config.Cluster.ID))
	if err != nil {
		return nil, fmt.Errorf("error getting cluster config: %v", err)
	}

	// Create a clientset to interact with the cluster
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %v", err)
	}

	// Get the list of nodes in the cluster
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting nodes: %v", err)
	}

	// add the worker nodes to the list
	for _, node := range nodes.Items {
		workerNodes = append(workerNodes, node.Name)
	}

	return workerNodes, nil
}

func checkWorkerNodesInAWS() (bool, error) {
	for _, node := range workerNodes {
		exists, err := aws.CcsAwsSession.CheckIfEC2ExistBasedOnNodeName(node)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}

	return true, nil
}
