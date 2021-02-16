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

const (
	annotationName  = "cookie-same-site"
	samesiteSetting = "Strict"
	monNamespace    = "openshift-monitoring"
	conNamespace    = "openshift-console"
)

var samesiteTestName string = "[Suite: informing] Samesite Cookie Strict"

var _ = ginkgo.Describe(samesiteTestName, func() {
	h := helper.New()

	ginkgo.Context("Validating samesite cookie", func() {
		ginkgo.It("should be set for openshift-monitoring OSD managed routes", func() {
			foundKey, err := managedRoutes(h, monNamespace)
			Expect(err).NotTo(HaveOccurred(), "failed getting routes for %v", monNamespace)
			Expect(foundKey).Should(BeTrue(), "%v namespace routes have samesite cookie set", monNamespace)
		}, 5)

		ginkgo.It("should be set for openshift-console OSD managed routes", func() {
			foundKey, err := managedRoutes(h, conNamespace)
			Expect(err).NotTo(HaveOccurred(), "failed getting routes for %v", conNamespace)
			Expect(foundKey).Should(BeTrue(), "%v namespace routes have samesite cookie set", conNamespace)
		}, 5)
	})
})

func managedRoutes(h *helper.H, namespace string) (bool, error) {
	route, err := h.Route().RouteV1().Routes(namespace).List(context.TODO(), metav1.ListOptions{})
	samesiteExists := false
	if err != nil || route == nil {
		return false, fmt.Errorf("failed requesting routes: %v", err)
	}

	for key, value := range route.Items[0].Annotations {
		if strings.Contains(key, annotationName) && strings.Contains(value, samesiteSetting) {
			samesiteExists = true
		}
	}
	return samesiteExists, nil
}
