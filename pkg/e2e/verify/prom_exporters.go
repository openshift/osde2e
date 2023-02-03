package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/cluster"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	clusterProviders "github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var promExportersTestname string = "[Suite: e2e] [OSD] Prometheus Exporters"

func init() {
	alert.RegisterGinkgoAlert(promExportersTestname, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(promExportersTestname, label.E2E, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("Prometheus Exporters (ebs-iops-reporter and stuck-ebs-vols) are not deployed to STS clusters")
		}
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("Prometheus Exporters (ebs-iops-reporter and stuck-ebs-vols) are not deployed to HyperShift clusters")
		}
	})

	const (
		// all represents all environments
		allProviders = "all"

		// aws represents tests to be run on AWS environments
		awsProvider = "aws"
	)

	promNamespace := "openshift-monitoring"

	servicesToCheck := map[string][]string{
		allProviders: {
			"sre-dns-latency-exporter",
		},
		awsProvider: {
			"sre-ebs-iops-reporter",
			"sre-stuck-ebs-vols",
		},
	}

	configMapsToCheck := map[string][]string{
		allProviders: {
			"sre-dns-latency-exporter-code",
		},
		awsProvider: {
			"sre-stuck-ebs-vols-code",
			"sre-ebs-iops-reporter-code",
		},
	}

	secretsToCheck := map[string][]string{
		awsProvider: {
			"sre-ebs-iops-reporter-aws-credentials",
			"sre-stuck-ebs-vols-aws-credentials",
		},
	}

	roleBindingsToCheck := map[string][]string{
		allProviders: {
			"sre-dns-latency-exporter",
		},
		awsProvider: {
			"sre-ebs-iops-reporter",
			"sre-stuck-ebs-vols",
		},
	}

	daemonSetsToCheck := map[string][]string{
		allProviders: {
			"sre-dns-latency-exporter",
		},
	}

	h := helper.New()

	ginkgo.It("should exist and be running in the cluster", func(ctx context.Context) {
		envs := []string{allProviders, viper.GetString(config.CloudProvider.CloudProviderID)}

		// Expect project to exist
		_, err := h.Project().ProjectV1().Projects().Get(ctx, promNamespace, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "project should have been created")

		// Ensure presence of config maps
		checkConfigMaps(ctx, promNamespace, configMapsToCheck, h, envs...)

		// Ensure presence of secrets
		checkSecrets(ctx, promNamespace, secretsToCheck, h, envs...)

		// Ensure presence of rolebindings
		checkRoleBindings(ctx, promNamespace, roleBindingsToCheck, h, envs...)

		// Ensure daemonsets are present and satisfied
		checkDaemonSets(ctx, promNamespace, daemonSetsToCheck, h, envs...)

		// Ensure services are present
		checkServices(ctx, promNamespace, servicesToCheck, h, envs...)
	})
})

func checkConfigMaps(ctx context.Context, namespace string, configMaps map[string][]string, h *helper.H, providers ...string) {
	for _, provider := range providers {
		for _, configMapName := range configMaps[provider] {
			_, err := h.Kube().CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get config map %v\n", configMapName)
		}
	}
}

func checkSecrets(ctx context.Context, namespace string, secrets map[string][]string, h *helper.H, providers ...string) {
	for _, provider := range providers {
		for _, secretName := range secrets[provider] {
			_, err := h.Kube().CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get secret %v\n", secretName)
		}
	}
}

func checkRoleBindings(ctx context.Context, namespace string, roleBindings map[string][]string, h *helper.H, providers ...string) {
	for _, provider := range providers {
		for _, roleBindingName := range roleBindings[provider] {
			_, err := h.Kube().RbacV1().RoleBindings(namespace).Get(ctx, roleBindingName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get role binding %v\n", roleBindingName)
		}
	}
}

func checkDaemonSets(ctx context.Context, namespace string, daemonSets map[string][]string, h *helper.H, providers ...string) {
	provider, err := clusterProviders.ClusterProvider()
	Expect(err).NotTo(HaveOccurred(), "error getting cluster provider")
	currentClusterVersion, err := cluster.GetClusterVersion(provider, viper.GetString(config.Cluster.ID))
	Expect(err).NotTo(HaveOccurred(), "error getting cluster version %s", viper.GetString(config.Cluster.Version))

	for _, provider := range providers {
		for _, daemonSetName := range daemonSets[provider] {
			// Use appv1 for clusters 4.4.0 or later
			if util.Version440.Check(currentClusterVersion) {
				daemonSet, err := h.Kube().AppsV1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get daemonset %v\n", daemonSetName)
				Expect(daemonSet.Status.DesiredNumberScheduled).Should(Equal(daemonSet.Status.CurrentNumberScheduled),
					"daemonset desired count should match currently running")
			} else {
				daemonSet, err := h.Kube().ExtensionsV1beta1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get daemonset %v\n", daemonSetName)
				Expect(daemonSet.Status.DesiredNumberScheduled).Should(Equal(daemonSet.Status.CurrentNumberScheduled),
					"daemonset desired count should match currently running")
			}
		}
	}
}

func checkServices(ctx context.Context, namespace string, services map[string][]string, h *helper.H, providers ...string) {
	for _, provider := range providers {
		for _, serviceName := range services[provider] {
			service, err := h.Kube().CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get service %v\n", serviceName)
			Expect(service.Spec.ClusterIP).Should(Not(BeNil()),
				"failed to get service cluster ip for %v\n", serviceName)
			Expect(service.Spec.Ports).Should(Not(BeEmpty()),
				"failed to get service cluster ports for %v\n", serviceName)
		}
	}
}
