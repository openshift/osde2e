package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
)

var imageStreamsTestName string = "[Suite: e2e] ImageStreams"

func init() {
	alert.RegisterGinkgoAlert(imageStreamsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(imageStreamsTestName, func() {
	h := helper.New()

	util.GinkgoIt("should exist in the cluster", func(ctx context.Context) {
		list, err := h.Image().ImageV1().ImageStreams(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list ImageStreams")
		Expect(list).NotTo(BeNil())

		numImages := len(list.Items)
		minImages := 50
		Expect(numImages).Should(BeNumerically(">", minImages), "need more images")
	}, 300)
})
