package cloudingress

import (
	"context"
	"fmt"
	"strings"
	"time"

	"net"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
)

// tests

var _ = ginkgo.Describe(CloudIngressTestName, func() {

	h := helper.New()

	testHostnameResolves(h)
	testCIDRBlockUpdates(h)

})

// utils
func testHostnameResolves(h *helper.H) {
	var err error
	ginkgo.Context("rh-api-test", func() {
		ginkgo.It("hostname should resolve", func() {
			wait.PollImmediate(30*time.Second, 15*time.Minute, func() (bool, error) {

				getOpts := metav1.GetOptions{}
				apiserver, err := h.Cfg().ConfigV1().APIServers().Get(context.TODO(), "cluster", getOpts)
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
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func testCIDRBlockUpdates(h *helper.H) {
	ginkgo.Context("rh-api-test", func() {
		ginkgo.It("cidr block changes should updated the service", func() {

			var APISchemeInstance cloudingressv1alpha1.APIScheme

			//Get the APIScheme
			APISchemeRawData, err := h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
			}).Namespace(CloudIngressNamespace).Get(context.TODO(), "rh-api", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			//structure data into a APIScheme object
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(APISchemeRawData.Object, &APISchemeInstance)
			Expect(err).NotTo(HaveOccurred())

			CIRDBlock := APISchemeInstance.Spec.ManagementAPIServerIngress.AllowedCIDRBlocks

			fmt.Printf("DEBUG: BEFORE TRUNC-> \nType: %T\n Value: %v\n", CIRDBlock, CIRDBlock)
			//remove one IP from the CIDR:
			CIRDBlock[len(CIRDBlock)-1] = ""         // Erase last element (write zero value)
			CIRDBlock = CIRDBlock[:len(CIRDBlock)-1] // Truncate slice

			APISchemeInstance.Spec.ManagementAPIServerIngress.AllowedCIDRBlocks = CIRDBlock

			APISchemeRawData.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&APISchemeInstance)
			Expect(err).NotTo(HaveOccurred())
			//DEBUG
			fmt.Printf("DEBUG: AFTER TRUNC-> \nType: %T\n Value: %v\n", CIRDBlock, CIRDBlock)

			// //Update the APIScheme
			// APISchemeRawData, err = h.Dynamic().Resource(schema.GroupVersionResource{
			// 	Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "apischemes",
			// }).Namespace(CloudIngressNamespace).Update(context.TODO(), APISchemeRawData, metav1.UpdateOptions{})
			// Expect(err).NotTo(HaveOccurred())

			// //check if the service is updated.

			kClient := &client.Client
			rhAPIService := &corev1.Service{}

			ns := types.NamespacedName{
				Namespace: "openshift-kube-apiserver",
				Name:      "rh-api",
			}

			fmt.Println("DEBUG: BEFORE THE PROBLEM")

			err = kClient.Get(context.TODO(), ns, rhAPIService) //problem line
			Expect(err).NotTo(HaveOccurred())

			fmt.Println("DEBUG: AFTER PROBLEM LINE")

			fmt.Printf("Type: %T\nValue: %v\n", rhAPIService, rhAPIService)
		})
	})
}

// common setup and utils are in cloudingress.go
