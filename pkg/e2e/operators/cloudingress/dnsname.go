package cloudingress

import (
	"context"
	"log"
	"time"

	"github.com/onsi/ginkgo"
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
	var dnsnameOriginal string
	ginkgo.Context("publishingstrategy-dnsname", func() {
		ginkgo.It("IngressController should be patched when update dnsname", func() {
			ingress1, _ := getingressController(h, "default")
			dnsnameOriginal = string(ingress1.Spec.Domain)
			log.Print(" the Domain name \n", dnsnameOriginal)
			updateDnsName(h, "foo")

			time.Sleep(time.Duration(60) * time.Second)
			ingress, _ := getingressController(h, "default")
			Expect(ingress.Generation == int64(1)).To(Equal(false))
			Expect(ingress.Annotations["Owner"]).To(Equal("cloud-ingress-operator"))
		})
		ginkgo.It("IngressController should be patched when return to the original dnsname", func() {
			updateDnsName(h, dnsnameOriginal)

			time.Sleep(time.Duration(60) * time.Second)
			ingress, _ := getingressController(h, "default")
			Expect(string(ingress.Spec.Domain)).To(Equal(dnsnameOriginal))
			Expect(ingress.Generation == int64(1)).To(Equal(false))
			Expect(ingress.Annotations["Owner"]).To(Equal("cloud-ingress-operator"))
		})
	})
})

func updateDnsName(h *helper.H, newName string) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(h)
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
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}
