package verify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"

	v1 "github.com/openshift/api/route/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
)

const (
	consoleNamespace = "openshift-console"
	consoleLabel     = "console"
	oauthNamespace   = "openshift-authentication"
	oauthName        = "oauth-openshift"
)

var routesTestName string = "[Suite: e2e] Routes"

func init() {
	alert.RegisterGinkgoAlert(routesTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(routesTestName, ginkgo.Ordered, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources

	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.It("exist for console and is operational", label.HyperShift, func(ctx context.Context) {
		var consoleRoute v1.Route
		err := client.Get(ctx, consoleLabel, consoleNamespace, &consoleRoute)
		expect.NoError(err)

		for _, ingress := range consoleRoute.Status.Ingress {
			response, err := validateRoute(ctx, ingress)
			Expect(err).NotTo(HaveOccurred(), "failed retrieving %s site", consoleRoute.Name)
			Expect(response.StatusCode).To(Equal(http.StatusOK))
		}
	})

	ginkgo.It("exist for oauth and is operational", func(ctx context.Context) {
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("OAuth route is not deployed to a ROSA hosted-cp cluster")
		}

		var oauthRoute v1.Route
		err := client.Get(ctx, oauthName, oauthNamespace, &oauthRoute)
		expect.NoError(err)

		for _, ingress := range oauthRoute.Status.Ingress {
			response, err := validateRoute(ctx, ingress)
			Expect(err).NotTo(HaveOccurred(), "failed retrieving %s site", oauthRoute.Name)
			Expect(response.StatusCode).To(Equal(http.StatusForbidden))
		}
	})
})

func validateRoute(ctx context.Context, ingress v1.RouteIngress) (http.Response, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	response, err := httpClient.Get(fmt.Sprintf("https://%s", ingress.Host))
	ExpectWithOffset(1, response.Proto).To(Equal("HTTP/1.1"))
	return *response, err
}
