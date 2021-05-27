package cloudingress

import (
	"context"
	"reflect"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudingressv1alpha1 "github.com/openshift/cloud-ingress-operator/pkg/apis/cloudingress/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// tests

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	h := helper.New()

	testRhSSHCIDRBlockUpdates(h)
	testDroppingSSHDService(h)
})

// testDroppingSSHDService deletes the rh-ssh service, then waits for the cloudingressoperator to recreate it
func testDroppingSSHDService(h *helper.H) {
	ginkgo.Context("rh-ssh service", func() {
		ginkgo.It("should be recreated after deletion", func() {
			err := h.Kube().CoreV1().Services("openshift-sre-sshd").Delete(context.TODO(), "rh-ssh", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			// wait 30 secs for sshd_controller to reconcile
			time.Sleep(30 * time.Second)

			_, err = h.Kube().CoreV1().Services("openshift-sre-sshd").Get(context.TODO(), "rh-ssh", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

// testRhSSHCIDRBlockUpdates compares the CIRDBlock on the related SSHD custom resource and the service
// after an update to make sure changes to the service is updated according to changes on the CR
func testRhSSHCIDRBlockUpdates(h *helper.H) {
	ginkgo.Context("rh-ssh-test", func() {
		ginkgo.It("cidr block changes should updated the service", func() {

			//Create SSHD Object
			var SSHDInstance cloudingressv1alpha1.SSHD

			//Get the SSHD CR
			SSHDRawData, err := h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "sshds",
			}).Namespace("openshift-sre-sshd").Get(context.TODO(), "rh-ssh", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			//structure the SSHD unstructured data into a SSHD object
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(SSHDRawData.Object, &SSHDInstance)
			Expect(err).NotTo(HaveOccurred())

			//Extract the CIDRblock into its own var for ease of use and readability
			CIDRBlock := SSHDInstance.Spec.AllowedCIDRBlocks

			//remove last IP from the CIDRBlock:
			CIDRBlock[len(CIDRBlock)-1] = ""         // Erase last element (write zero value)
			CIDRBlock = CIDRBlock[:len(CIDRBlock)-1] // Truncate slice

			//Put the new CIRDBlock ranges into the SSHD
			SSHDInstance.Spec.AllowedCIDRBlocks = CIDRBlock

			//Unstructure the Data in order to be usable for the update of the CR
			SSHDRawData.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&SSHDInstance)
			Expect(err).NotTo(HaveOccurred())

			//Update the SSHD
			SSHDRawData, err = h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "cloudingress.managed.openshift.io", Version: "v1alpha1", Resource: "sshds",
			}).Namespace("openshift-sre-sshd").Update(context.TODO(), SSHDRawData, metav1.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())

			//Create a service Object
			rhSSHService := &corev1.Service{}

			// wait 30 secs for apiserver to reconcile
			time.Sleep(30 * time.Second)

			//Extract the LoadBalancerSourceRanges from the service
			rhSSHService, err = h.Kube().CoreV1().Services("openshift-sre-sshd").Get(context.TODO(), "rh-ssh", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			//Make sure both the New CIDRBlock and the Service LoadBalancerSourceRanges are equal
			//If they are then the SSHD update also updated the service.
			res := reflect.DeepEqual(CIDRBlock, rhSSHService.Spec.LoadBalancerSourceRanges)
			Expect(res).Should(BeTrue())

		})
	})
}
