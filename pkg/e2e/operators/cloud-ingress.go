package operators

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudingress "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var cloudIngressOperatorTestName string = "[Suite: operators] [OSD] Cloud Ingress Operator"

func init() {
	alert.RegisterGinkgoAlert(cloudIngressOperatorTestName, "SD-SREP", "Alice Hubenko", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(cloudIngressOperatorTestName, func() {
	var operatorName = "cloud-ingress-operator"
	var operatorNamespace = "openshift-cloud-ingress-operator"
	var defaultDesiredReplicas int32 = 1

	h := helper.New()
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	testDaCRpublishingstrategies(h)
	testDaCRapischeme(h)

})

func testDaCRapischeme(h *helper.H) {
	ginkgo.Context("cloud-ingress-operator", func() {
		ginkgo.It("dedicated admin should not be allowed to manage apischemes CR", func() {
			apischeme := cloudingress.APIScheme{TypeMeta: metav1.TypeMeta{
				Kind: "APIScheme",
			},
				ObjectMeta: metav1.ObjectMeta{
					Name: "apischeme-CR-test",
				},
			}
			obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apischeme.DeepCopy())
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			unstructuredObj := unstructured.Unstructured{obj}
			_, err = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
			}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})
	})
}

func testDaCRpublishingstrategies(h *helper.H) {
	ginkgo.Context("cloud-ingress-operator", func() {
		ginkgo.It("dedicated admin should not be allowed to manage publishingstrategies CR", func() {
			publishingstrategies := cloudingress.PublishingStrategy{
				TypeMeta: metav1.TypeMeta{
					Kind: "PublishingStrategy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "publshingstrategy-CR-test",
				},
			}
			obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(publishingstrategies.DeepCopy())
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			unstructuredObj := unstructured.Unstructured{obj}
			_, err = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies",
			}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})
	})
}

func testAddApischeme(h *helper.H) {
	ginkgo.Context("cloud-ingress-operator", func() {
		ginkgo.It("admin should be allowed to manage apischemes CR", func() {
			apischeme := cloudingress.APIScheme{TypeMeta: metav1.TypeMeta{
				Kind: "APIScheme",
			},
				ObjectMeta: metav1.ObjectMeta{
					Name: "apischeme-CR-test",
				},
			}
			obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apischeme.DeepCopy())
			unstructuredObj := unstructured.Unstructured{obj}
			_, err = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
			}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

		})
	})
}

func testAddPublishingstrategies(h *helper.H) {
	ginkgo.Context("cloud-ingress-operator", func() {
		ginkgo.It("dedicated admin should not be allowed to manage publishingstrategies CR", func() {
			publishingstrategies := cloudingress.PublishingStrategy{
				TypeMeta: metav1.TypeMeta{
					Kind: "PublishingStrategy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "publshingstrategy-CR-test",
				},
			}
			obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(publishingstrategies.DeepCopy())
			unstructuredObj := unstructured.Unstructured{obj}
			_, err = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies",
			}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

		})
	})
}
