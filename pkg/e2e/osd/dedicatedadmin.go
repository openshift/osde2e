package osd

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var dedicatedAdminTestName string = "[Suite: informing] [OSD] dedicated-admin permissions"

func init() {
	alert.RegisterGinkgoAlert(dedicatedAdminTestName, "SD-SREP", "Matt Bargenquast", "#sd-cicd-alerts", "sd-cicd@redhat.com", 4)
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
	})
})
