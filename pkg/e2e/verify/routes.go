package verify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
)

var routesTestName string = "[Suite: e2e] Routes"

func init() {
	alert.RegisterGinkgoAlert(routesTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(routesTestName, ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.DescribeTable("exists and is operational", func(ctx context.Context, name, namespace string, statusCode int) {
		if name == "oauth-openshift" && viper.GetBool(config.Hypershift) {
			ginkgo.Skip("OAuth route is not available on a HyperShift cluster")
		}

		err := wait.For(func() (bool, error) {
			deployment := &appsv1.Deployment{}
			err := client.Get(ctx, name, namespace, deployment)
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			for _, cond := range deployment.Status.Conditions {
				if cond.Type == appsv1.DeploymentAvailable && cond.Status == v1.ConditionTrue {
					return true, nil
				}
			}
			return false, nil
		})
		expect.NoError(err, "deployment %s never became available", name)

		err = wait.For(func() (bool, error) {
			err := client.Get(ctx, name, namespace, &routev1.Route{})
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			return true, nil
		})
		expect.NoError(err)

		var route routev1.Route
		err = client.Get(ctx, name, namespace, &route)
		expect.NoError(err)

		ingress := route.Status.Ingress[0]
		Eventually(func(g Gomega) {
			response, err := httpClient.Get(fmt.Sprintf("https://%s", ingress.Host))
			g.Expect(err).To(BeNil())
			g.Expect(response).To(HaveHTTPStatus(statusCode))
			g.Expect(response.Proto).To(Equal("HTTP/1.1"))
		}).Should(Succeed())
	},
		ginkgo.Entry("Console", "console", "openshift-console", http.StatusOK),
		ginkgo.Entry("OAuth", "oauth-openshift", "openshift-authentication", http.StatusForbidden),
	)
})
