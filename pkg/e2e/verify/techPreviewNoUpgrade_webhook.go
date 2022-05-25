package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var techPreviewNoUpgradeWebhookTestName string = "[Suite: informing] [OSD] techpreviewnoupgrade blocking webhook"

func init() {
	alert.RegisterGinkgoAlert(techPreviewNoUpgradeWebhookTestName, "SD-SREP", "Tafhim Ul Islam", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(techPreviewNoUpgradeWebhookTestName, func() {

	h := helper.New()

	ginkgo.Context("techpreviewnoupgrade webhook", func() {
		util.GinkgoIt("Webhook will not block CREATE requests if TechPreviewNoUpgrade FeatureGate does not exist", func() {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"cluster-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			// first fetch the ResourceVersion
			existingFeatureGate, err := h.Cfg().ConfigV1().FeatureGates().Get(context.TODO(), "cluster", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			existingFeatureGate.Spec.FeatureSet = "TechPreviewNoUpgrade"

			_, err = h.Cfg().ConfigV1().FeatureGates().Update(context.TODO(), existingFeatureGate, metav1.UpdateOptions{})

			Expect(err).To((HaveOccurred()))
		})
	})

})
