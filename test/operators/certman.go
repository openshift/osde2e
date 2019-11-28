package operators

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("[OSD] Certman Operator", func() {
	h := helper.New()
	secretName := h.ClusterName + "-" + "primary-cert-bundle-secret"
	ginkgo.Context("certificate secret  should be applied when cluster installed", func() {

		ginkgo.It("certificate secret exist under openshift-config namespace", func() {
			getOpts := metav1.GetOptions{}
			secret, err := h.Kube().CoreV1().Secrets("openshift-config").Get(secretName, getOpts)
			Expect(err).NotTo(HaveOccurred())
			Expect(secret).ShouldNot(Equal(nil))
		}, float64(h.PollingTimeout))

		ginkgo.It("certificate secret should be applied to apiserver object", func() {
			getOpts := metav1.GetOptions{}
			apiserver, err := h.Cfg().ConfigV1().APIServers().Get("cluster", getOpts)
			Expect(err).NotTo(HaveOccurred())
			Expect(apiserver.Spec.ServingCerts.NamedCertificates[0].ServingCertificate.Name).Should(Equal(secretName))
		}, float64(h.PollingTimeout))
	})
})
