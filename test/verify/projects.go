// Package verify provides tests that validate the operation of an OSD cluster.
package verify

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("Projects", func() {
	h := helper.New()

	ginkgo.It("Empty Project should be created", func() {
		_, err := h.Project().ProjectV1().Projects().Get(h.CurrentProject(), metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "project should have been created")
	}, 300)
})
