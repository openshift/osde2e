package operators

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	osv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var certmanOperatorTestName string = "[Suite: operators] [OSD] Certman Operator"

func init() {
	alert.RegisterGinkgoAlert(certmanOperatorTestName, "SD-SREP", "@certman-operator", "hcm-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(certmanOperatorTestName, label.Operators, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("Certman Operator is not supported on HyperShift")
		}
	})

	h := helper.New()

	ginkgo.Context("certificate secret should be applied when cluster installed", func() {
		var secretName string
		var secrets *v1.SecretList
		var apiserver *osv1.APIServer

		// Waiting period to wait for certman resources to appear
		pollingDuration := 15 * time.Minute
		ginkgo.It("certificate secret exist under openshift-config namespace", func(ctx context.Context) {
			err := wait.PollUntilContextTimeout(ctx, 30*time.Second, pollingDuration, true, func(ctx context.Context) (bool, error) {
				listOpts := metav1.ListOptions{
					LabelSelector: "certificate_request",
				}
				var err error
				secrets, err = h.Kube().CoreV1().Secrets("openshift-config").List(ctx, listOpts)
				if err != nil {
					return false, err
				}
				if len(secrets.Items) < 1 {
					return false, nil
				}
				secretName = secrets.Items[0].Name
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(secrets.Items)).Should(Equal(1))
		}, pollingDuration.Seconds())

		ginkgo.It("certificate secret should be applied to apiserver object", func(ctx context.Context) {
			err := wait.PollUntilContextTimeout(ctx, 30*time.Second, pollingDuration, true, func(ctx context.Context) (bool, error) {
				getOpts := metav1.GetOptions{}
				var err error
				apiserver, err = h.Cfg().ConfigV1().APIServers().Get(ctx, "cluster", getOpts)
				if err != nil {
					return false, err
				}
				if len(apiserver.Spec.ServingCerts.NamedCertificates) < 1 {
					return false, nil
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(apiserver.Spec.ServingCerts.NamedCertificates)).Should(BeNumerically(">", 0))
			Expect(apiserver.Spec.ServingCerts.NamedCertificates[0].ServingCertificate.Name).Should(Equal(secretName))
		}, pollingDuration.Seconds())
	})
})
