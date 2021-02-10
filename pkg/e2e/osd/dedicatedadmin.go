package osd

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	projectv1 "github.com/openshift/api/project/v1"
	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/spf13/viper"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var dedicatedAdminTestName string = "[Suite: informing] [OSD] dedicated-admin permissions"

func init() {
	alert.RegisterGinkgoAlert(dedicatedAdminTestName, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(dedicatedAdminTestName, func() {
	ginkgo.Context("dedicated-admin group permissions", func() {

		// setup helper
		h := helper.New()

		ginkgo.It("cannot add members to cluster-admin", func() {

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			daGroup, err := h.User().UserV1().Groups().Get(context.TODO(), "dedicated-admins", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			daGroup.Users = append(daGroup.Users, "new-dummy-admin@redhat.com")
			_, err = h.User().UserV1().Groups().Update(context.TODO(), daGroup, metav1.UpdateOptions{})
			Expect(err).To(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("cannot delete members from cluster-admin", func() {

			// add dummy user
			daGroup, err := h.User().UserV1().Groups().Get(context.TODO(), "dedicated-admins", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			daGroup.Users = append(daGroup.Users, "user-to-delete@redhat.com")
			daGroup, err = h.User().UserV1().Groups().Update(context.TODO(), daGroup, metav1.UpdateOptions{})
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
		ginkgo.It("ded-admin SA can create projectrequest", func() {

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
			_, err := h.Project().ProjectV1().ProjectRequests().Create(context.TODO(), proj, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// regular dedicated-admin user can create 'admin' rolebinding
		ginkgo.It("ded-admin user can create 'admin' rolebinding", func() {

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
			_, err := createRolebinding(dummyNs, user, dummyKind, dummyKindName, h)
			Expect(err).NotTo(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// regular dedicated-admin user can create 'edit' rolebinding
		ginkgo.It("ded-admin user can create edit rolebinding", func() {

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
			_, err := createRolebinding(dummyNs, user, dummyKind, dummyKindName, h)
			Expect(err).NotTo(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin SA can create 'edit' rolebinding
		ginkgo.It("ded-admin SA can create 'edit' rolebinding", func() {

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
			_, err := createRolebinding(dummyNs, user, dummyKind, dummyKindName, h)
			Expect(err).NotTo(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin SA can create 'admin' rolebinding
		ginkgo.It("ded-admin SA can create 'admin' rolebinding", func() {

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
			_, err := createRolebinding(dummyNs, user, dummyKind, dummyKindName, h)
			Expect(err).NotTo(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admin SA can delete project
		ginkgo.It("ded-admin SA can delete project", func() {

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
			err := h.Project().ProjectV1().Projects().Delete(context.TODO(), proj.Name, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

	})
})

// createRolebinding takes in the desired namespace, user, and roleRef kind and kindName
// returns the corresponding rolebinding created on cluster
func createRolebinding(ns string, user *userv1.User, kind string, kindName string, h *helper.H) (*rbacv1.RoleBinding, error) {

	rb, err := h.Kube().RbacV1().RoleBindings(ns).Create(context.TODO(), &rbacv1.RoleBinding{
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
