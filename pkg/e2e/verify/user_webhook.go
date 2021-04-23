package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var userWebhookTestName string = "[Suite: service-definition] [OSD] user validating webhook"

func init() {
	alert.RegisterGinkgoAlert(userWebhookTestName, "SD-SREP", "Haoran Wang", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(userWebhookTestName, func() {
	h := helper.New()

	ginkgo.Context("user validating webhook", func() {

		// for all tests, "manage" is synonymous with "create/update/delete"

		//system:admin can do whatever it wants

		ginkgo.It("system:admin can manage redhat users with SRE IDP and RH group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("system:admin can manage redhat users with SRE IDP and no group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("system:admin can manage redhat users with other IDP and RH group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"otherIDP:testing_string"}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("system:admin can manage redhat users with other IDP and no group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"otherIDP:testing_string"}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("system:admin can manage customer users with SRE IDP and RH group", func() {
			userName := util.RandomStr(5) + "@customdomain"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("system:admin can manage customer users with SRE IDP and no group", func() {
			userName := util.RandomStr(5) + "@customdomain"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("system:admin can manage customer users with other IDP and RH group", func() {
			userName := util.RandomStr(5) + "@customdomain"
			identities := []string{"otherIDP:testing_string"}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("system:admin can manage customer users with other IDP and no group", func() {
			userName := util.RandomStr(5) + "@customdomain"
			identities := []string{"otherIDP:testing_string"}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// not testing osd-sre-admins because RBAC prevents them from managing
		// users, even though webhook allows them to

		// not testing system:serviceaccount:openshift-authentication:oauth-openshift
		// because RBAC prevents it from managing users, even though webhook
		// allows it to

		// osd-sre-cluster-admins can manage protected RH users as long as the
		// user is in one of the protected groups and is using the SRE IdP
		ginkgo.It("osd-sre-cluster-admins can manage protected redhat users with SRE IDP and RH group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// osd-sre-cluster-admins cannot create/update protected RH users if
		// the user is using the SRE IdP but is not in one of the protected
		// groups
		ginkgo.It("osd-sre-cluster-admins cannot create/update protected redhat users with SRE IDP and no group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// osd-sre-cluster-admins cannot create/update protected RH users if
		// the user is in one of the protected groups but is not using the SRE
		// IdP
		ginkgo.It("osd-sre-cluster-admins cannot create/update protected redhat users with other IDP and RH group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"otherIDP:testing_string"}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// osd-sre-cluster-admins can manage RH users if the user is both not
		// using the SRE IdP and is not in one of the protected groups
		ginkgo.It("osd-sre-cluster-admins can manage non-protected redhat users with other IDP and no group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"otherIDP:testing_string"}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// osd-sre-cluster-admins can manage customer users
		ginkgo.It("osd-sre-cluster-admins can manage customer users with other IDP and no group", func() {
			userName := util.RandomStr(5) + "@customdomain"
			identities := []string{"otherIDP:testing_string"}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admins can only manage non-RH users

		ginkgo.It("dedicated admins cannot manage redhat users with SRE IDP and RH group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins cannot manage redhat users with SRE IDP and no group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins cannot manage redhat users with other IDP and RH group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"otherIDP:testing_string"}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins cannot manage redhat users with other IDP and no group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"otherIDP:testing_string"}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins can manage customer users with other IdP and no group", func() {
			userName := util.RandomStr(5) + "@customdomain"
			identities := []string{"otherIDP:testing_string"}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins can manage customer users with other IdP and customer group", func() {
			userName := util.RandomStr(5) + "@customdomain"
			identities := []string{"otherIDP:testing_string"}
			groups := []string{"dedicated-admins"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// clean-up operations. osd-sre-cluster-admins should be allowed to
		// delete a protected RH user that is either using the SRE IdP and is
		// not in a protected group, or that is in a protected group and is
		// not using the SRE IdP - we do not want these combinations to be
		// possible, so we must have a way to clean them up if we find users
		// who match these combinations (e.g. were created before webhook was
		// put into place)

		ginkgo.It("osd-sre-cluster-admins can delete protected redhat users with SRE IDP and no group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("osd-sre-cluster-admins can delete protected redhat users with other IDP and RH group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"otherIDP:testing_string"}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteUser(user.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// dedicated-admins should not be able to delete these users, however

		ginkgo.It("dedicated-admins cannot delete protected redhat users with SRE IDP and no group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"OpenShift_SRE:" + util.RandomStr(5)}
			user, err := createUser(userName, identities, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteUser(user.Name, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated-admins cannot delete protected redhat users with other IDP and RH group", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			identities := []string{"otherIDP:testing_string"}
			groups := []string{"osd-devaccess"}
			// we need to add the username to the group before we create the user,
			// because the user object cannot be created until the username is in
			// the group
			addUserToGroup(userName, groups[0], h)
			user, err := createUser(userName, identities, groups, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteUser(user.Name, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

	})
})

// createUser creates the given user.
// Note that it may take time for operators to reconcile the permissions of new users,
// so it's best to poll your first attempt to use the resulting user for a couple minutes.
func createUser(userName string, identities []string, groups []string, h *helper.H) (*userv1.User, error) {
	user := &userv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName,
		},
		Identities: identities,
		Groups:     groups,
	}
	return h.User().UserV1().Users().Create(context.TODO(), user, metav1.CreateOptions{})
}

func deleteUser(userName string, h *helper.H) error {
	return h.User().UserV1().Users().Delete(context.TODO(), userName, metav1.DeleteOptions{})
}

// addUserToGroup adds a user to the given group.
// Note that it may take time for operators to reconcile the permissions of new users,
// so it's best to poll your first attempt to use the resulting user for a couple minutes.
func addUserToGroup(userName string, groupName string, h *helper.H) (result *userv1.Group, err error) {
	group, err := h.User().UserV1().Groups().Get(context.TODO(), groupName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	group.Users = append(group.Users, userName)
	return h.User().UserV1().Groups().Update(context.TODO(), group, metav1.UpdateOptions{})
}
