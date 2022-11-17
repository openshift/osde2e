package cloudingress

import (
	"context"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/util/wait"
)

// tests

var _ = ginkgo.Describe(constants.SuiteOperators+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
	})

	h := helper.New()

	testHostnameResolves(h)
	testCIDRBlockUpdates(h)
})

// testHostnameResolves Confirms hostname on the cluster resolves
func testHostnameResolves(h *helper.H) {
	var err error

	hostnameResolvePollDuration := 15 * time.Minute
	ginkgo.Context("rh-api-test", func() {
		util.GinkgoIt("hostname should resolve", func(ctx context.Context) {
			wait.PollImmediate(30*time.Second, hostnameResolvePollDuration, func() (bool, error) {
				getOpts := metav1.GetOptions{}
				apiserver, err := h.Cfg().ConfigV1().APIServers().Get(ctx, "cluster", getOpts)
				if err != nil {
					return false, err
				}
				if len(apiserver.Spec.ServingCerts.NamedCertificates) < 1 {
					return false, nil
				}

				for _, namedCert := range apiserver.Spec.ServingCerts.NamedCertificates {
					for _, name := range namedCert.Names {
						if strings.HasPrefix("rh-api", name) {
							_, err := net.LookupHost(name)
							if err != nil {
								return false, err
							}
						}
					}
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, (hostnameResolvePollDuration + 1*time.Minute).Seconds())
	})
}

// testCIDRBlockUpdates compares the CIRDBlock on the related apischeme and the service
// after an update to make sure changes to the apischem
func testCIDRBlockUpdates(h *helper.H) {
	ginkgo.Context("rh-api-test", func() {
		util.GinkgoIt("cidr block changes should updated the service", func(ctx context.Context) {
			// Create APISScheme Object
			var APISchemeInstance cloudingressv1alpha1.APIScheme

			// Get the APIScheme
			APISchemeRawData, err := h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
			}).Namespace(OperatorNamespace).Get(ctx, apiSchemeResourceName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			// structure the APIScheme unstructured data into a APIScheme object
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(APISchemeRawData.Object, &APISchemeInstance)
			Expect(err).NotTo(HaveOccurred())

			// Extract the CIDRblock into its own var for ease of use and readability
			CIDRBlock := APISchemeInstance.Spec.ManagementAPIServerIngress.AllowedCIDRBlocks

			// remove last IP from the CIDRBlock:
			CIDRBlock[len(CIDRBlock)-1] = ""         // Erase last element (write zero value)
			CIDRBlock = CIDRBlock[:len(CIDRBlock)-1] // Truncate slice

			// Put the new CIRDBlock ranges into the APIScheme
			APISchemeInstance.Spec.ManagementAPIServerIngress.AllowedCIDRBlocks = CIDRBlock

			// Unstructure the Data in order to be usable for the update of the CR
			APISchemeRawData.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&APISchemeInstance)
			Expect(err).NotTo(HaveOccurred())

			// //Update the APIScheme
			APISchemeRawData, err = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
			}).Namespace(OperatorNamespace).Update(ctx, APISchemeRawData, metav1.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())

			// Create a service Object
			var rhAPIService *corev1.Service

			// wait 30 secs for apiserver to reconcile
			time.Sleep(30 * time.Second)

			// Extract the LoadBalancerSourceRanges from the service
			rhAPIService, err = h.Kube().
				CoreV1().
				Services("openshift-kube-apiserver").
				Get(ctx, apiSchemeResourceName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			// Make sure both the New CIDRBlock and the Service LoadBalancerSourceRanges are equal
			// If they are then the APIScheme update also updated the service.
			res := reflect.DeepEqual(CIDRBlock, rhAPIService.Spec.LoadBalancerSourceRanges)
			Expect(res).Should(BeTrue())
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})
}

// utils
// common setup and utils are in cloudingress.go
