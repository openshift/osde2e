package cloudingress

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// tests
var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	h := helper.New()

	ginkgo.Context("publishingstrategy-public-private", func() {
		ginkgo.It("is a placeholder", func() {
			_, _ = h.Kube().CoreV1().Pods(TestPrefix).Get(context.TODO(), "something", metav1.GetOptions{})

			tests := []struct {
				Name string
			}{}

			for _, test := range tests {
				Expect(test.Name).To(Equal(""))
			}
		})
	})
})

// utils

// common setup and utils are in cloudingress.go
