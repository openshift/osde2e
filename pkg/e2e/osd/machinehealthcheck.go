package osd

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	machineV1beta1 "github.com/openshift/api/machine/v1beta1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	machineAPINamespace = "openshift-machine-api"
)

var machineHealthTestName string = "[Suite: e2e] MachineHealthChecks"
var metalInstanceTypes = []string{"m5.metal", "m5d.metal", "m5n.metal", "m5dn.metal", "m5zn.metal", "m6a.metal", "m6i.metal", "m6id.metal", "r5.metal", "r5d.metal", "r5n.metal", "r5dn.metal", "r6a.metal", "r6i.metal", "r6id.metal", "x2iezn.metal", "z1d.metal", "c5.metal", "c5d.metal", "c5n.metal", "c6a.metal", "c6i.metal", "c6id.metal", "i3.metal", "i3en.metal"}

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

var _ = ginkgo.Describe(machineHealthTestName, label.E2E, func() {
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
				Timeout: metav1.Duration{Duration: 480 * time.Second},
			},
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionUnknown,
				Timeout: metav1.Duration{Duration: 480 * time.Second},
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
			metav1.LabelSelectorRequirement{
				Key:      "machine.openshift.io/instance-type",
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   metalInstanceTypes,
			},
		))

		// verify the unhealthy conditions are on all nodes
		Expect(mhc.Spec.UnhealthyConditions).To(ConsistOf(
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionFalse,
				Timeout: metav1.Duration{Duration: 480 * time.Second},
			},
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionUnknown,
				Timeout: metav1.Duration{Duration: 480 * time.Second},
			},
		))
	}, float64(500))

	util.GinkgoIt("metal worker MHC should exist", func(ctx context.Context) {
		mhc, err := h.Machine().
			MachineV1beta1().
			MachineHealthChecks(machineAPINamespace).
			Get(ctx, "srep-metal-worker-healthcheck", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// verify there's an MHC for metal worker nodes
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
			metav1.LabelSelectorRequirement{
				Key:      "machine.openshift.io/instance-type",
				Operator: metav1.LabelSelectorOpIn,
				Values:   metalInstanceTypes,
			},
		))

		// verify the unhealthy conditions are on all nodes
		Expect(mhc.Spec.UnhealthyConditions).To(ConsistOf(
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionFalse,
				Timeout: metav1.Duration{Duration: 8 * time.Minute},
			},
			machineV1beta1.UnhealthyCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionUnknown,
				Timeout: metav1.Duration{Duration: 15 * time.Minute},
			},
		))
	}, float64(500))
})
