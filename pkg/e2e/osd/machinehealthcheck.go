package osd

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	machineV1beta1 "github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	machineAPINamespace = "openshift-machine-api"
)

var machineHealthTestName string = "[Suite: e2e] MachineHealthChecks"

func init() {
	alert.RegisterGinkgoAlert(
		machineHealthTestName,
		"SD-SRE",
		"Alex Chvatal",
		"sd-cicd-alerts",
		"sd-cicd@redhat.com",
		4,
	)
}

var _ = ginkgo.Describe(machineHealthTestName, func() {
	h := helper.New()

	util.GinkgoIt("infra MHC should exist", func(ctx context.Context) {
		mhc, err := h.Machine().
			MachineV1beta1().
			MachineHealthChecks(machineAPINamespace).
			Get(ctx, "srep-infra-healthcheck", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// verify there's an MHC for infra nodes
		Expect(mhc.Spec.Selector.MatchExpressions).To(ConsistOf(
			metav1.LabelSelectorRequirement{
				Key:      "machine.openshift.io/cluster-api-machine-role",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"infra"},
			},
			metav1.LabelSelectorRequirement{
				Key:      "machine.openshift.io/cluster-api-machineset",
				Operator: metav1.LabelSelectorOpExists,
			},
		))

		// verify the unhealthy conditions are on all nodes
		Expect(mhc.Spec.UnhealthyConditions).To(ConsistOf(
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionFalse,
				Timeout: "480s",
			},
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionUnknown,
				Timeout: "480s",
			},
		))
	}, float64(500))

	util.GinkgoIt("worker MHC should exist", func(ctx context.Context) {
		mhc, err := h.Machine().
			MachineV1beta1().
			MachineHealthChecks(machineAPINamespace).
			Get(ctx, "srep-worker-healthcheck", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// verify there's an MHC for worker nodes
		Expect(mhc.Spec.Selector.MatchExpressions).To(ConsistOf(
			metav1.LabelSelectorRequirement{
				Key:      "machine.openshift.io/cluster-api-machine-role",
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   []string{"infra", "master"},
			},
			metav1.LabelSelectorRequirement{
				Key:      "machine.openshift.io/cluster-api-machineset",
				Operator: metav1.LabelSelectorOpExists,
			},
		))

		// verify the unhealthy conditions are on all nodes
		Expect(mhc.Spec.UnhealthyConditions).To(ConsistOf(
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionFalse,
				Timeout: "480s",
			},
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionUnknown,
				Timeout: "480s",
			},
		))
	}, float64(500))
})
