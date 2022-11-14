package verify

import (
	"context"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
)

var validationWebhookTestName string = "[Suite: e2e] Validation Webhook"

func init() {
	alert.RegisterGinkgoAlert(validationWebhookTestName, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(validationWebhookTestName, func() {
	namespace := "openshift-validation-webhook"
	service := "validation-webhook"
	configMapName := "webhook-cert"
	secretName := "webhook-cert"

	h := helper.New()

	util.GinkgoIt("should exist and be running in the cluster", func(ctx context.Context) {
		// Expect project to exist
		_, err := h.Project().ProjectV1().Projects().Get(ctx, namespace, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "project should have been created")

		// Ensure presence of config map
		_, err = h.Kube().CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "failed to get config map %v\n", configMapName)

		// Ensure presence of secret
		_, err = h.Kube().CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "failed to get secretName %v\n", secretName)

		// Ensure pods are in a running state
		listOpts := metav1.ListOptions{}
		list, err := h.Kube().CoreV1().Pods(namespace).List(ctx, listOpts)
		Expect(err).NotTo(HaveOccurred(), "failed to get running pods\n")
		Expect(list).NotTo(BeNil())
		Expect(list.Items).ShouldNot(HaveLen(0), "at least one pod should be present")
		for _, pod := range list.Items {
			phase := pod.Status.Phase
			Expect(phase).Should(Equal(kubev1.PodRunning), "pod should be in running state")
		}

		// Ensure service is present
		_, err = h.Kube().CoreV1().Services(namespace).Get(ctx, service, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "service should have been created")
	}, 300)
})
