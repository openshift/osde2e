package verify

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

const (
	consoleNamespace = "openshift-console"
	consoleLabel     = "console"
)

var _ = ginkgo.Describe("[Suite: e2e] Routes", func() {
	h := helper.New()

	ginkgo.It("should be created for Console", func() {
		consoleRoutes(h)
	}, 300)

	ginkgo.It("should be functioning for Console", func() {
		for _, route := range consoleRoutes(h) {
			testRouteIngresses(route)
		}
	}, 300)
})

func consoleRoutes(h *helper.H) []v1.Route {
	labelSelector := fmt.Sprintf("app=%s", consoleLabel)
	list, err := h.Route().RouteV1().Routes(consoleNamespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil || list == nil {
		err = fmt.Errorf("failed requesting routes: %v", err)
	} else if len(list.Items) == 0 {
		err = fmt.Errorf("no routes matching '%s' in namespace '%s'", labelSelector, consoleNamespace)
	}

	Expect(err).NotTo(HaveOccurred(), "failed getting routes for console")
	return list.Items
}

func testRouteIngresses(route v1.Route) {
	Expect(route.Status.Ingress).ShouldNot(HaveLen(0),
		"no ingresses have been setup for the route '%s/%s'", route.Namespace, route.Name)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	for _, ingress := range route.Status.Ingress {
		consoleURL := fmt.Sprintf("https://%s", ingress.Host)

		resp, err := client.Get(consoleURL)
		Expect(err).NotTo(HaveOccurred(), "failed retrieving Console site")
		Expect(resp).NotTo(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	}
}
