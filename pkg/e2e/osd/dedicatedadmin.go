package osd

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	projectv1 "github.com/openshift/api/project/v1"
	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	operatorv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

var dedicatedAdminTestName string = "[Suite: informing] [OSD] dedicated-admin permissions"

var _ = ginkgo.Describe(dedicatedAdminTestName, label.Informing, func() {
	ginkgo.Context("dedicated-admin group permissions", func() {
		// list of namespaces to loop through
		namespaceList := []string{
			"openshift-operators",
			"openshift-operators-redhat",
		}

		// setup helper
		h := helper.New()

		// dedicated-admin SA can create projectrequest object
		ginkgo.It("ded-admin SA can create projectrequest", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"system:serviceaccounts:dedicated-admin",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			proj := &projectv1.ProjectRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "osde2e-sample-cust-proj",
				},
				DisplayName: "osde2e-sample-cust-proj",
			}
			_, _ = h.Project().ProjectV1().ProjectRequests().Create(ctx, proj, metav1.CreateOptions{})
		})

		// regular dedicated-admin user can create 'admin' rolebinding
		ginkgo.It("ded-admin user can create 'admin' rolebinding", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			// create dummy user to attach to 'admin'
			user := &userv1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummy-user@redhat.com",
				},
			}

			dummyKindName := "admin"
			_, _ = createRolebinding(ctx, user, dummyKindName, h)
		})

		// regular dedicated-admin user can create 'edit' rolebinding
		ginkgo.It("ded-admin user can create edit rolebinding", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			// create dummy user to attach to 'admin'
			user := &userv1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummy-user@redhat.com",
				},
			}

			dummyKindName := "edit"
			_, _ = createRolebinding(ctx, user, dummyKindName, h)
		})

		// dedicated-admin SA can create 'edit' rolebinding
		ginkgo.It("ded-admin SA can create 'edit' rolebinding", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"system:serviceaccounts:dedicated-admin",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			// create dummy user to attach to 'admin'
			user := &userv1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummy-user@redhat.com",
				},
			}

			dummyKindName := "edit"
			_, _ = createRolebinding(ctx, user, dummyKindName, h)
		})

		// dedicated-admin SA can create 'admin' rolebinding
		ginkgo.It("ded-admin SA can create 'admin' rolebinding", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"system:serviceaccounts:dedicated-admin",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			// create dummy user to attach to 'admin'
			user := &userv1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummy-user@redhat.com",
				},
			}

			dummyKindName := "admin"
			_, _ = createRolebinding(ctx, user, dummyKindName, h)
		})

		// dedicated-admin SA can delete project
		ginkgo.It("ded-admin SA can delete project", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"system:serviceaccounts:dedicated-admin",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			proj := &projectv1.ProjectRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "osde2e-sample-cust-proj",
				},
				DisplayName: "osde2e-sample-cust-proj",
			}
			_ = h.Project().ProjectV1().Projects().Delete(ctx, proj.Name, metav1.DeleteOptions{})
		})

		// dedicated-admin can manage secrets
		// in selected namespaces
		ginkgo.It("ded-admin can manage secrets in selected namespaces", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			manageSecrets(ctx, namespaceList, h)
		})

		// dedicated-admin can manage subscriptions
		// in selected namespaces
		ginkgo.It("ded-admin can manage subscriptions in selected namespaces", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			manageSubscriptions(ctx, namespaceList, h)
		})

		ginkgo.It("dedicated-admin user can patch consoles.operator.openshift.io CR", func(ctx context.Context) {
			// Impersonate ded-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			patchData := []byte(`{"spec":{"plugins":["test"]}}`)

			_, _ = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "operator.openshift.io", Version: "v1",
				Resource: "consoles",
			}).Patch(ctx, "cluster", types.MergePatchType, patchData, metav1.PatchOptions{})
			// revret the changes
			patchEmpty := []byte(`{"spec":{"plugins":[""]}}`)

			_, _ = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "operator.openshift.io", Version: "v1",
				Resource: "consoles",
			}).Patch(ctx, "cluster", types.MergePatchType, patchEmpty, metav1.PatchOptions{})
		})
	})
})

