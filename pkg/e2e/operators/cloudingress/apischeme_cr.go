package cloudingress

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cloudingress "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

var _ = ginkgo.Describe(constants.SuiteOperators+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool("rosa.STS") {
			ginkgo.Skip("STS does not support MVO")
		}
	})
	h := helper.New()
	ginkgo.Context("apischeme", func() {
		util.GinkgoIt("apischemes CR instance must be present on cluster", func(ctx context.Context) {
			err := wait.PollImmediate(2*time.Second, 2*time.Minute, func() (bool, error) {
				if _, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
				}).Namespace(OperatorNamespace).Get(ctx, apiSchemeResourceName, metav1.GetOptions{}); err != nil {
					return false, nil
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		})

		util.GinkgoIt("dedicated admin should not be allowed to manage apischemes CR", func(ctx context.Context) {
			user := "test-user"
			impersonateDedicatedAdmin(h, user)

			defer func() {
				apiSchemeCleanup(ctx, h, "apischeme-osde2e-test")
			}()
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			as := createApischeme("apischeme-osde2e-test")
			err := addApischeme(ctx, h, as)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

			_, err = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
			}).Namespace(OperatorNamespace).Get(ctx, "apischeme-osde2e-test", metav1.GetOptions{})
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})

		util.GinkgoIt("cluster admin should be allowed to manage apischemes CR", func(ctx context.Context) {
			as := createApischeme("apischeme-cr-test")
			defer apiSchemeCleanup(ctx, h, "apischeme-cr-test")
			err := addApischeme(ctx, h, as)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func createApischeme(name string) cloudingress.APIScheme {
	apischeme := cloudingress.APIScheme{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIScheme",
			APIVersion: cloudingress.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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

func addApischeme(ctx context.Context, h *helper.H, apischeme cloudingress.APIScheme) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apischeme.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
	}).Namespace(OperatorNamespace).Create(ctx, &unstructuredObj, metav1.CreateOptions{})
	return err
}

func apiSchemeCleanup(ctx context.Context, h *helper.H, apiSchemeName string) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
	}).Namespace(OperatorNamespace).Delete(ctx, apiSchemeName, metav1.DeleteOptions{})
}

func impersonateDedicatedAdmin(h *helper.H, user string) *helper.H {
	h.Impersonate(rest.ImpersonationConfig{
		UserName: user,
		Groups: []string{
			"dedicated-admins",
		},
	})

	return h
}
