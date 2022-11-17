package cloudingress

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	operatorv1 "github.com/openshift/api/operator/v1"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/e2e/operators"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
	})

	h := helper.New()
	ginkgo.Context("Delete apps2 ingresscontroller", func() {
		ginkgo.It(
			"Should not be able to delete the ApplicationIngress that doesn't belong to CIO",
			func(ctx context.Context) {
				// 1. create a secondaryIngress that does belong to cloud-ingress-operator
				secondaryIngress := secondaryIngress(ctx, h)

				// only create the secondary ingress if it doesn't exist already in the publishing strategy
				if _, exists, _ := appIngressExits(ctx, h, false, secondaryIngress.DNSName); !exists {
					addAppIngress(ctx, h, secondaryIngress)
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
				Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
				Expect(deployment).NotTo(BeNil(), "deployment is nil")

				// 2.delete the annotation
				apps2Ingress, _ := getingressController(ctx, h, ingressControllerName)
				log.Printf("The ingresscontroller object annotation : %+v\n", apps2Ingress.ObjectMeta.Annotations)
				newAnnotations := updateAnnotation(ctx, h, ingressControllerName, "Owner", "cloud-ingress-operator")
				apps2Ingress = newAnnotations
				// 3. Delete secondaryIngress in publishingstrategy
				removeIngressController(ctx, h, ingressControllerName)
				Expect(err).NotTo(HaveOccurred())
				// check that the ingresscontroller app-e2e-apps was deleted
				ingressControllerExists(ctx, h, ingressControllerName, true)
				apps2Ingress = updateAnnotation(ctx, h, ingressControllerName, "Owner", "cloud-ingress-operator")
				removeIngressController(ctx, h, ingressControllerName)
				ingressControllerExists(ctx, h, ingressControllerName, false)
			},
			(10 * time.Minute).Seconds(),
		)
	})
})

func removeIngressController(ctx context.Context, h *helper.H, name string) error {
	_, exists, index := appIngressExits(ctx, h, false, name)
	// only remove the secondary ingress if it already exist in the publishing strategy
	if exists {
		removeAppIngress(ctx, h, index)
	}
	err := wait.PollImmediate(10*time.Second, 2*time.Minute, func() (bool, error) {
		_, err := h.Dynamic().
			Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).
			Namespace("openshift-ingress-operator").
			Get(ctx, name, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
	return err
}

func updateAnnotation(
	ctx context.Context,
	h *helper.H,
	name string,
	annotation1 string,
	annotation2 string,
) operatorv1.IngressController {
	var ingressController operatorv1.IngressController
	ingresscontroller, err := h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).
		Namespace("openshift-ingress-operator").
		Get(ctx, name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ingresscontroller.Object, &ingressController)
	Expect(err).NotTo(HaveOccurred())

	temp := ingressController.ObjectMeta
	// if annotation exists, delete it
	if temp.Annotations[annotation1] == annotation2 {
		delete(temp.Annotations, annotation1)
		ingressController.ObjectMeta = temp
		ingresscontroller.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&ingressController)
		Expect(err).NotTo(HaveOccurred())

		ingresscontroller, err = h.Dynamic().
			Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).
			Namespace("openshift-ingress-operator").
			Update(ctx, ingresscontroller, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
	} else {
		// if there's no annotation, add it
		annotation := map[string]string{
			annotation1: annotation2,
		}
		ingressController.ObjectMeta.Annotations = annotation
		ingresscontroller.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&ingressController)
		Expect(err).NotTo(HaveOccurred())

		ingresscontroller, err = h.Dynamic().
			Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).
			Namespace("openshift-ingress-operator").
			Update(ctx, ingresscontroller, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		updatePublishingStrategy(ctx, h, ingressController, name)
	}
	return ingressController
}

func updatePublishingStrategy(
	ctx context.Context,
	h *helper.H,
	ingressController operatorv1.IngressController,
	name string,
) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(ctx, h)
	var AppIngress cloudingressv1alpha1.ApplicationIngress
	// AppIngress.Listening = ingressController.Spec.EndpointPublishingStrategy.LoadBalancer.Scope
	AppIngress.Default = false
	AppIngress.DNSName = name
	PublishingStrategyInstance.Spec.ApplicationIngress = append(
		PublishingStrategyInstance.Spec.ApplicationIngress,
		AppIngress,
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