// createRolebinding takes in the desired user and roleRef kindName
// returns the corresponding rolebinding created on cluster
func createRolebinding(
	ctx context.Context,
	user *userv1.User,
	kindName string,
	h *helper.H,
) (*rbacv1.RoleBinding, error) {
	const ns = "osde2e-sample-cust-proj"
	const kind = "ClusterRole"

	rb, err := h.Kube().RbacV1().RoleBindings(ns).Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "osde2e-admin-rolebind",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: rbacv1.UserKind,
				Name: user.Name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     kind,
			Name:     kindName,
		},
	}, metav1.CreateOptions{})
	return rb, err
}

// manageSecrets takes in a list of namespaces
func manageSecrets(ctx context.Context, nsList []string, h *helper.H) {
	for _, ns := range nsList {

		newSecretName := "sample-cust-secret"

		// check 'create' permission
		secrets := h.Kube().CoreV1().Secrets(ns)
		_, err := secrets.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      newSecretName,
				Namespace: ns,
			},
		}, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to create secret %s in namespace %s", newSecretName, ns))

		// check 'get' permission
		dummySecret, err := secrets.Get(ctx, newSecretName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to get secret %s in namespace %s", newSecretName, ns))

		// check 'list' permission
		_, err = secrets.List(ctx, metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to list secret %s in namespace %s", newSecretName, ns))

		// check 'update' permission
		dummySecret.Type = corev1.SecretTypeOpaque
		_, err = secrets.Update(ctx, dummySecret, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to update secret %s in namespace %s", newSecretName, ns))

		// check 'delete' permission
		err = secrets.Delete(ctx, newSecretName, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to delete secret %s in namespace %s", newSecretName, ns))
	}
}

// manageSubscription takes in a list of namespaces
func manageSubscriptions(ctx context.Context, nsList []string, h *helper.H) {
	newSubscriptionName := "sample-cust-subscription"

	for _, ns := range nsList {

		// check 'create permission
		subscriptions := h.Operator().OperatorsV1alpha1().Subscriptions(ns)
		_, err := subscriptions.Create(ctx, &operatorv1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      newSubscriptionName,
				Namespace: ns,
			},
			Spec: &operatorv1.SubscriptionSpec{
				Channel: "alpha",
			},
		}, metav1.CreateOptions{})
		Expect(
			err,
		).NotTo(HaveOccurred(), fmt.Sprintf("failed to create subscription %s in namespace %s", newSubscriptionName, ns))

		// check 'get' permission
		_, err = subscriptions.Get(ctx, newSubscriptionName, metav1.GetOptions{})
		Expect(
			err,
		).NotTo(HaveOccurred(), fmt.Sprintf("failed to get subscription %s in namespace %s", newSubscriptionName, ns))

		// check 'list' permission
		_, err = subscriptions.List(ctx, metav1.ListOptions{})
		Expect(
			err,
		).NotTo(HaveOccurred(), fmt.Sprintf("failed to list subscription %s in namespace %s", newSubscriptionName, ns))

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			sub, err := subscriptions.Get(ctx, newSubscriptionName, metav1.GetOptions{})
			Expect(
				err,
			).NotTo(HaveOccurred(), fmt.Sprintf("failed to get subscription %s in namespace %s", newSubscriptionName, ns))
			// check 'update' permission
			sub.Spec.Channel = "beta"
			_, err = subscriptions.Update(ctx, sub, metav1.UpdateOptions{})
			return err
		})
		Expect(
			err,
		).NotTo(HaveOccurred(), fmt.Sprintf("failed to update subscription %s in namespace %s", newSubscriptionName, ns))

		// check 'delete' permission
		err = subscriptions.Delete(ctx, newSubscriptionName, metav1.DeleteOptions{})
		Expect(
			err,
		).NotTo(HaveOccurred(), fmt.Sprintf("failed to delete subscription %s in namespace %s", newSubscriptionName, ns))
	}
}
