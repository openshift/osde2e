package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var testAlert alert.MetricAlert

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: e2e] ImageStreams",
		TeamOwner:        "SD-CICD",
		PrimaryContact:   "Jeffrey Sica",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 4,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name, func() {
	h := helper.New()

	ginkgo.It("should exist in the cluster", func() {
		list, err := h.Image().ImageV1().ImageStreams(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list ImageStreams")
		Expect(list).NotTo(BeNil())

		numImages := len(list.Items)
		minImages := 50
		Expect(numImages).Should(BeNumerically(">", minImages), "need more images")
	}, 300)
})
