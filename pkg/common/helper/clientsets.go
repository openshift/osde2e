package helper

import (
	. "github.com/onsi/gomega"

	config "github.com/openshift/client-go/config/clientset/versioned"
	image "github.com/openshift/client-go/image/clientset/versioned"
	project "github.com/openshift/client-go/project/clientset/versioned"
	route "github.com/openshift/client-go/route/clientset/versioned"
	user "github.com/openshift/client-go/user/clientset/versioned"
	machine "github.com/openshift/machine-api-operator/pkg/generated/clientset/versioned"
	operator "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	velero "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Impersonate sets impersonate user headers
func (h *H) Impersonate(user rest.ImpersonationConfig) *H {
	h.restConfig.Impersonate = user
	return h
}

// Cfg return a client for the Config API.
func (h *H) Cfg() config.Interface {
	client, err := config.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Config client")
	return client
}

// Dynamic returns a client that works on arbitrary types.
func (h *H) Dynamic() dynamic.Interface {
	client, err := dynamic.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Dynamic client")
	return client
}

// Kube returns the clientset for Kubernetes upstream.
func (h *H) Kube() kubernetes.Interface {
	client, err := kubernetes.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Kubernetes clientset")
	return client
}

// Velero returns the clientset for Velero objects.
func (h *H) Velero() velero.Interface {
	client, err := velero.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Velero clientset")
	return client
}

// Image returns the clientset for images.
func (h *H) Image() image.Interface {
	client, err := image.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Image clientset")
	return client
}

// Route returns the clientset for routing.
func (h *H) Route() route.Interface {
	client, err := route.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Route clientset")
	return client
}

// Project returns the clientset for projects.
func (h *H) Project() project.Interface {
	client, err := project.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Project clientset")
	return client
}

// User returns the clientset for users
func (h *H) User() user.Interface {
	client, err := user.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Project clientset")
	return client
}

// Operator returns the clientset for operator-lifecycle-manager
func (h *H) Operator() operator.Interface {
	client, err := operator.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Operator clientset")
	return client
}

// Machine returns the clientset for openshift-machine-api
func (h *H) Machine() machine.Interface {
	client, err := machine.NewForConfig(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Operator clientset")
	return client
}

// REST returns a client for generic operations.
func (h *H) REST() *rest.RESTClient {
	client, err := rest.RESTClientFor(h.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure REST client")
	return client
}
