package cloudingress

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/api/v1alpha1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/e2e/operators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = ginkgo.Describe("[Suite: informing] "+TestPrefix, label.Informing, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
	})

	h := helper.New()
	ginkgo.Context("secondary router", func() {
		// How long to wait for resources to be created
		pollingDuration := 2 * time.Minute

		util.GinkgoIt("should be created when added to publishingstrategy ", func(ctx context.Context) {
			secondaryIngress := secondaryIngress(ctx, h)

			// only create the secondary ingress if it doesn't exist already in the publishing strategy
			if _, exists, _ := appIngressExits(ctx, h, false, secondaryIngress.DNSName); !exists {
				addAppIngress(ctx, h, secondaryIngress)
				// wait 2 minute for all resources to be created
				time.Sleep(pollingDuration)
			}

			// from DNSName app-e2e-apps.cluster.mfvz.s1.devshift.org,
			// the ingresscontroller name is app-e2e-apps: everything before the first period
			ingressControllerName := strings.Split(secondaryIngress.DNSName, ".")[0]

			// check that the ingresscontroller app-e2e-apps was created
			ingressControllerExists(ctx, h, ingressControllerName, true)

			// check if the secondary router is created
			// the created router name should be router-app-e2e-apps
			deploymentName := "router-" + ingressControllerName
			deployment, err := operators.PollDeployment(ctx, h, "openshift-ingress", deploymentName)
			ingress, _ := getingressController(ctx, h, ingressControllerName)

			Expect(ingress.Annotations["Owner"]).To(Equal("cloud-ingress-operator"))
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")
			Expect(deployment.Status.ReadyReplicas).To(BeNumerically("==", deployment.Status.Replicas))
		}, pollingDuration.Seconds()+viper.GetFloat64(config.Tests.PollingTimeout))

		util.GinkgoIt("should be deleted when removed from publishingstrategy", func(ctx context.Context) {
			secondaryIngress := secondaryIngress(ctx, h)

			_, exists, index := appIngressExits(ctx, h, false, secondaryIngress.DNSName)
			// only remove the secondary ingress if it already exist in the publishing strategy
			if exists {
				removeAppIngress(ctx, h, index)
				// wait 2 minute for all resources to be deleted
				time.Sleep(pollingDuration)
			}
			Expect(len(strings.Split(secondaryIngress.DNSName, "."))).To(BeNumerically(">", 1))
			ingressControllerName := strings.Split(secondaryIngress.DNSName, ".")[1]
			// check that the ingresscontroller app-e2e-apps was deleted
			ingressControllerExists(ctx, h, ingressControllerName, false)
		}, pollingDuration.Seconds()+viper.GetFloat64(config.Tests.PollingTimeout))
	})
})

// appIngressExits returns the appIngress matching the criteria if it exists
func appIngressExits(
	ctx context.Context,
	h *helper.H,
	isdefault bool,
	dnsname string,
) (appIngress cloudingressv1alpha1.ApplicationIngress, exists bool, index int) {
	PublishingStrategyInstance, _ := getPublishingStrategy(ctx, h)

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
func secondaryIngress(ctx context.Context, h *helper.H) cloudingressv1alpha1.ApplicationIngress {
	// first get the default ingresscontroller
	secondaryIngress, exists, _ := appIngressExits(ctx, h, true, "")
	Expect(exists).To(BeTrue())

	// then update it to create a secondary ingress
	secondaryIngress.Default = false
	secondaryIngress.DNSName = "app-e2e-" + secondaryIngress.DNSName

	return secondaryIngress
}

// getPublishingStrategy returns publishing strategies
func getPublishingStrategy(
	ctx context.Context,
	h *helper.H,
) (cloudingressv1alpha1.PublishingStrategy, *unstructured.Unstructured) {
	var PublishingStrategyInstance cloudingressv1alpha1.PublishingStrategy

	// Check that the PublishingStrategy CR is present
	ps, err := h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).
		Namespace(OperatorNamespace).
		Get(ctx, "publishingstrategy", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ps.Object, &PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	return PublishingStrategyInstance, ps
}

// addAppIngress  adds an application ingress to the default publishing strategy 's ApplicationIngressList
func addAppIngress(ctx context.Context, h *helper.H, appIngressToAppend cloudingressv1alpha1.ApplicationIngress) {
	var err error

	PublishingStrategyInstance, ps := getPublishingStrategy(ctx, h)
	PublishingStrategyInstance.Spec.ApplicationIngress = append(
		PublishingStrategyInstance.Spec.ApplicationIngress,
		appIngressToAppend,
	)

	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).
		Namespace(OperatorNamespace).
		Update(ctx, ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

// removeAppIngress removes the application ingress at index `index` from the publishing strategy's ApplicationIngressList
func removeAppIngress(ctx context.Context, h *helper.H, index int) {
	var err error

	PublishingStrategyInstance, ps := getPublishingStrategy(ctx, h)

	// remove application ingress at index `index`
	appIngressList := PublishingStrategyInstance.Spec.ApplicationIngress
	PublishingStrategyInstance.Spec.ApplicationIngress = append(appIngressList[:index], appIngressList[index+1:]...)

	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).
		Namespace(OperatorNamespace).
		Update(ctx, ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

// ingressControllerExists checks if an Ingress controller was created or deleted
func ingressControllerExists(ctx context.Context, h *helper.H, ingressControllerName string, shouldexist bool) {
	_, err := h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).
		Namespace("openshift-ingress-operator").
		Get(ctx, ingressControllerName, metav1.GetOptions{})
	if shouldexist {
		Expect(err).NotTo(HaveOccurred())
	} else {
		Expect(err).Should(MatchError(fmt.Sprintf("ingresscontrollers.operator.openshift.io \"%v\" not found", ingressControllerName)))
	}
}
