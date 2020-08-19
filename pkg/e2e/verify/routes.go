package verify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

const (
	consoleNamespace = "openshift-console"
	consoleLabel     = "console"
	oauthNamespace   = "openshift-authentication"
	oauthName        = "oauth-openshift"
)

var routesTestName string = "[Suite: e2e] Routes"

func init() {
	alert.RegisterGinkgoAlert(routesTestName, "SD-CICD", "Jeffrey Sica", "#sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(routesTestName, func() {
	h := helper.New()

	ginkgo.It("should be created for Console", func() {
		consoleRoutes(h)
	}, 300)

	ginkgo.It("should be functioning for Console", func() {
		for _, route := range consoleRoutes(h) {
			testRouteIngresses(route, http.StatusOK)
		}
	}, 300)

	ginkgo.It("should be created for oauth", func() {
		oauthRoute(h)
	}, 300)

	ginkgo.It("should be functioning for oauth", func() {
		testRouteIngresses(oauthRoute(h), http.StatusForbidden)
	}, 300)

})

func consoleRoutes(h *helper.H) []v1.Route {
	labelSelector := fmt.Sprintf("app=%s", consoleLabel)
	list, err := h.Route().RouteV1().Routes(consoleNamespace).List(context.TODO(), metav1.ListOptions{
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

func oauthRoute(h *helper.H) v1.Route {
	route, err := h.Route().RouteV1().Routes(oauthNamespace).Get(context.TODO(), oauthName, metav1.GetOptions{})
	if err != nil || route == nil {
		err = fmt.Errorf("failed requesting routes: %v", err)
	}
	Expect(err).NotTo(HaveOccurred(), "failed getting routes for oauth")
	return *route
}

func testRouteIngresses(route v1.Route, status int) {
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
		Expect(resp.StatusCode).To(Equal(status))
		// By default all http request should be protocol HTTP/1.1, see details: https://bugzilla.redhat.com/show_bug.cgi?id=1825354
		Expect(resp.Proto).To(Equal("HTTP/1.1"))
	}
}
