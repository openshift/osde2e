package operators

import (
	"context"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	osv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var certmanOperatorTestName string = "[Suite: operators] [OSD] Certman Operator"

func init() {
	alert.RegisterGinkgoAlert(certmanOperatorTestName, "SD-SREP", "Christoph Blecker", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(certmanOperatorTestName, func() {
	h := helper.New()

	ginkgo.Context("certificate secret should be applied when cluster installed", func() {
		var secretName string
		var secrets *v1.SecretList
		var err error
		var apiserver *osv1.APIServer

		ginkgo.It("certificate secret exist under openshift-config namespace", func() {
			wait.PollImmediate(30*time.Second, 15*time.Minute, func() (bool, error) {
				listOpts := metav1.ListOptions{
					LabelSelector: "certificate_request",
				}
				secrets, err = h.Kube().CoreV1().Secrets("openshift-config").List(context.TODO(), listOpts)
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
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("certificate secret should be applied to apiserver object", func() {
			wait.PollImmediate(30*time.Second, 15*time.Minute, func() (bool, error) {
				getOpts := metav1.GetOptions{}
				apiserver, err = h.Cfg().ConfigV1().APIServers().Get(context.TODO(), "cluster", getOpts)
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
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
