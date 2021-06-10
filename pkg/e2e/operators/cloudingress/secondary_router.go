package cloudingress

import (
	"context"
	"fmt"

	"strings"

	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/e2e/operators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	h := helper.New()
	ginkgo.Context("secondary router", func() {
		ginkgo.It("should be created when added to publishingstrategy ", func() {

			secondaryIngress := secondaryIngress(h)

			//only create the secondary ingress if it doesn't exist already in the publishing strategy
			if _, exists, _ := appIngressExits(h, false, secondaryIngress.DNSName); !exists {
				addAppIngress(h, secondaryIngress)
			}

			// from DNSName app-e2e-apps.cluster.mfvz.s1.devshift.org,
			// the ingresscontroller name is app-e2e-apps: everything before the first period
			ingressControllerName := strings.Split(secondaryIngress.DNSName, ".")[0]

			// check that the ingresscontroller app-e2e-apps was created
			ingressControllerExists(h, ingressControllerName, true)

			// check if the secondary router is created
			// the created router name should be router-app-e2e-apps
			deploymentName := "router-" + ingressControllerName
			deployment, err := operators.PollDeployment(h, "openshift-ingress", deploymentName)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")

			// wait 1 minute for all routers to start
			time.Sleep(time.Duration(60) * time.Second)
			Expect(deployment.Status.ReadyReplicas).To(BeNumerically("==", deployment.Status.Replicas))
		})

		ginkgo.It("should be deleted when removed from publishingstrategy", func() {
			secondaryIngress := secondaryIngress(h)

			_, exists, index := appIngressExits(h, false, secondaryIngress.DNSName)
			// only remove the secondary ingress if it already exist in the publishing strategy
			if exists {
				removeAppIngress(h, index)
				// wait 2 minute for all resources to be deleted
				time.Sleep(time.Duration(120) * time.Second)
			}

			ingressControllerName := strings.Split(secondaryIngress.DNSName, ".")[0]
			// check that the ingresscontroller app-e2e-apps was deleted
			ingressControllerExists(h, ingressControllerName, false)
		})
	})

})

// appIngressExits returns the appIngress matching the criteria if it exists
func appIngressExits(h *helper.H, isdefault bool, dnsname string) (appIngress cloudingressv1alpha1.ApplicationIngress, exists bool, index int) {
	PublishingStrategyInstance, _ := getPublishingStrategy(h)

	// Grab the current list of Application Ingresses from the Publishing Strategy
	AppIngressList := PublishingStrategyInstance.Spec.ApplicationIngress

	// Find the application ingress matching our criteria
	for i, v := range AppIngressList {
		if v.Default == isdefault && strings.HasPrefix(v.DNSName, dnsname) {
			appIngress = v
			exists = true
			index = i
			break
		}
	}
	return appIngress, exists, index
}

// secondaryIngress builds the secondary applicationIngress which is used in above tests
// by tweaking the default applicationIngress
func secondaryIngress(h *helper.H) cloudingressv1alpha1.ApplicationIngress {
	// first get the default ingresscontroller
	secondaryIngress, exists, _ := appIngressExits(h, true, "")
	Expect(exists).To(BeTrue())

	// then update it to create a secondary ingress
	secondaryIngress.Default = false
	secondaryIngress.DNSName = "app-e2e-" + secondaryIngress.DNSName

	return secondaryIngress
}

// getPublishingStrategy returns publishing strategies
func getPublishingStrategy(h *helper.H) (cloudingressv1alpha1.PublishingStrategy, *unstructured.Unstructured) {
	var PublishingStrategyInstance cloudingressv1alpha1.PublishingStrategy

	// Check that the PublishingStrategy CR is present
	ps, err := h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Get(context.TODO(), "publishingstrategy", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ps.Object, &PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	return PublishingStrategyInstance, ps
}

// addAppIngress  adds an application ingress to the default publishing strategy 's ApplicationIngressList
func addAppIngress(h *helper.H, appIngressToAppend cloudingressv1alpha1.ApplicationIngress) {
	var err error

	PublishingStrategyInstance, ps := getPublishingStrategy(h)
	PublishingStrategyInstance.Spec.ApplicationIngress = append(PublishingStrategyInstance.Spec.ApplicationIngress, appIngressToAppend)

	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

// removeAppIngress removes the application ingress at index `index` from the publishing strategy's ApplicationIngressList
func removeAppIngress(h *helper.H, index int) {
	var err error

	PublishingStrategyInstance, ps := getPublishingStrategy(h)

	// remove application ingress at index `index`
	appIngressList := PublishingStrategyInstance.Spec.ApplicationIngress
	PublishingStrategyInstance.Spec.ApplicationIngress = append(appIngressList[:index], appIngressList[index+1:]...)

	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

// ingressControllerExists checks if an Ingress controller was created or deleted
func ingressControllerExists(h *helper.H, ingressControllerName string, shouldexist bool) {
	_, err := h.Dynamic().Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).Namespace("openshift-ingress-operator").Get(context.TODO(), ingressControllerName, metav1.GetOptions{})
	if shouldexist {
		Expect(err).NotTo(HaveOccurred())
	} else {
		Expect(err).Should(MatchError(fmt.Sprintf("ingresscontrollers.operator.openshift.io \"%v\" not found", ingressControllerName)))
	}
}
