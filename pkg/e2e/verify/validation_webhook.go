package verify

import (
	"context"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: e2e] Validation Webhook",
		TeamOwner:        "SD-SREP",
		PrimaryContact:   "Matt Bargenquast",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 4,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name, func() {

	var namespace = "openshift-validation-webhook"
	var service = "validation-webhook"
	var configMapName = "webhook-cert"
	var secretName = "webhook-cert"

	h := helper.New()

	ginkgo.It("should exist and be running in the cluster", func() {

		// Expect project to exist
		_, err := h.Project().ProjectV1().Projects().Get(context.TODO(), namespace, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "project should have been created")

		// Ensure presence of config map
		_, err = h.Kube().CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "failed to get config map %v\n", configMapName)

		// Ensure presence of secret
		_, err = h.Kube().CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "failed to get secretName %v\n", secretName)

		// Ensure pods are in a running state
		listOpts := metav1.ListOptions{}
		list, err := h.Kube().CoreV1().Pods(namespace).List(context.TODO(), listOpts)
		Expect(err).NotTo(HaveOccurred(), "failed to get running pods\n")
		Expect(list).NotTo(BeNil())
		Expect(list.Items).ShouldNot(HaveLen(0), "at least one pod should be present")
		for _, pod := range list.Items {
			phase := pod.Status.Phase
			Expect(phase).Should(Equal(kubev1.PodRunning), "pod should be in running state")
		}

		// Ensure service is present
		_, err = h.Kube().CoreV1().Services(namespace).Get(context.TODO(), service, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "service should have been created")

	}, 300)

})
