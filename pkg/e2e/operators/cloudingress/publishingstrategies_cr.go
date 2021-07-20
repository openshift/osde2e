package cloudingress

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudingress "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

var _ = ginkgo.Describe(constants.SuiteOperators+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool("rosa.STS") {
			ginkgo.Skip("STS does not support MVO")
		}
	})
	h := helper.New()
	ginkgo.Context("publishingstrategies", func() {
		ginkgo.It("dedicated admin should not be allowed to manage publishingstrategies CR", func() {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()
			ps := createPublishingstrategies("publishingstrategy-cr-test-1")
			err := addPublishingstrategy(h, ps)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

			//ingress, _ := getingressController(h, "default")
			//Expect(ingress.Annotations["Owner"]).To(Equal("cloud-ingress-operator"))
		})

		ginkgo.It("cluster admin should be allowed to manage publishingstrategies CR", func() {
			publishingstrategyName := "publishingstrategy-cr-test-2"
			ps := createPublishingstrategies(publishingstrategyName)
			err := addPublishingstrategy(h, ps)
			defer func() {
				publishingstrategyCleanup(h, publishingstrategyName)
			}()
			Expect(err).NotTo(HaveOccurred())
			//check the annotation for owned ingress
			//ingress, _ := getingressController(h, "default")
			//Expect(ingress.Annotations["Owner"]).To(Equal("cloud-ingress-operator"))
		})

	})

})

func createPublishingstrategies(name string) cloudingress.PublishingStrategy {
	publishingstrategy := cloudingress.PublishingStrategy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PublishingStrategy",
			APIVersion: cloudingress.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: cloudingress.PublishingStrategySpec{
			DefaultAPIServerIngress: cloudingress.DefaultAPIServerIngress{},
			ApplicationIngress:      []cloudingress.ApplicationIngress{},
		},
	}
	return publishingstrategy
}

func addPublishingstrategy(h *helper.H, publishingstrategy cloudingress.PublishingStrategy) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(publishingstrategy.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies",
	}).Namespace(OperatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
	return err
}

func publishingstrategyCleanup(h *helper.H, publishingstrategyName string) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies",
	}).Namespace(OperatorNamespace).Delete(context.TODO(), publishingstrategyName, metav1.DeleteOptions{})

}
