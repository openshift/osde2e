package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var storageTestName string = "[Suite: e2e] Storage"

func init() {
	alert.RegisterGinkgoAlert(storageTestName, "SD-SREP", "Christoph Blecker", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(storageTestName, func() {
	h := helper.New()
	ginkgo.It("should be able to be expanded", func() {
		scList, err := h.Kube().StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list StorageClasses")
		Expect(scList).NotTo(BeNil())

		for _, sc := range scList.Items {
			Expect(sc.AllowVolumeExpansion).To(Not(BeNil()))
			Expect(*sc.AllowVolumeExpansion).To(BeTrue())
		}

	}, 300)
})
