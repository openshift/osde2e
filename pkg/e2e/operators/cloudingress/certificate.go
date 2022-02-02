package cloudingress

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// tests
var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	h := helper.New()
	var originalCert string

	// How long to wait for IngressController changes
	pollingDuration := 120 * time.Second
	ginkgo.Context("publishingstrategy-certificate", func() {
		ginkgo.It("IngressController should be patched when update Certificate", func() {
			ingress1, _ := getingressController(h, "default")
			originalCert = string(ingress1.Spec.DefaultCertificate.Name)
			updateCertificate(h, "foo-bar")
			time.Sleep(pollingDuration)
			ingress, _ := getingressController(h, "default")

			Expect(string(ingress.Spec.DefaultCertificate.Name)).To(Equal("foo-bar"))
			Expect(ingress.Generation == int64(1)).To(Equal(false))
			Expect(ingress.Annotations["Owner"]).To(Equal("cloud-ingress-operator"))
		}, pollingDuration.Seconds())

		ginkgo.It("IngressController should be patched when return the original Certificate", func() {
			updateCertificate(h, originalCert)
			time.Sleep(pollingDuration)
			ingress, _ := getingressController(h, "default")
			Expect(string(ingress.Spec.DefaultCertificate.Name)).To(Equal(originalCert))
			Expect(ingress.Generation == int64(1)).To(Equal(false))
			Expect(ingress.Annotations["Owner"]).To(Equal("cloud-ingress-operator"))
		}, pollingDuration.Seconds())
	})
})

func updateCertificate(h *helper.H, newName string) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(h)

	// Grab the current list of Application Ingresses from the Publishing Strategy
	AppIngress := PublishingStrategyInstance.Spec.ApplicationIngress
	name := newName
	// Find the default router and update its scheme
	for i, v := range AppIngress {
		if v.Default == true {
			AppIngress[i].Certificate.Name = name
		}
	}

	PublishingStrategyInstance.Spec.ApplicationIngress = AppIngress

	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}
