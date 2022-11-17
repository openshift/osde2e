package cloudingress

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cloudingress "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/common/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

var _ = ginkgo.Describe(constants.SuiteOperators+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
	})
	h := helper.New()
	ginkgo.Context("publishingstrategies", func() {
		util.GinkgoIt(
			"dedicated admin should not be allowed to manage publishingstrategies CR",
			func(ctx context.Context) {
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
				err := addPublishingstrategy(ctx, h, ps)
				Expect(apierrors.IsForbidden(err)).To(BeTrue())
			},
		)
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

func addPublishingstrategy(ctx context.Context, h *helper.H, publishingstrategy cloudingress.PublishingStrategy) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(publishingstrategy.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().
		Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).
		Namespace(OperatorNamespace).
		Create(ctx, &unstructuredObj, metav1.CreateOptions{})
	return err
}
