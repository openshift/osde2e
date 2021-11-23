package cloudingress

import (
	"context"
	"log"
	"strings"
<<<<<<< HEAD
=======

>>>>>>> 14cdbbc12c93f30a7157c3b8488debda75369133
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	operatorv1 "github.com/openshift/api/operator/v1"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/e2e/operators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
<<<<<<< HEAD
	"k8s.io/apimachinery/pkg/util/wait"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
=======
>>>>>>> 14cdbbc12c93f30a7157c3b8488debda75369133
)

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	h := helper.New()
	ginkgo.Context("Delete apps2 ingresscontroller", func() {
		ginkgo.It("Should not be able to delete the ApplicationIngress that doesn't belong to CIO", func() {
			//1. create a secondaryIngress that does belong to cloud-ingress-operator
			secondaryIngress := secondaryIngress(h)

			//only create the secondary ingress if it doesn't exist already in the publishing strategy
			if _, exists, _ := appIngressExits(h, false, secondaryIngress.DNSName); !exists {
				addAppIngress(h, secondaryIngress)
			}
<<<<<<< HEAD
=======
			time.Sleep(time.Duration(30) * time.Second)
>>>>>>> 14cdbbc12c93f30a7157c3b8488debda75369133
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

<<<<<<< HEAD

=======
			// wait 1 minute for all routers to start

			time.Sleep(time.Duration(60) * time.Second)
>>>>>>> 14cdbbc12c93f30a7157c3b8488debda75369133
			//2.delete the annotation
			apps2Ingress, _ := getingressController(h, ingressControllerName)
			log.Printf("The ingresscontroller object annotation : %+v\n", apps2Ingress.ObjectMeta.Annotations)
			newAnnotations := updateAnnotation(h, ingressControllerName, "Owner", "cloud-ingress-operator")
			apps2Ingress = newAnnotations
			//3. Delete secondaryIngress in publishingstrategy
			removeIngressController(h, ingressControllerName)
<<<<<<< HEAD
			Expect(err).NotTo(HaveOccurred())
=======
			//long sleep statement to make sure secondaryIngress is done deleting
			time.Sleep(time.Duration(360) * time.Second)
>>>>>>> 14cdbbc12c93f30a7157c3b8488debda75369133
			// check that the ingresscontroller app-e2e-apps was deleted
			ingressControllerExists(h, ingressControllerName, true)
			apps2Ingress = updateAnnotation(h, ingressControllerName, "Owner", "cloud-ingress-operator")
			removeIngressController(h, ingressControllerName)
			ingressControllerExists(h, ingressControllerName, false)
		})
	})
})

<<<<<<< HEAD
func removeIngressController(h *helper.H, name string) error {
=======
func removeIngressController(h *helper.H, name string) {
>>>>>>> 14cdbbc12c93f30a7157c3b8488debda75369133
	_, exists, index := appIngressExits(h, false, name)
	// only remove the secondary ingress if it already exist in the publishing strategy
	if exists {
		removeAppIngress(h, index)
<<<<<<< HEAD
	}
	err := wait.PollImmediate(10*time.Second, 2*time.Minute, func() (bool, error) {
		_, err := h.Dynamic().Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).Namespace("openshift-ingress-operator").Get(context.TODO(), name, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
	return err
=======
		//wait 2 minute for all resources to be deleted
		time.Sleep(time.Duration(120) * time.Second)
	}
>>>>>>> 14cdbbc12c93f30a7157c3b8488debda75369133
}
func updateAnnotation(h *helper.H, name string, annotation1 string, annotation2 string) operatorv1.IngressController {
	var ingressController operatorv1.IngressController
	ingresscontroller, err := h.Dynamic().Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).Namespace("openshift-ingress-operator").Get(context.TODO(), name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ingresscontroller.Object, &ingressController)
	Expect(err).NotTo(HaveOccurred())

	temp := ingressController.ObjectMeta
	//if annotation exists, delete it
	if temp.Annotations[annotation1] == annotation2 {
		delete(temp.Annotations, annotation1)
		ingressController.ObjectMeta = temp
		ingresscontroller.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&ingressController)
		Expect(err).NotTo(HaveOccurred())

		ingresscontroller, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).Namespace("openshift-ingress-operator").Update(context.TODO(), ingresscontroller, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
	} else {
		//if there's no annotation, add it
		annotation := map[string]string{
			annotation1: annotation2,
		}
		ingressController.ObjectMeta.Annotations = annotation
		ingresscontroller.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&ingressController)
		Expect(err).NotTo(HaveOccurred())

		ingresscontroller, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).Namespace("openshift-ingress-operator").Update(context.TODO(), ingresscontroller, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		updatePublishingStrategy(h, ingressController, name)
	}
	return ingressController
}

func updatePublishingStrategy(h *helper.H, ingressController operatorv1.IngressController, name string) {
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(h)
	var AppIngress cloudingressv1alpha1.ApplicationIngress
	//AppIngress.Listening = ingressController.Spec.EndpointPublishingStrategy.LoadBalancer.Scope
	AppIngress.Default = false
	AppIngress.DNSName = name
	PublishingStrategyInstance.Spec.ApplicationIngress = append(PublishingStrategyInstance.Spec.ApplicationIngress, AppIngress)
	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())
	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}
