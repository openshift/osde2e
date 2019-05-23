// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

const (
	testsNamespace   = "openshift"
	testsImageStream = "tests"
)

var _ = ginkgo.Describe("OpenShift E2E", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	ginkgo.It("should run until completion", func() {
		// setup permissions
		h.GiveCurrentProjectClusterAdmin()

		// get name of latest test image from ImageStream
		testImageName := ""
		stream, err := h.Image().ImageV1().ImageStreams(testsNamespace).Get(testsImageStream, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, tag := range stream.Spec.Tags {
			if tag.Name == "latest" {
				Expect(tag.From).NotTo(BeNil())
				testImageName = tag.From.Name
			}
		}
		Expect(testImageName).NotTo(BeEmpty(), "no latest tests")

		// run tests inside Pod
		testPod, err := createOpenShiftTestsPod(h, testImageName)
		Expect(err).NotTo(HaveOccurred())

		// get results of test
		gatherResults(h, testPod)
	})
})
