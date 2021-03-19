package verify

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/api/config/v1"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	annotationName  = "cookie-same-site"
	samesiteSetting = "Strict"
	monNamespace    = "openshift-monitoring"
	conNamespace    = "openshift-console"
	supportVersion  = 46 // samesite cookie is only supported on >= v4.6.x
)

var samesiteTestName string = "[Suite: e2e] [OSD] Samesite Cookie Strict"

var _ = ginkgo.Describe(samesiteTestName, func() {
	h := helper.New()

	ginkgo.Context("Validating samesite cookie", func() {

		ginkgo.It("should be set for openshift-monitoring OSD managed routes", func() {
			clusterVersion, majMinVersion, err := getClusterVersion(h)
			Expect(err).NotTo(HaveOccurred(), "failed getting cluster version")
			Expect(clusterVersion).NotTo(BeNil())

			if majMinVersion < supportVersion {
				ginkgo.Skip("skipping due to unsupported cluster version. Must be >=4.6.0")
			}

			foundKey, err := managedRoutes(h, monNamespace)
			Expect(err).NotTo(HaveOccurred(), "failed getting routes for %v", monNamespace)
			Expect(foundKey).Should(BeTrue(), "%v namespace routes have samesite cookie set", monNamespace)
		}, 5)

		ginkgo.It("should be set for openshift-console OSD managed routes", func() {
			clusterVersion, majMinVersion, err := getClusterVersion(h)
			Expect(err).NotTo(HaveOccurred(), "failed getting cluster version")
			Expect(clusterVersion).NotTo(BeNil())

			if majMinVersion < supportVersion {
				ginkgo.Skip("skipping due to unsupported cluster version. Must be >=4.6.0")
			}

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

func getClusterVersion(h *helper.H) (*v1.ClusterVersion, int, error) {
	cfgClient := h.Cfg()
	getOpts := metav1.GetOptions{}
	clusterVersion, err := cfgClient.ConfigV1().ClusterVersions().Get(context.TODO(), "version", getOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("couldn't get current ClusterVersion '%s': %v", "version", err)
	}
	// Get the cluster version and slice it, then convert the major/minor version to int Ex. majMinVersion := 46
	splitVersion := strings.Split(clusterVersion.Status.Desired.Version, ".")
	majMinVersion, err := strconv.Atoi(splitVersion[0] + splitVersion[1])
	Expect(err).NotTo(HaveOccurred(), "failed normalizing major/minor version to integer %v", err)

	return clusterVersion, majMinVersion, nil
}
