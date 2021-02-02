package cloudingress

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// tests
var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	h := helper.New()

	ginkgo.Context("publishingstrategy-public-private", func() {
		ginkgo.It("should be able to toggle the default applicationingress from public to private", func() {

			updateApplicationIngress(h, "internal")

			// Wait for the router-default service to have the correct annotation
			err := pollAppIngressService(h, "router-default", "openshift-ingress", "private")
			Expect(err).NotTo(HaveOccurred())
		})
		ginkgo.It("should be able to toggle the default applicationingress from private to public", func() {
			updateApplicationIngress(h, "external")

			// Wait for the router-default service to have the correct annotation
			err := pollAppIngressService(h, "router-default", "openshift-ingress", "public")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// utils

func updateApplicationIngress(h *helper.H, lbscheme string) {
	var PublishingStrategyInstance cloudingressv1alpha1.PublishingStrategy

	// Check that the PublishingStrategy CR is present
	ps, err := h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Get(context.TODO(), "publishingstrategy", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ps.Object, &PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Grab the current list of Application Ingresses from the Publishing Strategy
	AppIngress := PublishingStrategyInstance.Spec.ApplicationIngress

	// Find the default router and update its scheme
	for i, v := range AppIngress {
		if v.Default == true {
			AppIngress[i].Listening = cloudingressv1alpha1.Listening(lbscheme)
		}
	}

	PublishingStrategyInstance.Spec.ApplicationIngress = AppIngress

	ps.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&PublishingStrategyInstance)
	Expect(err).NotTo(HaveOccurred())

	// Update the publishingstrategy
	ps, err = h.Dynamic().Resource(schema.GroupVersionResource{Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "publishingstrategies"}).Namespace(OperatorNamespace).Update(context.TODO(), ps, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

func pollAppIngressService(h *helper.H, serviceName string, svcNamespace string, lbscheme string) error {
	// pollAppIngressService will check for the existence of a service
	// in the specified project with the specified load balancer scheme
	// and wait for it to exist until a timeout

	var err error
	// interval is the duration in seconds between polls
	// values here for humans

	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		service, err := h.Kube().CoreV1().Services(svcNamespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			if lbscheme == "private" {
				if _, ok := service.Annotations["service.beta.kubernetes.io/aws-load-balancer-internal"]; ok == true {
					log.Printf("%s service switched a %v LoadBalancer scheme after polling for %v", serviceName, lbscheme, elapsed)
					break Loop
				}
			}
			if lbscheme == "public" {
				//Success
				log.Printf("%s service switched a %v LoadBalancer scheme after polling for %v", serviceName, lbscheme, elapsed)
				break Loop
			}
		case strings.Contains(err.Error(), "forbidden"):
			return err
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s service to exist with a %v LoadBalancer scheme", (timeoutDuration - elapsed), serviceName, lbscheme)
				time.Sleep(intervalDuration)
			} else {
				err = fmt.Errorf("Failed to get service %s before timeout", serviceName)
				break Loop
			}
		}
	}
	return err
}

// common setup and utils are in cloudingress.go
