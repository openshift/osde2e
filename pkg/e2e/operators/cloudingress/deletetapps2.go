package cloudingress

import (
	"context"
	"log"
	"strings"

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
			time.Sleep(time.Duration(30) * time.Second)
			log.Print("Created a secondary ingress")
			// from DNSName app-e2e-apps.cluster.mfvz.s1.devshift.org,
			// the ingresscontroller name is app-e2e-apps: everything before the first period
			ingressControllerName := strings.Split(secondaryIngress.DNSName, ".")[0]
			log.Print("Got ingresscontroller Name")
			// check that the ingresscontroller app-e2e-apps was created
			ingressControllerExists(h, ingressControllerName, true)
			log.Printf("Check to see of the ingressControllerExists")
			// check if the secondary router is created
			// the created router name should be router-app-e2e-apps
			deploymentName := "router-" + ingressControllerName
			deployment, err := operators.PollDeployment(h, "openshift-ingress", deploymentName)
			log.Print("Check to see if the secondary router exists")
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")

			// wait 1 minute for all routers to start

			time.Sleep(time.Duration(60) * time.Second)
			//2.delete the annotation
			apps2Ingress, _ := getingressController(h, ingressControllerName)
			log.Printf("The ingresscontroller object annotation currently is: %+v\n", apps2Ingress.ObjectMeta.Annotations)
			newAnnotations := updateAnnotation(h, ingressControllerName, "Owner", "cloud-ingress-operator")
			apps2Ingress = newAnnotations
			log.Printf("Deleted the Annotation. This ingresscontroller should not belong to CIO anymore. the object's annotation now looks like: %+v\n", apps2Ingress.ObjectMeta.Annotations)
			//3. Delete secondaryIngress in publishingstrategy
			removeIngressController(h, ingressControllerName)
			log.Print("Try to delete the secondaryIngress that doesn't contains the annotation")
			time.Sleep(time.Duration(360) * time.Second)
			// check that the ingresscontroller app-e2e-apps was deleted
			ingressControllerExists(h, ingressControllerName, true)
			log.Print("Gonna clean up now")
			apps2Ingress = updateAnnotation(h, ingressControllerName, "Owner", "cloud-ingress-operator")
			log.Print("Added the Annotations back to IngressOperator")
			removeIngressController(h, ingressControllerName)
			log.Print("Called removeIngressController ")
			ingressControllerExists(h, ingressControllerName, false)
			log.Print("ingressControllerExist shouldn't throw any errors")
		})
	})
})

func removeIngressController(h *helper.H, name string) {
	_, exists, index := appIngressExits(h, false, name)
	// only remove the secondary ingress if it already exist in the publishing strategy
	if exists {
		removeAppIngress(h, index)
		//wait 2 minute for all resources to be deleted
		time.Sleep(time.Duration(120) * time.Second)
	}
}
func updateAnnotation(h *helper.H, name string, annotation1 string, annotation2 string) operatorv1.IngressController {
	var ingressController operatorv1.IngressController
	log.Print("Gonna start the annotation deletion process")
	ingresscontroller, err := h.Dynamic().Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).Namespace("openshift-ingress-operator").Get(context.TODO(), name, metav1.GetOptions{})
	log.Print("Getting the ingressController object")
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
		log.Printf("Set the IngressController Annotations to now: %+v", ingressController.ObjectMeta.Annotations)
		ingresscontroller.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&ingressController)
		Expect(err).NotTo(HaveOccurred())

		ingresscontroller, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "ingresscontrollers"}).Namespace("openshift-ingress-operator").Update(context.TODO(), ingresscontroller, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		updatePublishingStrategy(h, ingressController, name)
	}
	return ingressController
}

func updatePublishingStrategy(h *helper.H, ingressController operatorv1.IngressController, name string) {
	log.Print("Gonna update the PublishingStrategy")
	var err error
	PublishingStrategyInstance, ps := getPublishingStrategy(h)
	log.Print("Adding the IngressController back to the publishingstrategy")
	var AppIngress cloudingressv1alpha1.ApplicationIngress
	//AppIngress.Listening = ingressController.Spec.EndpointPublishingStrategy.LoadBalancer.Scope
	AppIngress.Default = false
	AppIngress.DNSName = name
	PublishingStrategyInstance.Spec.ApplicationIngress = append(PublishingStrategyInstance.Spec.ApplicationIngress, AppIngress)
	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())
	log.Print("Updated the publishingstrategy. should be able to delete this ingresscontroller")
	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}
