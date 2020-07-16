package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	SRE_PROVIDER_NAME      = "OpenShift_SRE"
	CUSTOMER_PROVIDER_NAME = "CUSTOM"
)

var _ = ginkgo.Describe("[Suite: service-definition] [OSD] user validating webhook", func() {
	h := helper.New()

	ginkgo.Context("user validating webhook", func() {
		ginkgo.It("dedicated admins cannot manage redhat users", func() {
			userName := util.RandomStr(5) + "@redhat.com"
			user, err := createUser(userName, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				deleteUser(user.Name, h)
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
				},
			})
			err = deleteUser(userName, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins cannot manage redhat user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := SRE_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, SRE_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
				Expect(err).NotTo(HaveOccurred())
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
				},
			})
			err = deleteIdentity(idName, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins can manage customer user identity", func() {
			providerUsername := util.RandomStr(5)
			idName := CUSTOMER_PROVIDER_NAME + ":" + util.RandomStr(5)
			identity, err := createIdentity(idName, CUSTOMER_PROVIDER_NAME, providerUsername, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				err = deleteIdentity(identity.Name, h)
				Expect(err).NotTo(HaveOccurred())
			}()
			Expect(err).NotTo(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"dedicated-admins",
				},
			})
			err = deleteIdentity(idName, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})

func createUser(userName string, groups []string, h *helper.H) (*userv1.User, error) {
	user := &userv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName,
		},
		Groups: groups,
	}
	return h.User().UserV1().Users().Create(context.TODO(), user, metav1.CreateOptions{})
}

func deleteUser(userName string, h *helper.H) error {
	return h.User().UserV1().Users().Delete(context.TODO(), userName, metav1.DeleteOptions{})
}

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
