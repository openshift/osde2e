package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var imageStreamsTestName string = "[Suite: e2e] ImageStreams"

func init() {
	alert.RegisterGinkgoAlert(imageStreamsTestName, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(imageStreamsTestName, func() {
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
