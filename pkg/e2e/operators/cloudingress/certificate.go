package cloudingress

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// tests
var _ = ginkgo.Describe("[Suite: informing] "+TestPrefix, label.Informing, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
	})

	h := helper.New()
	var originalCert string

	// How long to wait for IngressController changes
	pollingDuration := 120 * time.Second
	ginkgo.Context("publishingstrategy-certificate", func() {
		ginkgo.It("IngressController should be patched when update Certificate", func(ctx context.Context) {
			ingress1, _ := getingressController(ctx, h, "default")
			originalCert = string(ingress1.Spec.DefaultCertificate.Name)
			updateCertificate(ctx, h, "foo-bar")
			time.Sleep(pollingDuration)
			ingress, _ := getingressController(ctx, h, "default")

			Expect(string(ingress.Spec.DefaultCertificate.Name)).To(Equal("foo-bar"))
		}, pollingDuration.Seconds())

		ginkgo.It("IngressController should be patched when return the original Certificate", func(ctx context.Context) {
			updateCertificate(ctx, h, originalCert)
			time.Sleep(pollingDuration)
			ingress, _ := getingressController(ctx, h, "default")
			Expect(string(ingress.Spec.DefaultCertificate.Name)).To(Equal(originalCert))
		}, pollingDuration.Seconds())
	})
})

func updateCertificate(ctx context.Context, h *helper.H, newName string) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(ctx, h)

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
	ps, err = h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).
		Namespace(OperatorNamespace).
		Update(ctx, ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}
