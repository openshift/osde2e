package machinepools

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	machinepoolcreate "github.com/openshift/rosa/cmd/create/machinepool"
	machinepooldelete "github.com/openshift/rosa/cmd/dlt/machinepool"
	machinepooledit "github.com/openshift/rosa/cmd/edit/machinepool"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
)

var _ = ginkgo.Describe("ROSA Machine Pools", ginkgo.Ordered, label.ROSA, label.HyperShift, func() {
	var (
		h               *helper.H
		client          *resources.Resources
		clusterName     string
		machinePoolName string
	)

	ginkgo.BeforeAll(func() {
		if !viper.GetBool(config.Hypershift) {
			ginkgo.Skip("test is currently not supported on non-HyperShift (SDCICD-925)")
		}

		h = helper.New()
		client = h.AsUser("")

		// TODO: can i get the cluster name from the cluster?
		clusterName = viper.GetString(config.Cluster.Name)
		Expect(clusterName).ShouldNot(BeEmpty())

		machinePoolName = clusterName
	})

	nodesScaledTo := func(ctx context.Context, count int) func() (bool, error) {
		// TODO: what is the label on non-hypershift?
		lbls := labels.FormatLabels(map[string]string{"hypershift.openshift.io/nodePool": clusterName})
		return func() (bool, error) {
			var nodes v1.NodeList
			if err := client.List(ctx, &nodes, resources.WithLabelSelector(lbls)); err != nil {
				return false, fmt.Errorf("error listing nodes: %v", err)
			}
			for _, node := range nodes.Items {
				for _, condition := range node.Status.Conditions {
					if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
						return false, nil
					}
				}
			}
			return len(nodes.Items) == count, nil
		}
	}

	ginkgo.It("can be created", func(ctx context.Context) {
		replicaCount := 3
		cmd := machinepoolcreate.Cmd
		cmd.SetArgs([]string{fmt.Sprintf("--name=%s", machinePoolName), fmt.Sprintf("--cluster=%s", clusterName), fmt.Sprintf("--replicas=%d", replicaCount)})
		expect.NoError(cmd.Execute(), "failed to create machinepool")
		expect.NoError(wait.For(nodesScaledTo(ctx, replicaCount), wait.WithTimeout(10*time.Minute)), "nodes never scaled up")
	})

	ginkgo.It("can be scaled", func(ctx context.Context) {
		replicaCount := 1
		cmd := machinepooledit.Cmd
		cmd.SetArgs([]string{fmt.Sprintf("--replicas=%d", replicaCount), fmt.Sprintf("--cluster=%s", clusterName), machinePoolName})
		expect.NoError(cmd.Execute(), "failed to edit machinepool")
		expect.NoError(wait.For(nodesScaledTo(ctx, replicaCount), wait.WithTimeout(10*time.Minute)), "machinepool never scaled down")
	})

	ginkgo.It("can be deleted", func(ctx context.Context) {
		cmd := machinepooldelete.Cmd
		cmd.SetArgs([]string{"--yes", fmt.Sprintf("--cluster=%s", clusterName), machinePoolName})
		expect.NoError(cmd.Execute(), "failed to delete machinepool")
		expect.NoError(wait.For(nodesScaledTo(ctx, 0), wait.WithTimeout(5*time.Minute)), "nodes weren't deleted")
	})
})
