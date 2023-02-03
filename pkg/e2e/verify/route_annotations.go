package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"k8s.io/kubectl/pkg/util/slice"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

var routeAnnotationsTestName = "Route Annotations"

var _ = ginkgo.Describe(routeAnnotationsTestName+" Samesite Cookie Strict", ginkgo.Ordered, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	ginkgo.BeforeAll(func() {
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("Route Annotations are not configured on a HyperShift cluster")
		}
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.DescribeTable("should be set for routes in namespace", func(ctx context.Context, namespace string, ignoreRoutes ...string) {
		const annotationName = "router.openshift.io/cookie-same-site"

		routeList := &routev1.RouteList{}
		expect.NoError(client.WithNamespace(namespace).List(ctx, routeList))

		for _, route := range routeList.Items {
			if slice.ContainsString(ignoreRoutes, route.GetName(), nil) {
				continue
			}
			annotationValue, ok := route.GetAnnotations()[annotationName]
			Expect(ok).To(BeTrue(), "route %s did not contain %s annotation", route.GetName(), annotationName)
			Expect(annotationValue).To(ContainSubstring("Strict"))
		}
	},
		ginkgo.Entry("openshift-monitoring", "openshift-monitoring", "prometheus-k8s-federate"),
		ginkgo.Entry("openshift-console", "openshift-console", "downloads"),
	)
})

var _ = ginkgo.Describe(routeAnnotationsTestName+" HTTP Strict Transport Security", ginkgo.Ordered, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	ginkgo.BeforeAll(func() {
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("Route Annotations are not configured on a HyperShift cluster")
		}
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.DescribeTable("should be set for routes in namespace", func(ctx context.Context, namespace string, ignoreRoutes ...string) {
		const annotationName = "haproxy.router.openshift.io/hsts_header"

		routeList := &routev1.RouteList{}
		expect.NoError(client.WithNamespace(namespace).List(ctx, routeList))

		for _, route := range routeList.Items {
			if slice.ContainsString(ignoreRoutes, route.GetName(), nil) {
				continue
			}
			annotationValue, ok := route.GetAnnotations()[annotationName]
			Expect(ok).To(BeTrue(), "route %s did not contain %s annotation", route.GetName(), annotationName)
			Expect(annotationValue).To(ContainSubstring("max-age=31536000;preload"))
		}
	},
		ginkgo.Entry("openshift-monitoring", "openshift-monitoring", "prometheus-k8s-federate"),
		ginkgo.Entry("openshift-console", "openshift-console", "downloads"),
	)
})
