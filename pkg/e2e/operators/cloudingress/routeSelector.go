package cloudingress

import (
	"context"
	"reflect"
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

	// How long to wait for IngressController changes
	pollingDuration := 120 * time.Second
	ginkgo.Context("publishingstrategy-route-selector", func() {
		ginkgo.It("IngressController should be patched when update routeSelector matchLabels", func() {
			updateMatchLabels(h, "tier", "frontend")

			ingress, _ := getingressController(h, "default")
			Expect(string(ingress.Spec.RouteSelector.MatchLabels["tier"])).To(Equal("frontend"))
		}, pollingDuration.Seconds())
		ginkgo.It("IngressController should be patched when update routeSelector matchExpressions", func() {
			updateMatchExpressions(h, "foo", "In", "bar")

			ingress, _ := getingressController(h, "default")
			expectedExpressions := []metav1.LabelSelectorRequirement{
				{"foo", metav1.LabelSelectorOperator("In"), []string{"bar"}},
			}
			for j := range ingress.Spec.RouteSelector.MatchExpressions {
				Expect(reflect.DeepEqual(ingress.Spec.RouteSelector.MatchExpressions[j], expectedExpressions)).To(BeTrue())
			}
		}, pollingDuration.Seconds())
		ginkgo.It("IngressController should be patched when reset matchLabels and matchExpressions", func() {
			resetRouteSelector(h)

			ingress, _ := getingressController(h, "default")
			Expect(ingress.Spec.RouteSelector.MatchLabels).To(BeNil())
			Expect(ingress.Spec.RouteSelector.MatchExpressions).To(BeNil())

		}, pollingDuration.Seconds())
	})
})

func updateMatchLabels(h *helper.H, tier string, routeS string) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(h)

	// Grab the current list of Application Ingresses from the Publishing Strategy
	AppIngress := PublishingStrategyInstance.Spec.ApplicationIngress
	temp := map[string]string{
		tier: routeS,
	}
	// Find the default router and update its scheme
	for i, v := range AppIngress {
		if v.Default == true {
			AppIngress[i].RouteSelector.MatchLabels = temp
		}
	}

	PublishingStrategyInstance.Spec.ApplicationIngress = AppIngress

	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

func updateMatchExpressions(h *helper.H, key string, operator string, values string) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(h)
	// Grab the current list of Application Ingresses from the Publishing Strategy
	AppIngress := PublishingStrategyInstance.Spec.ApplicationIngress
	// Find the default router and update its scheme
	tempVal := []string{values}
	tempOp := metav1.LabelSelectorOperator(operator)
	temp := metav1.LabelSelectorRequirement{key, tempOp, tempVal}
	for i, v := range AppIngress {
		if v.Default == true {
			AppIngress[i].RouteSelector.MatchExpressions = []metav1.LabelSelectorRequirement{temp}
		}
	}
	PublishingStrategyInstance.Spec.ApplicationIngress = AppIngress
	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}
func resetRouteSelector(h *helper.H) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(h)
	// Grab the current list of Application Ingresses from the Publishing Strategy
	AppIngress := PublishingStrategyInstance.Spec.ApplicationIngress
	// Find the default router and update its scheme
	for i, v := range AppIngress {
		if v.Default == true {
			AppIngress[i].RouteSelector.MatchExpressions = append(AppIngress[i].RouteSelector.MatchExpressions[:i], AppIngress[i].RouteSelector.MatchExpressions[i+1:]...)
			delete(AppIngress[i].RouteSelector.MatchLabels, "tier")
		}
	}
	PublishingStrategyInstance.Spec.ApplicationIngress = AppIngress
	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}
