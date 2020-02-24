package operators

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstruct "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = ginkgo.Describe("[Suite: operators] [OSD] Curator Operator", func() {
	h := helper.New()
	ginkgo.Context("operator source should be curated", func() {

		ginkgo.It("we should use curated operator source", func() {
			listOpts := metav1.ListOptions{}
			rList, err := h.Dynamic().Resource(schema.GroupVersionResource{
				Group:    "operators.coreos.com",
				Version:  "v1",
				Resource: "operatorsources",
			}).List(listOpts)

			Expect(err).NotTo(HaveOccurred())

			for _, os := range rList.Items {
				name := os.GetName()
				Expect(name).Should(HavePrefix("osd-curated"))
				registryNamespace, ok, err := unstruct.NestedString(os.Object, "spec", "registryNamespace")
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).Should(Equal(true))
				Expect(registryNamespace).Should(HavePrefix("curated"))
			}
		}, float64(config.Instance.Tests.PollingTimeout))

	})
})
