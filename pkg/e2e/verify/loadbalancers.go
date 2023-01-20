package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

const (
	routerIngressLoadBalancerNamespace = "openshift-ingress"
	routerIngressLoadBalancer          = "router-default"
	externalLoadBalancerNamespace      = "openshift-kube-apiserver"
	externalLoadBalancer               = "rh-api"
)

var loadBalancersTestName string = "[Suite: informing] Load Balancers"

func init() {
	alert.RegisterGinkgoAlert(loadBalancersTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(loadBalancersTestName, ginkgo.Ordered, label.Informing, func() {
	var h *helper.H
	var client *resources.Resources
	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.DescribeTable("should exist", func(ctx context.Context, name, namespace string) {
		if name == externalLoadBalancer && (viper.GetBool(config.Hypershift) || viper.GetBool(rosaprovider.STS)) {
			ginkgo.Skip("rh-api load balancer is not deployed to ROSA or HyperShift clusters")
		}

		service := &v1.Service{}
		expect.NoError(client.Get(ctx, name, namespace, service))
		Expect(service.Spec.Type).To(Equal(v1.ServiceTypeLoadBalancer), "expected a load balancer service but got %s", service.Spec.Type)

	},
		ginkgo.Entry("router-default", routerIngressLoadBalancer, routerIngressLoadBalancerNamespace),
		ginkgo.Entry("rh-api", externalLoadBalancer, externalLoadBalancerNamespace),
	)
})
