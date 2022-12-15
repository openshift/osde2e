package cloudingress

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	operatorv1 "github.com/openshift/api/operator/v1"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/api/v1alpha1"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/common/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

// tests
var _ = ginkgo.Describe("[Suite: informing] "+TestPrefix, label.Informing, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
	})

	h := helper.New()

	// How long to wait for changes to be effected on cluster
	pollingDuration := 5 * time.Minute

	ginkgo.Context("publishingstrategy-public-private", func() {
		util.GinkgoIt(
			"should be able to toggle the default applicationingress from public to private",
			func(ctx context.Context) {
				updateApplicationIngress(ctx, h, "internal")

				// wait for router-default service loadbalancer to have an annotation indicating its scheme is internal
				err := wait.PollImmediate(10*time.Second, pollingDuration, func() (bool, error) {
					service, err := h.Kube().
						CoreV1().
						Services("openshift-ingress").
						Get(ctx, "router-default", metav1.GetOptions{})
					if err != nil {
						return false, nil
					}
					if _, ok := service.Annotations["service.beta.kubernetes.io/aws-load-balancer-internal"]; ok == true {
						return true, nil
					}
					return false, nil
				})
				Expect(err).NotTo(HaveOccurred())

				ingress, _ := getingressController(ctx, h, "default")
				Expect(string(ingress.Spec.EndpointPublishingStrategy.LoadBalancer.Scope)).To(Equal("Internal"))
			},
			pollingDuration.Seconds(),
		)

		util.GinkgoIt(
			"should be able to toggle the default applicationingress from private to public",
			func(ctx context.Context) {
				updateApplicationIngress(ctx, h, "external")
				// wait for router-default service loadbalancer to NOT have an annotation indicating its scheme is internal
				err := wait.PollImmediate(10*time.Second, pollingDuration, func() (bool, error) {
					service, err := h.Kube().
						CoreV1().
						Services("openshift-ingress").
						Get(ctx, "router-default", metav1.GetOptions{})
					if err != nil {
						return false, nil
					}
					if _, ok := service.Annotations["service.beta.kubernetes.io/aws-load-balancer-internal"]; ok == false {
						return true, nil
					}

					return false, nil
				})
				Expect(err).NotTo(HaveOccurred())

				ingress_controller, exists, _ := appIngressExits(ctx, h, true, "")
				Expect(exists).To(BeTrue())
				Expect(string(ingress_controller.Listening)).To(Equal("external"))
			},
			pollingDuration.Seconds(),
		)
	})
})

// utils

func updateApplicationIngress(ctx context.Context, h *helper.H, lbscheme string) {
	var PublishingStrategyInstance cloudingressv1alpha1.PublishingStrategy

	// Check that the PublishingStrategy CR is present
	ps, err := h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).
		Namespace(OperatorNamespace).
		Get(ctx, "publishingstrategy", metav1.GetOptions{})
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
	ps, err = h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).
		Namespace(OperatorNamespace).
		Update(ctx, ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

func getingressController(
	ctx context.Context,
	h *helper.H,
	name string,
) (operatorv1.IngressController, *unstructured.Unstructured) {
	var ingressController operatorv1.IngressController
	ingresscontroller, err := h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).
		Namespace("openshift-ingress-operator").
		Get(ctx, name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ingresscontroller.Object, &ingressController)
	Expect(err).NotTo(HaveOccurred())

	return ingressController, ingresscontroller
}

// common setup and utils are in cloudingress.go
