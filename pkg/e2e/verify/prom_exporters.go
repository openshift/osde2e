package verify

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	clusterProviders "github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/util"
)

var promExportersTestname string = "[Suite: e2e] [OSD] Prometheus Exporters"

func init() {
	alert.RegisterGinkgoAlert(promExportersTestname, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(promExportersTestname, func() {
	const (
		// all represents all environments
		allProviders = "all"

		// aws represents tests to be run on AWS environments
		awsProvider = "aws"
	)

	var promNamespace = "openshift-monitoring"

	var servicesToCheck = map[string][]string{
		allProviders: []string{
			"sre-dns-latency-exporter",
		},
		awsProvider: []string{
			"sre-ebs-iops-reporter",
			"sre-stuck-ebs-vols",
		},
	}

	var configMapsToCheck = map[string][]string{
		allProviders: []string{
			"sre-dns-latency-exporter-code",
		},
		awsProvider: []string{
			"sre-stuck-ebs-vols-code",
			"sre-ebs-iops-reporter-code",
		},
	}

	var secretsToCheck = map[string][]string{
		awsProvider: []string{
			"sre-ebs-iops-reporter-aws-credentials",
			"sre-stuck-ebs-vols-aws-credentials",
		},
	}

	var roleBindingsToCheck = map[string][]string{
		allProviders: []string{
			"sre-dns-latency-exporter",
		},
		awsProvider: []string{
			"sre-ebs-iops-reporter",
			"sre-stuck-ebs-vols",
		},
	}

	var daemonSetsToCheck = map[string][]string{
		allProviders: []string{
			"sre-dns-latency-exporter",
		},
	}

	h := helper.New()

	ginkgo.It("should exist and be running in the cluster", func() {

		envs := []string{allProviders, viper.GetString(config.CloudProvider.CloudProviderID)}

		// Expect project to exist
		_, err := h.Project().ProjectV1().Projects().Get(context.TODO(), promNamespace, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "project should have been created")

		// Ensure presence of config maps
		checkConfigMaps(promNamespace, configMapsToCheck, h, envs...)

		// Ensure presence of secrets
		checkSecrets(promNamespace, secretsToCheck, h, envs...)

		// Ensure presence of rolebindings
		checkRoleBindings(promNamespace, roleBindingsToCheck, h, envs...)

		// Ensure daemonsets are present and satisfied
		checkDaemonSets(promNamespace, daemonSetsToCheck, h, envs...)

		// Ensure services are present
		checkServices(promNamespace, servicesToCheck, h, envs...)
	}, 300)

})

func checkConfigMaps(namespace string, configMaps map[string][]string, h *helper.H, providers ...string) {
	for _, provider := range providers {
		for _, configMapName := range configMaps[provider] {
			_, err := h.Kube().CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get config map %v\n", configMapName)
		}
	}
}

func checkSecrets(namespace string, secrets map[string][]string, h *helper.H, providers ...string) {
	for _, provider := range providers {
		for _, secretName := range secrets[provider] {
			_, err := h.Kube().CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get secret %v\n", secretName)
		}
	}
}

func checkRoleBindings(namespace string, roleBindings map[string][]string, h *helper.H, providers ...string) {
	for _, provider := range providers {
		for _, roleBindingName := range roleBindings[provider] {
			_, err := h.Kube().RbacV1().RoleBindings(namespace).Get(context.TODO(), roleBindingName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get role binding %v\n", roleBindingName)
		}
	}
}

func checkDaemonSets(namespace string, daemonSets map[string][]string, h *helper.H, providers ...string) {
	provider, err := clusterProviders.ClusterProvider()
	Expect(err).NotTo(HaveOccurred(), "error getting cluster provider")
	currentClusterVersion, err := cluster.GetClusterVersion(provider, viper.GetString(config.Cluster.ID))
	Expect(err).NotTo(HaveOccurred(), "error getting cluster version %s", viper.GetString(config.Cluster.Version))

	for _, provider := range providers {
		for _, daemonSetName := range daemonSets[provider] {
			// Use appv1 for clusters 4.4.0 or later
			if util.Version440.Check(currentClusterVersion) {
				daemonSet, err := h.Kube().AppsV1().DaemonSets(namespace).Get(context.TODO(), daemonSetName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get daemonset %v\n", daemonSetName)
				Expect(daemonSet.Status.DesiredNumberScheduled).Should(Equal(daemonSet.Status.CurrentNumberScheduled),
					"daemonset desired count should match currently running")
			} else {
				daemonSet, err := h.Kube().ExtensionsV1beta1().DaemonSets(namespace).Get(context.TODO(), daemonSetName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get daemonset %v\n", daemonSetName)
				Expect(daemonSet.Status.DesiredNumberScheduled).Should(Equal(daemonSet.Status.CurrentNumberScheduled),
					"daemonset desired count should match currently running")
			}
		}
	}
}

func checkServices(namespace string, services map[string][]string, h *helper.H, providers ...string) {
	for _, provider := range providers {
		for _, serviceName := range services[provider] {
			service, err := h.Kube().CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get service %v\n", serviceName)
			Expect(service.Spec.ClusterIP).Should(Not(BeNil()),
				"failed to get service cluster ip for %v\n", serviceName)
			Expect(service.Spec.Ports).Should(Not(BeEmpty()),
				"failed to get service cluster ports for %v\n", serviceName)
		}
	}
}
