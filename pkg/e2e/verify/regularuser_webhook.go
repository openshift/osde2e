package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/api/config/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var regularuserWebhookTestName string = "[Suite: service-definition] [OSD] regularuser validating webhook"

func init() {
	alert.RegisterGinkgoAlert(regularuserWebhookTestName, "SD-SREP", "Max Whittingham", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(regularuserWebhookTestName, label.ServiceDefinition, func() {
	h := helper.New()

	ginkgo.Context("regularuser validating webhook", func() {
		util.GinkgoIt("Privledged users allowed to create autoscalers and delete clusterversion objects", func(ctx context.Context) {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			defer func() {
				err := h.Cfg().ConfigV1().ClusterVersions().Delete(ctx, "osde2e-version", metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			}()
			_, err := h.Cfg().ConfigV1().ClusterVersions().Create(ctx, &v1.ClusterVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name: "osde2e-version",
				},
			}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}, float64(150))
	})
})
