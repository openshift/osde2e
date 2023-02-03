package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

var imageStreamsTestName string = "[Suite: e2e] ImageStreams"

func init() {
	alert.RegisterGinkgoAlert(imageStreamsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(imageStreamsTestName, ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.It("should exist in the cluster", func(ctx context.Context) {
		var list imagev1.ImageList
		Eventually(func(g Gomega) int {
			err := client.WithNamespace(metav1.NamespaceAll).List(ctx, &list)
			g.Expect(err).ToNot(HaveOccurred(), "unable to list images")
			return len(list.Items)
		}, "5m").Should(BeNumerically(">", 0), "no images found")
	})
})
