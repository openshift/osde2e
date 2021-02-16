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

var hstsTestName string = "[Suite: informing] HTTP Strict Transport Security"

var _ = ginkgo.Describe(hstsTestName, func() {
	h := helper.New()
	namespaces := [2]string{"openshift-monitoring", "openshift-console"}

	ginkgo.Context("Validating HTTP strict transport security", func() {
		for _, namespace := range namespaces {
			ginkgo.It("should be set for openshift-monitoring OSD managed routes", func() {
				foundKey, err := hstsManagedRoutes(h, namespace)
				Expect(err).NotTo(HaveOccurred(), "failed getting routes for %v", namespace)
				Expect(foundKey).Should(BeTrue(), "%v namespace routes have HTTP strict transport security set", namespace)
			}, 5)
		}
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
