package verify

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/onsi/ginkgo"

	"github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	consoleNamespace = "openshift-console"
	consoleLabel     = "console"
)

var _ = ginkgo.Describe("Routes", func() {
	defer ginkgo.GinkgoRecover()
	_, cluster := NewCluster()

	ginkgo.It("should be created for Console", func() {
		consoleRoutes(cluster)
	})

	ginkgo.It("should be functioning for Console", func() {
		for _, route := range consoleRoutes(cluster) {
			testRouteIngresses(route)
		}
	})
})

func consoleRoutes(cluster *Cluster) []v1.Route {
	labelSelector := fmt.Sprintf("app=%s", consoleLabel)
	list, err := cluster.Route().RouteV1().Routes(consoleNamespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil || list == nil {
		err = fmt.Errorf("failed requesting routes: %v", err)
	} else if len(list.Items) == 0 {
		err = fmt.Errorf("no routes matching '%s' in namespace '%s'", labelSelector, consoleNamespace)
	}

	if err != nil {
		ginkgo.Fail("Failed getting routes for console: " + err.Error())
	}
	return list.Items
}

func testRouteIngresses(route v1.Route) {
	if len(route.Status.Ingress) == 0 {
		msg := fmt.Sprintf("no ingresses have been setup for the route '%s/%s'", route.Namespace, route.Name)
		ginkgo.Fail(msg)
	}

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
		if err != nil {
			err = fmt.Errorf("failed retrieving Console site: %v", err)
		} else if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("expected status code '%d' but got '%d' instead", http.StatusOK, resp.StatusCode)
		}

		if err != nil {
			ginkgo.Fail("Failed retrieving Console: " + err.Error())
		}
	}
}
