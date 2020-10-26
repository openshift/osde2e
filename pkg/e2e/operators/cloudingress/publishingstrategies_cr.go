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
	testDaCRpublishingstrategies(h, operatorNamespace)
	testCRpublishingstrategies(h, operatorNamespace)

})

func createPublishingstrategies() cloudingress.PublishingStrategy {
	publishingstrategy := cloudingress.PublishingStrategy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PublishingStrategy",
			APIVersion: cloudingress.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "publshingstrategy-CR-test",
		},
	}
	return publishingstrategy
}

func addPublishingstrategy(h *helper.H, publishingstrategy cloudingress.PublishingStrategy, operatorNamespace string) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(publishingstrategy.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies",
	}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
	return err
}

func testDaCRpublishingstrategies(h *helper.H, operatorNamespace string) {
	ginkgo.Context("cloud-ingress-operator", func() {
		ginkgo.It("dedicated admin should not be allowed to manage publishingstrategies CR", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			ps := createPublishingstrategies()
			err := addPublishingstrategy(h, ps, operatorNamespace)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})
	})
}

func testCRpublishingstrategies(h *helper.H, operatorNamespace string) {
	ginkgo.Context("cloud-ingress-operator", func() {
		ginkgo.It("admin should be allowed to manage publishingstrategies CR", func() {
			ps := createPublishingstrategies()
			err := addPublishingstrategy(h, ps, operatorNamespace)
			Expect(err).NotTo(HaveOccurred())

		})
	})
}
