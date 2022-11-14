package osd

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	projectv1 "github.com/openshift/api/project/v1"
	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	operatorv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

var dedicatedAdminTestName string = "[Suite: informing] [OSD] dedicated-admin permissions"

func init() {
	alert.RegisterGinkgoAlert(dedicatedAdminTestName, "SD-SREP", "@dedicated-admin-operator", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(dedicatedAdminTestName, func() {
	ginkgo.Context("dedicated-admin group permissions", func() {
		// list of namespaces to loop through
		namespaceList := []string{
			"openshift-operators",
			"openshift-operators-redhat",
		}

		// setup helper
		h := helper.New()

		util.GinkgoIt("cannot add members to cluster-admin", func(ctx context.Context) {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			daGroup, err := h.User().UserV1().Groups().Get(ctx, "dedicated-admins", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			daGroup.Users = append(daGroup.Users, "new-dummy-admin@redhat.com")
			_, err = h.User().UserV1().Groups().Update(ctx, daGroup, metav1.UpdateOptions{})
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		util.GinkgoIt("cannot delete members from cluster-admin", func(ctx context.Context) {
			// add dummy user
			daGroup, err := h.User().UserV1().Groups().Get(ctx, "dedicated-admins", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			daGroup.Users = append(daGroup.Users, "user-to-delete@redhat.com")
			daGroup, err = h.User().UserV1().Groups().Update(ctx, daGroup, metav1.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())

			// remove dummy user as dedicated-admin
			daGroup.Users = []string{}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()
			_, err = h.User().UserV1().Groups().Update(context.TODO(), daGroup, metav1.UpdateOptions{})
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin SA can create projectrequest object
		util.GinkgoIt("ded-admin SA can create projectrequest", func(ctx context.Context) {
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
			_, err := h.Project().ProjectV1().ProjectRequests().Create(ctx, proj, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// regular dedicated-admin user can create 'admin' rolebinding
		util.GinkgoIt("ded-admin user can create 'admin' rolebinding", func(ctx context.Context) {
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

			dummyNs := "osde2e-sample-cust-proj"
			dummyKind := "ClusterRole"
			dummyKindName := "admin"
			_, err := createRolebinding(ctx, dummyNs, user, dummyKind, dummyKindName, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// regular dedicated-admin user can create 'edit' rolebinding
		util.GinkgoIt("ded-admin user can create edit rolebinding", func(ctx context.Context) {
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

			dummyNs := "osde2e-sample-cust-proj"
			dummyKind := "ClusterRole"
			dummyKindName := "edit"
			_, err := createRolebinding(ctx, dummyNs, user, dummyKind, dummyKindName, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin SA can create 'edit' rolebinding
		util.GinkgoIt("ded-admin SA can create 'edit' rolebinding", func(ctx context.Context) {
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

			dummyNs := "osde2e-sample-cust-proj"
			dummyKind := "ClusterRole"
			dummyKindName := "edit"
			_, err := createRolebinding(ctx, dummyNs, user, dummyKind, dummyKindName, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin SA can create 'admin' rolebinding
		util.GinkgoIt("ded-admin SA can create 'admin' rolebinding", func(ctx context.Context) {
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

			dummyNs := "osde2e-sample-cust-proj"
			dummyKind := "ClusterRole"
			dummyKindName := "admin"
			_, err := createRolebinding(ctx, dummyNs, user, dummyKind, dummyKindName, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin SA can delete project
		util.GinkgoIt("ded-admin SA can delete project", func(ctx context.Context) {
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
			err := h.Project().ProjectV1().Projects().Delete(ctx, proj.Name, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin can manage secrets
		// in selected namespaces
		util.GinkgoIt("ded-admin can manage secrets in selected namespaces", func(ctx context.Context) {
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

			err := manageSecrets(ctx, namespaceList, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin can manage subscriptions
		// in selected namespaces
		util.GinkgoIt("ded-admin can manage subscriptions in selected namespaces", func(ctx context.Context) {
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

			err := manageSubscriptions(ctx, namespaceList, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})

// createRolebinding takes in the desired namespace, user, and roleRef kind and kindName
// returns the corresponding rolebinding created on cluster
func createRolebinding(
	ctx context.Context,
	ns string,
	user *userv1.User,
	kind string,
	kindName string,
	h *helper.H,
) (*rbacv1.RoleBinding, error) {
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
// and returns error if an action fails
func manageSecrets(ctx context.Context, nsList []string, h *helper.H) error {
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
	return nil
}

// manageSubscription takes in a list of namespaces
// and returns error if an action fails
func manageSubscriptions(ctx context.Context, nsList []string, h *helper.H) error {
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
	return nil
}
