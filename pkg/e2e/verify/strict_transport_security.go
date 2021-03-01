package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var hstsTestName string = "[Suite: e2e] [OSD] HTTP Strict Transport Security"

var _ = ginkgo.Describe(hstsTestName, func() {
	h := helper.New()

	consoleNamespace := "openshift-console"
	monitorNamespace := "openshift-monitoring"

	ginkgo.Context("Validating HTTP strict transport security", func() {
		ginkgo.It("should be set for openshift-console OSD managed routes", func() {
			foundKey, err := hstsManagedRoutes(h, consoleNamespace)
			Expect(err).NotTo(HaveOccurred(), "failed getting routes for %v", consoleNamespace)
			Expect(foundKey).Should(BeTrue(), "%v namespace routes have HTTP strict transport security set", consoleNamespace)
		}, 5)
		ginkgo.It("should be set for openshift-monitoring OSD managed routes", func() {
			foundKey, err := hstsManagedRoutes(h, monitorNamespace)
			Expect(err).NotTo(HaveOccurred(), "failed getting routes for %v", monitorNamespace)
			Expect(foundKey).Should(BeTrue(), "%v namespace routes have HTTP strict transport security set", monitorNamespace)
		}, 5)
	})
})

func hstsManagedRoutes(h *helper.H, namespace string) (bool, error) {
	route, err := h.Route().RouteV1().Routes(namespace).List(context.TODO(), metav1.ListOptions{})
	hstsExists := false
	annotationHsts := "hsts_header"
	hstsSetting := "max-age=31536000;preload"

	if err != nil || route == nil {
		return false, fmt.Errorf("failed requesting routes: %v", err)
	}

	for key, value := range route.Items[0].Annotations {
		if strings.Contains(key, annotationHsts) && strings.Contains(value, hstsSetting) {
			hstsExists = true
		}
	}
	return hstsExists, nil
}
