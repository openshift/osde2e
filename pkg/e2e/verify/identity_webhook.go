package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	SRE_PROVIDER_NAME      = "OpenShift_SRE"
	CUSTOMER_PROVIDER_NAME = "CUSTOM"
)

var identityWebhookTestName string = "[Suite: e2e] [OSD] identity validating webhook"

func init() {
	alert.RegisterGinkgoAlert(identityWebhookTestName, "SD-SREP", "Candace Sheremeta", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(identityWebhookTestName, func() {
	h := helper.New()

	ginkgo.Context("identity validating webhook", func() {

		// note: no kube-admin tests since we cannot impersonate kube:admin

		ginkgo.It("system:admin can manage redhat user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := SRE_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, SRE_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("system:admin can manage customer user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := CUSTOMER_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, CUSTOMER_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:admin",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("oauth service account can manage redhat user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := SRE_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, SRE_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:serviceaccount:openshift-authentication:oauth-openshift",
				Groups: []string{
					"system:serviceaccount:openshift-authentication:oauth-openshift",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("oauth service account can manage customer user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := CUSTOMER_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, CUSTOMER_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "system:serviceaccount:openshift-authentication:oauth-openshift",
				Groups: []string{
					"system:serviceaccount:openshift-authentication:oauth-openshift",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// osd-sre-admins cannot manage identities due to RBAC, even though
		// webhook will allow it.
		ginkgo.It("osd-sre-admins cannot manage redhat user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := SRE_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, SRE_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "no-reply@redhat.com",
				Groups: []string{
					"osd-sre-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// osd-sre-admins cannot manage identities due to RBAC, even though
		// webhook will allow it.
		ginkgo.It("osd-sre-admins cannot manage customer user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := CUSTOMER_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, CUSTOMER_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "no-reply@redhat.com",
				Groups: []string{
					"osd-sre-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("osd-sre-cluster-admins can manage redhat user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := SRE_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, SRE_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "no-reply@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("osd-sre-cluster-admins can manage customer user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := CUSTOMER_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, CUSTOMER_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "no-reply@redhat.com",
				Groups: []string{
					"osd-sre-cluster-admins",
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins cannot manage redhat user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := SRE_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, SRE_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
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
			err = deleteIdentity(identity.Name, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins can manage customer user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := CUSTOMER_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, CUSTOMER_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
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
			err = deleteIdentity(identity.Name, h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// random authenticated user cannot manage identities due to RBAC, even
		// though webhook will allow it
		ginkgo.It("random authenticated user cannot manage redhat user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := SRE_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, SRE_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// random authenticated user cannot manage identities due to RBAC, even
		// though webhook will allow it
		ginkgo.It("random authenticated user cannot manage customer user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := CUSTOMER_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, CUSTOMER_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteIdentity(identity.Name, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})

func createIdentity(idName string, providername string, providerUserName string, h *helper.H) (*userv1.Identity, error) {
	identity := &userv1.Identity{
		ObjectMeta: metav1.ObjectMeta{
			Name: idName,
		},
		ProviderName:     providername,
		ProviderUserName: providerUserName,
	}
	return h.User().UserV1().Identities().Create(context.TODO(), identity, metav1.CreateOptions{})
}

func deleteIdentity(idName string, h *helper.H) error {
	return h.User().UserV1().Identities().Delete(context.TODO(), idName, metav1.DeleteOptions{})
}
