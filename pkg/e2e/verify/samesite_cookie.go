package verify

import (
	"context"
	"fmt"
	"strconv"
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
	// samesite cookie is only supported on >= v4.6.x
	supportMajorVersion = 4
	supportMinorVersion = 6
)

var samesiteTestName string = "[Suite: e2e] [OSD] Samesite Cookie Strict"

var _ = ginkgo.Describe(samesiteTestName, func() {
	h := helper.New()

	ginkgo.Context("Validating samesite cookie", func() {

		supportedVersion := verifyVersionSupport(h)

		ginkgo.It("should be set for openshift-monitoring OSD managed routes", func() {
			if supportedVersion() {
				foundKey, err := managedRoutes(h, monNamespace)
				Expect(err).NotTo(HaveOccurred(), "failed getting routes for %v", monNamespace)
				Expect(foundKey).Should(BeTrue(), "%v namespace routes have samesite cookie set", monNamespace)
			} else {
				ginkgo.Skip("skipping due to unsupported cluster version. Must be >=4.6.0")
			}
		}, 5)

		ginkgo.It("should be set for openshift-console OSD managed routes", func() {
			if supportedVersion() {
				foundKey, err := managedRoutes(h, conNamespace)
				Expect(err).NotTo(HaveOccurred(), "failed getting routes for %v", conNamespace)
				Expect(foundKey).Should(BeTrue(), "%v namespace routes have samesite cookie set", conNamespace)
			} else {
				ginkgo.Skip("skipping due to unsupported cluster version. Must be >=4.6.0")
			}
		}, 5)
	})
})

func verifyVersionSupport(h *helper.H) func() bool {
	return func() bool {
		clusterVersionObj, err := h.GetClusterVersion()
		Expect(err).NotTo(HaveOccurred(), "failed getting cluster version")
		Expect(clusterVersionObj).NotTo(BeNil())

		splitVersion := strings.Split(clusterVersionObj.Status.Desired.Version, ".")

		// check that semver is 3 elements e.g. 4.6.0, but we only need to verify major/minor version
		if len(splitVersion) == 3 {
			majorVersion, err := strconv.Atoi(splitVersion[0])
			Expect(err).NotTo(HaveOccurred(), "failed getting major version %v. Error: %v", majorVersion, err)
			minorVersion, err := strconv.Atoi(splitVersion[1])
			Expect(err).NotTo(HaveOccurred(), "failed getting minor version %v. Error: %v", minorVersion, err)

			if majorVersion < supportMajorVersion {
				return false
			} else if majorVersion == supportMajorVersion && minorVersion < supportMinorVersion {
				return false
			} else {
				return true
			}
		} else {
			// semver not in correct format if anything other than 3 elements exist
			return false
		}
	}
}

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
