package cloudingress

import (
	"context"
	"log"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"

	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

// tests
var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	h := helper.New()

	ginkgo.Context("publishingstrategy-public-private", func() {
		ginkgo.It("should be able to toggle the default applicationingress from public to private", func() {

			updateApplicationIngress(h, "internal")

			//wait for router-default service loadbalancer to have an annotation indicating its scheme is internal
			err := wait.PollImmediate(10*time.Second, 5*time.Minute, func() (bool, error) {
				service, err := h.Kube().CoreV1().Services("openshift-ingress").Get(context.TODO(), "router-default", metav1.GetOptions{})
				if err != nil {
					log.Printf("Waiting for router-default service in openshift-ingress namespace to be private")
					return false, nil
				}
				if _, ok := service.Annotations["service.beta.kubernetes.io/aws-load-balancer-internal"]; ok == true {
					log.Printf("router-default service in openshift-ingress namespace successfully switched to private")
					return true, nil
				}
				log.Printf("Waiting for router-default service in openshift-ingress namespace to be private")
				return false, nil
			})
			Expect(err).NotTo(HaveOccurred())
		})
		ginkgo.It("should be able to toggle the default applicationingress from private to public", func() {

			updateApplicationIngress(h, "external")

			//wait for router-default service loadbalancer to NOT have an annotation indicating its scheme is internal
			err := wait.PollImmediate(10*time.Second, 5*time.Minute, func() (bool, error) {
				service, err := h.Kube().CoreV1().Services("openshift-ingress").Get(context.TODO(), "router-default", metav1.GetOptions{})
				if err != nil {
					log.Printf("Waiting for router-default service in openshift-ingress namespace to be public")
					return false, nil
				}
				if _, ok := service.Annotations["service.beta.kubernetes.io/aws-load-balancer-internal"]; ok == false {
					log.Printf("router-default service in openshift-ingress namespace successfully switched to public")
					return true, nil
				}
				log.Printf("Waiting for router-default service in openshift-ingress namespace to be public")
				return false, nil
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// utils

func updateApplicationIngress(h *helper.H, lbscheme string) {
	var PublishingStrategyInstance cloudingressv1alpha1.PublishingStrategy

	// Check that the PublishingStrategy CR is present
	ps, err := h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Get(context.TODO(), "publishingstrategy", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ps.Object, &PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Grab the current list of Application Ingresses from the Publishing Strategy
	AppIngress := PublishingStrategyInstance.Spec.ApplicationIngress

	// Find the default router and update its scheme
	for i, v := range AppIngress {
		if v.Default == true {
			AppIngress[i].Listening = cloudingressv1alpha1.Listening(lbscheme)
		}
	}

	PublishingStrategyInstance.Spec.ApplicationIngress = AppIngress

	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

// common setup and utils are in cloudingress.go
