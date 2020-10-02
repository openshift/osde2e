package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

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
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteUser(userName, h)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("dedicated admins can manage customer users", func() {
			userName := util.RandomStr(5) + "@customdomain"
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
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			err = deleteUser(userName, h)
			Expect(err).NotTo(HaveOccurred())
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
