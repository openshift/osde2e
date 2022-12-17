package cloudingress

import (
	"context"
	"log"
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
	var dnsnameOriginal string
	// How long to wait for IngressController changes
	pollingDuration := 120 * time.Second
	ginkgo.Context("publishingstrategy-dnsname", func() {
		ginkgo.It("IngressController should be patched when update dnsname", func(ctx context.Context) {
			ingress1, _ := getingressController(ctx, h, "default")
			dnsnameOriginal = string(ingress1.Spec.Domain)
			log.Print(" the Domain name \n", dnsnameOriginal)
			updateDnsName(ctx, h, "foo")

			ingress, _ := getingressController(ctx, h, "default")
			log.Print(" The new Generation is \n", ingress.Generation)
		}, pollingDuration.Seconds())

		ginkgo.It("IngressController should be patched when return to the original dnsname", func(ctx context.Context) {
			updateDnsName(ctx, h, dnsnameOriginal)

			ingress, _ := getingressController(ctx, h, "default")
			Expect(string(ingress.Spec.Domain)).To(Equal(dnsnameOriginal))
		}, pollingDuration.Seconds())
	})
})

func updateDnsName(ctx context.Context, h *helper.H, newName string) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(ctx, h)
	AppIngress := PublishingStrategyInstance.Spec.ApplicationIngress

	for i, v := range AppIngress {
		if v.Default == true {
			AppIngress[i].DNSName = newName
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
