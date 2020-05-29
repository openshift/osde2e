package operators

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const sharedVolumeCRDName = "sharedvolumes.aws-efs.managed.openshift.io"

var _ = ginkgo.Describe("[Suite: wip] [OSD] AWS EFS Operator", func() {
	h := helper.New()

	ginkgo.Context("Sniff", func() {
		ginkgo.FIt("SharedVolume CRD exists", func() {
			crd, err := h.Dynamic().Resource(
				schema.GroupVersionResource{
					Group:    "apiextensions.k8s.io",
					Resource: "customresourcedefinitions",
					Version:  "v1",
				},
			).Get(sharedVolumeCRDName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.GetName()).Should(Equal(sharedVolumeCRDName))
		})

		ginkgo.FIt("No SharedVolume resources exist", func() {
			sharedVolumes, err := h.Dynamic().Resource(
				schema.GroupVersionResource{
					Group:    "aws-efs.managed.openshift.io",
					Resource: "sharedvolumes",
					Version:  "v1alpha1",
				},
			).List(metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(sharedVolumes.Items)).Should(Equal(0))

		})
	})
})
