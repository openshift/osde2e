package verify

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/helper"
)

var _ = ginkgo.Describe("[Suite: e2e] Storage", func() {
	h := helper.New()
	ginkgo.It("should be able to be expanded", func() {
		scList, err := h.Kube().StorageV1().StorageClasses().List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list StorageClasses")
		Expect(scList).NotTo(BeNil())

		for _, sc := range scList.Items {
			Expect(sc.AllowVolumeExpansion).To(Not(BeNil()))
			Expect(*sc.AllowVolumeExpansion).To(BeTrue())
		}

	}, 300)
})
