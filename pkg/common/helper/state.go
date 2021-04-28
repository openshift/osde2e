package helper

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// cluster-wide resources that are retrieved from cluster
	desiredClusterResources = []schema.GroupVersionResource{
		// apis
		{Group: "apiregistration.k8s.io", Version: "v1", Resource: "apiservices"},

		// cloud credentials
		{Group: "cloudcredential.openshift.io", Version: "v1", Resource: "credentialsrequests"},

		// openshift config
		{Group: "config.openshift.io", Version: "v1", Resource: "apiservers"},
		{Group: "config.openshift.io", Version: "v1", Resource: "authentications"},
		{Group: "config.openshift.io", Version: "v1", Resource: "builds"},
		{Group: "config.openshift.io", Version: "v1", Resource: "clusteroperators"},
		{Group: "config.openshift.io", Version: "v1", Resource: "clusterversions"},
		{Group: "config.openshift.io", Version: "v1", Resource: "consoles"},
		{Group: "config.openshift.io", Version: "v1", Resource: "dnses"},
		{Group: "config.openshift.io", Version: "v1", Resource: "featuregates"},
		{Group: "config.openshift.io", Version: "v1", Resource: "images"},
		{Group: "config.openshift.io", Version: "v1", Resource: "infrastructures"},
		{Group: "config.openshift.io", Version: "v1", Resource: "ingresses"},
		{Group: "config.openshift.io", Version: "v1", Resource: "networks"},
		{Group: "config.openshift.io", Version: "v1", Resource: "oauths"},
		{Group: "config.openshift.io", Version: "v1", Resource: "projects"},
		{Group: "config.openshift.io", Version: "v1", Resource: "schedulers"},

		// machine config
		{Group: "machineconfiguration.openshift.io", Version: "v1", Resource: "machineconfigpools"},
		{Group: "machineconfiguration.openshift.io", Version: "v1", Resource: "machineconfigs"},

		// operators
		{Group: "operator.openshift.io", Version: "v1", Resource: "kubeapiservers"},
		{Group: "operator.openshift.io", Version: "v1", Resource: "kubecontrollermanagers"},
		{Group: "operator.openshift.io", Version: "v1", Resource: "openshiftapiservers"},

		// core
		{Group: "", Version: "v1", Resource: "namespaces"},
		{Group: "", Version: "v1", Resource: "nodes"},
	}

	// namespaced resources that are retrieved from cluster
	desiredResources = []schema.GroupVersionResource{
		// apps
		{Group: "apps", Version: "v1", Resource: "statefulsets"},

		// rbac
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},

		// machine
		{Group: "machine.openshift.io", Version: "v1beta1", Resource: "machines"},

		// core
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "endpoints"},
		{Group: "", Version: "v1", Resource: "events"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "persistentvolumes"},
		{Group: "", Version: "v1", Resource: "pods"},
	}
)

// GetClusterState retrieves the current objects in desiredClusterResources and desiredResources.
func (h *H) GetClusterState() (resources map[schema.GroupVersionResource]*unstructured.UnstructuredList) {
	// setup client
	client := h.Dynamic()

	numItems := len(desiredClusterResources) + len(desiredResources)
	resources = make(map[schema.GroupVersionResource]*unstructured.UnstructuredList, numItems)
	listOpts := metav1.ListOptions{}
	// retrieve cluster-wide resources
	for _, r := range desiredClusterResources {
		if list, err := client.Resource(r).List(context.TODO(), listOpts); err != nil {
			log.Printf("Encountered error listing getting resource '%s': %v", r, err)
		} else {
			resources[r] = list
		}
	}

	// retrieve namespaces resources
	for _, r := range desiredResources {
		if list, err := client.Resource(r).Namespace(metav1.NamespaceAll).List(context.TODO(), listOpts); err != nil {
			log.Printf("Encountered error listing getting resource '%s': %v", r, err)
		} else {
			resources[r] = list
		}
	}
	return
}
