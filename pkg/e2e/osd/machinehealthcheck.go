package osd

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	machineV1beta1 "github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	machineAPINamespace = "openshift-machine-api"
)

var machineHealthTestName string = "[Suite: e2e] MachineHealthChecks"

func init() {
	alert.RegisterGinkgoAlert(machineHealthTestName, "SD-SRE", "Alex Chvatal", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(machineHealthTestName, func() {
	h := helper.New()

	ginkgo.It("infra MHC should exist", func() {
		mhc, err := h.Machine().MachineV1beta1().MachineHealthChecks(machineAPINamespace).Get(context.TODO(), "srep-infra-healthcheck", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// verify there's an MHC for infra nodes
		Expect(mhc.Spec.Selector.MatchLabels["machine.openshift.io/cluster-api-machine-role"]).To(Equal("infra"))

		// verify the unhealthy conditions are on all nodes
		Expect(mhc.Spec.UnhealthyConditions).To(SatisfyAll(
			ContainElement(
				machineV1beta1.UnhealthyCondition{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionFalse,
					Timeout: "480s",
				},
			),
			ContainElement(
				machineV1beta1.UnhealthyCondition{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionUnknown,
					Timeout: "480s",
				},
			),
		))
	})

	ginkgo.It("worker MHC should exist", func() {
		mhc, err := h.Machine().MachineV1beta1().MachineHealthChecks(machineAPINamespace).Get(context.TODO(), "srep-worker-healthcheck", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// verify there's an MHC for worker nodes
		Expect(mhc.Spec.Selector.MatchLabels["machine.openshift.io/cluster-api-machine-role"]).To(Equal("worker"))

		// verify the unhealthy conditions are on all nodes
		Expect(mhc.Spec.UnhealthyConditions).To(SatisfyAll(
			ContainElement(
				machineV1beta1.UnhealthyCondition{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionFalse,
					Timeout: "480s",
				},
			),
			ContainElement(
				machineV1beta1.UnhealthyCondition{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionUnknown,
					Timeout: "480s",
				},
			),
		))
	})

	ginkgo.It("should replace unhealthy nodes", func() {
		r := h.Runner("chroot /host -- systemctl stop kubelet")
		r.Name = "stop-kubelet"
		// i can't believe SecurityContext.Privileged is a pointer to a bool
		truePointer := true
		r.PodSpec.Containers[0].SecurityContext.Privileged = &truePointer

		// get original list of machines to compare against later
		originalMachines, err := h.Machine().MachineV1beta1().Machines(machineAPINamespace).List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		// modify the MHC to have a very short unhealthy time
		mhc, err := h.Machine().MachineV1beta1().MachineHealthChecks(machineAPINamespace).Get(context.TODO(), "srep-worker-healthcheck", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		h.Machine().MachineV1beta1().MachineHealthChecks(machineAPINamespace).Patch(
			context.TODO(),
			mhc.ObjectMeta.Name,
			types.MergePatchType,
			[]byte("{'spec':{'unhealthyConditions[0]':{'timeout':10}}}"),
			metav1.PatchOptions{},
		)

		// execute the runner
		stopCh := make(chan struct{})
		err = r.Run(30, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// wait and confirm that there's a new machine
		newMachines, err := h.Machine().MachineV1beta1().Machines(machineAPINamespace).List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(originalMachines).NotTo(Equal(newMachines))
	})
})
