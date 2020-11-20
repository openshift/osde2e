package cloudingress

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudingress "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	"github.com/openshift/osde2e/pkg/common/helper"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = ginkgo.Describe(CloudIngressTestName, func() {
	var operatorNamespace = "openshift-cloud-ingress-operator"

	h := helper.New()
	testCRapiSchemesPresent(h)
	testDaCRapischemes(h, operatorNamespace)
	testCRapischemes(h, operatorNamespace)

})

func createApischeme() cloudingress.APIScheme {
	apischeme := cloudingress.APIScheme{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIScheme",
			APIVersion: cloudingress.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "apischeme-cr-test",
		},
		Spec: cloudingress.APISchemeSpec{
			ManagementAPIServerIngress: cloudingress.ManagementAPIServerIngress{
				Enabled:           false,
				DNSName:           "osde2e",
				AllowedCIDRBlocks: []string{},
			},
		},
	}
	return apischeme
}

func addApischeme(h *helper.H, apischeme cloudingress.APIScheme, operatorNamespace string) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apischeme.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
	}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
	return err
}

func testDaCRapischemes(h *helper.H, operatorNamespace string) {
	ginkgo.Context("apischeme-cr-test", func() {
		ginkgo.It("dedicated admin should not be allowed to manage apischemes CR", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			as := createApischeme()
			err := addApischeme(h, as, operatorNamespace)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})
	})
}

func testCRapischemes(h *helper.H, operatorNamespace string) {
	ginkgo.Context("apischeme-cr-test", func() {
		ginkgo.It("admin should be allowed to manage apischemes CR", func() {
			as := createApischeme()
			err := addApischeme(h, as, operatorNamespace)
			Expect(err).NotTo(HaveOccurred())

		})
	})
}

func testCRapiSchemesPresent(h *helper.H) {
	ginkgo.Context("apischeme-cr-test", func() {
		ginkgo.It("ApiSchemeCR must be present on cluster", func() {
			_, err := h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
			}).Namespace(CloudIngressNamespace).Get(context.TODO(), apiSchemeResourceName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
