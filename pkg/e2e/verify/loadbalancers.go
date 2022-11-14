package verify

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
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

var _ = ginkgo.Describe(loadBalancersTestName, func() {
	h := helper.New()

	util.GinkgoIt("router/ingress load balancer should exist", func(ctx context.Context) {
		exists, err := loadBalancerExists(ctx, h, routerIngressLoadBalancerNamespace, routerIngressLoadBalancer)
		Expect(err).ToNot(HaveOccurred(), "an error should not have occurred when looking for the load balancer")
		Expect(exists).To(BeTrue(), "the load balancer should exist")
	}, 10)

	util.GinkgoIt("external load balancer should exist", func(ctx context.Context) {
		exists, err := loadBalancerExists(ctx, h, externalLoadBalancerNamespace, externalLoadBalancer)
		Expect(err).ToNot(HaveOccurred(), "an error should not have occurred when looking for the load balancer")
		Expect(exists).To(BeTrue(), "the load balancer should exist")
	}, 10)
})

func loadBalancerExists(ctx context.Context, h *helper.H, namespace string, loadBalancer string) (bool, error) {
	service, err := h.Kube().CoreV1().Services(namespace).Get(ctx, loadBalancer, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("error getting loadbalancer: %v", err)
	}

	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		return false, fmt.Errorf("namespace %s, service %s is not a load balancer, but is type %v", namespace, loadBalancer, service.Spec.Type)
	}

	return true, nil
}
