package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
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
		util.GinkgoIt("Webhook will block UPDATE requests if TechPreviewNoUpgrade feature set exists in FeatureGate", func(ctx context.Context) {
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
			existingFeatureGate, err := h.Cfg().ConfigV1().FeatureGates().Get(ctx, "cluster", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			existingFeatureGate.Spec.FeatureSet = "TechPreviewNoUpgrade"

			_, err = h.Cfg().ConfigV1().FeatureGates().Update(ctx, existingFeatureGate, metav1.UpdateOptions{})

			Expect(err).To((HaveOccurred()))
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
