package helper

import (
	"github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"

	imagev1 "github.com/openshift/api/image/v1"
	quotav1 "github.com/openshift/api/quota/v1"
	route "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	customdomainv1alpha1 "github.com/openshift/custom-domains-operator/api/v1alpha1"
	mustgatherv1alpha1 "github.com/openshift/must-gather-operator/api/v1alpha1"
	operatorhubv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func (h *H) AsUser(username string, groups ...string) *resources.Resources {
	if username != "" {
		// these groups are required for impersonating a user
		groups = append(groups, "system:authenticated", "system:authenticated:oauth")
	}

	h.Impersonate(rest.ImpersonationConfig{
		UserName: username,
		Groups:   groups,
	})

	client, err := resources.New(h.GetConfig())
	gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred(), "failed to create resources client object")

	// register schemas here
	configv1.AddToScheme(client.GetScheme())
	quotav1.AddToScheme(client.GetScheme())
	securityv1.AddToScheme(client.GetScheme())
	monitoringv1.AddToScheme(client.GetScheme())
	mustgatherv1alpha1.AddToScheme(client.GetScheme())
	customdomainv1alpha1.AddToScheme(client.GetScheme())
	route.AddToScheme(client.GetScheme())
	operatorhubv1.AddToScheme(client.GetScheme())
	imagev1.AddToScheme(client.GetScheme())

	return client
}

func (h *H) AsServiceAccount(name string) *resources.Resources {
	h.ServiceAccount = name
	return h.AsUser(name)
}

func (h *H) AsDedicatedAdmin() *resources.Resources {
	return h.AsUser("test-user@redhat.com", "dedicated-admins")
}

func (h *H) AsClusterAdmin() *resources.Resources {
	return h.AsUser("system:admin", "cluster-admins")
}
