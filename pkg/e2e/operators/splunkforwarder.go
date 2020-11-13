package operators

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	sfv1alpha1 "github.com/openshift/splunk-forwarder-operator/pkg/apis/splunkforwarder/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var splunkForwarderBlocking string = "[Suite: operators] [OSD] Splunk Forwarder Operator"

func init() {
	alert.RegisterGinkgoAlert(splunkForwarderBlocking, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(splunkForwarderBlocking, func() {

	var operatorName = "splunk-forwarder-operator"
	var operatorNamespace string = "openshift-splunk-forwarder-operator"
	var operatorLockFile string = "splunk-forwarder-operator-lock"
	var defaultDesiredReplicas int32 = 1
	var clusterRoleBindings = []string{
		"splunk-forwarder-operator-clusterrolebinding",
	}

	var clusterRoles = []string{
		"splunk-forwarder-operator",
		"splunk-forwarder-operator-og-admin",
		"splunk-forwarder-operator-og-edit",
		"splunk-forwarder-operator-og-view",
	}

	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkClusterRoleBindings(h, clusterRoleBindings, false)
	checkClusterRoles(h, clusterRoles, false)
	checkUpgrade(helper.New(), "openshift-splunk-forwarder-operator",
		"openshift-splunk-forwarder-operator", "splunk-forwarder-operator",
		"splunk-forwarder-operator-catalog")

	ginkgo.Context("splunkforwarders", func() {
		ginkgo.It("dedicated admin should not be able to manage splunkforwarders CR", func() {
			sf := makeMinimalSplunkforwarder("SplunkForwarder", "splunkforwarder.managed.openshift.io/v1alpha1", "osde2e-splunkforwarder-test-1")
			err := dedicatedAaddSplunkforwarder(sf, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})

	ginkgo.Context("splunkforwarders", func() {
		ginkgo.It("admin should be able to manage splunkforwarders CR", func() {
			sf := makeMinimalSplunkforwarder("SplunkForwarder", "splunkforwarder.managed.openshift.io/v1alpha1", "osde2e-splunkforwarder-test-2")
			err := addSplunkforwarder(sf, h)
			Expect(err).NotTo(HaveOccurred())

			h.Dynamic().Resource(schema.GroupVersionResource{
				Group: "splunkforwarder.managed.openshift.io", Version: "v1alpha1", Resource: "splunkforwarders",
			}).Namespace(operatorNamespace).Delete(context.TODO(), "osde2e-splunkforwarder-test-2", metav1.DeleteOptions{})
		})
	})

})

func makeMinimalSplunkforwarder(kind string, apiversion string, name string) sfv1alpha1.SplunkForwarder {
	sf := sfv1alpha1.SplunkForwarder{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: apiversion,
		},
		ObjectMeta: metav1.ObjectMeta{

			Name: name,
		},
		Spec: sfv1alpha1.SplunkForwarderSpec{
			SplunkLicenseAccepted: true,
			UseHeavyForwarder:     false,
			SplunkInputs: []sfv1alpha1.SplunkForwarderInputs{
				{
					Path: "",
				},
			},
		},
	}
	return sf
}

func dedicatedAaddSplunkforwarder(SplunkForwarder sfv1alpha1.SplunkForwarder, h *helper.H) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(SplunkForwarder.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "splunkforwarder.managed.openshift.io", Version: "v1alpha1", Resource: "splunkforwarders",
	}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
	return (err)
}

func addSplunkforwarder(SplunkForwarder sfv1alpha1.SplunkForwarder, h *helper.H) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(SplunkForwarder.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "splunkforwarder.managed.openshift.io", Version: "v1alpha1", Resource: "splunkforwarders",
	}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
	return (err)
}
