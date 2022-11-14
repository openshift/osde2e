package operators

import (
	"context"
	"log"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	sfv1alpha1 "github.com/openshift/splunk-forwarder-operator/pkg/apis/splunkforwarder/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

var (
	splunkForwarderBlocking  string = "[Suite: operators] [OSD] Splunk Forwarder Operator"
	splunkForwarderInforming string = "[Suite: informing] [OSD] Splunk Forwarder Operator"
)

func init() {
	alert.RegisterGinkgoAlert(splunkForwarderBlocking, "SD-SREP", "@srep-security-team", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	alert.RegisterGinkgoAlert(splunkForwarderInforming, "SD-SREP", "@srep-security-team", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

// Blocking SplunkForwarder Signal
var _ = ginkgo.Describe(splunkForwarderBlocking, func() {
	operatorName := "splunk-forwarder-operator"
	var operatorNamespace string = "openshift-splunk-forwarder-operator"
	var operatorLockFile string = "splunk-forwarder-operator-lock"
	var defaultDesiredReplicas int32 = 1
	clusterRoleBindings := []string{
		"splunk-forwarder-operator-clusterrolebinding",
	}

	clusterRoles := []string{
		"splunk-forwarder-operator",
		"splunk-forwarder-operator-og-admin",
		"splunk-forwarder-operator-og-edit",
		"splunk-forwarder-operator-og-view",
	}

	splunkforwarder_names := []string{
		"osde2e-dedicated-admin-splunkforwarder-x",
		"osde2e-splunkforwarder-test-2",
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

	// Clean up splunkforwarders after tests
	ginkgo.JustAfterEach(func(ctx context.Context) {
		namespace := "openshift-splunk-forwarder-operator"
		for _, name := range splunkforwarder_names {
			err := deleteSplunkforwarder(ctx, name, namespace, h)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	ginkgo.Context("splunkforwarders", func() {
		util.GinkgoIt("admin should be able to manage SplunkForwarders CR", func(ctx context.Context) {
			name := "osde2e-splunkforwarder-test-2"
			sf := makeMinimalSplunkforwarder("SplunkForwarder", "splunkforwarder.managed.openshift.io/v1alpha1", name)
			err := addSplunkforwarder(ctx, sf, "openshift-splunk-forwarder-operator", h)
			Expect(err).NotTo(HaveOccurred())
		}, float64(defaultTimeout))
	})
})

// Informing SplunkForwarder Signal
var _ = ginkgo.Describe(splunkForwarderInforming, func() {
	operatorName := "splunk-forwarder-operator"
	var operatorNamespace string = "openshift-splunk-forwarder-operator"
	var operatorLockFile string = "splunk-forwarder-operator-lock"
	var defaultDesiredReplicas int32 = 1
	clusterRoleBindings := []string{
		"splunk-forwarder-operator-clusterrolebinding",
	}

	clusterRoles := []string{
		"splunk-forwarder-operator",
		"splunk-forwarder-operator-og-admin",
		"splunk-forwarder-operator-og-edit",
		"splunk-forwarder-operator-og-view",
	}

	splunkforwarder_names := []string{
		"osde2e-dedicated-admin-splunkforwarder-x",
		"osde2e-splunkforwarder-test-2",
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

	// Clean up splunkforwarders after tests
	ginkgo.JustAfterEach(func(ctx context.Context) {
		namespace := "openshift-splunk-forwarder-operator"
		for _, name := range splunkforwarder_names {
			err := deleteSplunkforwarder(ctx, name, namespace, h)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	ginkgo.Context("splunkforwarders", func() {
		util.GinkgoIt("dedicated admin should not be able to manage SplunkForwarders CR", func(ctx context.Context) {
			name := "osde2e-dedicated-admin-splunkforwarder-x"
			sf := makeMinimalSplunkforwarder("SplunkForwarder", "splunkforwarder.managed.openshift.io/v1alpha1", name)
			err := dedicatedAaddSplunkforwarder(ctx, sf, "openshift-splunk-forwarder-operator", h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		}, float64(defaultTimeout))
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

func dedicatedAaddSplunkforwarder(ctx context.Context, SplunkForwarder sfv1alpha1.SplunkForwarder, namespace string, h *helper.H) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(SplunkForwarder.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	h.Impersonate(rest.ImpersonationConfig{
		UserName: "test-user@redhat.com",
		Groups: []string{
			"dedicated-admins",
		},
	})
	defer func() {
		h.Impersonate(rest.ImpersonationConfig{})
	}()
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "splunkforwarder.managed.openshift.io", Version: "v1alpha1", Resource: "splunkforwarders",
	}).Namespace(namespace).Create(ctx, &unstructuredObj, metav1.CreateOptions{})
	return (err)
}

func addSplunkforwarder(ctx context.Context, SplunkForwarder sfv1alpha1.SplunkForwarder, namespace string, h *helper.H) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(SplunkForwarder.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "splunkforwarder.managed.openshift.io", Version: "v1alpha1", Resource: "splunkforwarders",
	}).Namespace(namespace).Create(ctx, &unstructuredObj, metav1.CreateOptions{})
	return (err)
}

func deleteSplunkforwarder(ctx context.Context, name string, namespace string, h *helper.H) error {
	_, err := h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "splunkforwarder.managed.openshift.io", Version: "v1alpha1", Resource: "splunkforwarders",
	}).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	log.Printf("Get splunkforwarder %s in namespace %s; Error:(%v)", name, operatorNamespace, err)
	if err == nil {
		e := h.Dynamic().Resource(schema.GroupVersionResource{
			Group: "splunkforwarder.managed.openshift.io", Version: "v1alpha1", Resource: "splunkforwarders",
		}).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		log.Printf("Delete splunkforwarder %s in namespace %s; Error:(%v)", name, operatorNamespace, e)
		return (e)
	}
	return nil
}
