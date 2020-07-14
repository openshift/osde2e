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
	OperatorNamespace = "openshift-machine-api"
)

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: informing] MachineHealthChecks",
		TeamOwner:        "SD-SRE",
		PrimaryContact:   "Alex Chvatal",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 1,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name, func() {
	h := helper.New()

	ginkgo.It("should exist", func() {
		machineHealthChecks, err := h.Machine().MachineV1beta1().MachineHealthChecks(OperatorNamespace).List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		for _, mhc := range machineHealthChecks.Items {
			// verify there's an MHC for infra and worker nodes
			Expect(mhc.Spec.Selector.MatchLabels["machine.openshift.io/cluster-api-machine-role"]).To(SatisfyAny(
				Equal("infra"),
				Equal("worker"),
			))

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
		}
	})

	ginkgo.It("should replace unhealthy nodes", func() {
		r := h.Runner("chroot /host -- systemctl stop kubelet")
		r.Name = "stop-kubelet"
		// i can't believe SecurityContext.Privileged is a pointer to a bool
		truePointer := true
		r.PodSpec.Containers[0].SecurityContext.Privileged = &truePointer

		// get original list of machines to compare against later
		originalMachines, err := h.Machine().MachineV1beta1().Machines(OperatorNamespace).List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		// modify the MHC to have a very short unhealthy time
		machineHealthChecks, err := h.Machine().MachineV1beta1().MachineHealthChecks(OperatorNamespace).List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, m := range machineHealthChecks.Items {
			h.Machine().MachineV1beta1().MachineHealthChecks(OperatorNamespace).Patch(
				context.TODO(),
				m.ObjectMeta.Name,
				types.MergePatchType,
				[]byte("{'spec':{'unhealthyConditions[0]':{'timeout':10}}}"),
				metav1.PatchOptions{},
			)
		}

		// execute the runner
		stopCh := make(chan struct{})
		err = r.Run(30, stopCh)
		Expect(err).NotTo(HaveOccurred())

		// wait and confirm that there's a new machine
		newMachines, err := h.Machine().MachineV1beta1().Machines(OperatorNamespace).List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(originalMachines).NotTo(Equal(newMachines))
	})
})
