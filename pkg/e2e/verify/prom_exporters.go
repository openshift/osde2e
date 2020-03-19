package verify

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/osd"
	"github.com/openshift/osde2e/pkg/common/state"
)

var _ = ginkgo.Describe("[Suite: e2e] [OSD] Prometheus Exporters", func() {

	var namespace = "openshift-monitoring"

	var services = []string{
		"sre-ebs-iops-reporter",
		"sre-dns-latency-exporter",
		"sre-stuck-ebs-vols",
	}

	var configMaps = []string{
		"sre-dns-latency-exporter-code",
		"sre-stuck-ebs-vols-code",
		"sre-ebs-iops-reporter-code",
	}

	var secrets = []string{
		"sre-ebs-iops-reporter-aws-credentials",
		"sre-stuck-ebs-vols-aws-credentials",
	}

	var roleBindings = []string{
		"sre-dns-latency-exporter",
		"sre-ebs-iops-reporter",
		"sre-stuck-ebs-vols",
	}

	var daemonSets = []string{
		"sre-dns-latency-exporter",
	}

	h := helper.New()

	ginkgo.It("should exist and be running in the cluster", func() {

		// Expect project to exist
		_, err := h.Project().ProjectV1().Projects().Get(namespace, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "project should have been created")

		// Ensure presence of config maps
		for _, configMapName := range configMaps {
			_, err = h.Kube().CoreV1().ConfigMaps(namespace).Get(configMapName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get config map %v\n", configMapName)
		}

		// Ensure presence of secrets
		for _, secretName := range secrets {
			_, err = h.Kube().CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get secret %v\n", secretName)
		}

		// Ensure presence of rolebindings
		for _, roleBindingName := range roleBindings {
			_, err = h.Kube().RbacV1().RoleBindings(namespace).Get(roleBindingName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get role binding %v\n", roleBindingName)
		}

		// Ensure daemonsets are present and satisfied
		currentClusterVersion, err := osd.OpenshiftVersionToSemver(state.Instance.Cluster.Version)
		Expect(err).NotTo(HaveOccurred(), "error parsing cluster version %s", state.Instance.Cluster.Version)

		for _, daemonSetName := range daemonSets {
			// Use appv1 for clusters 4.4.0 or later
			if osd.Version440.Check(currentClusterVersion) {
				daemonSet, err := h.Kube().AppsV1().DaemonSets(namespace).Get(daemonSetName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get daemonset %v\n", daemonSetName)
				Expect(daemonSet.Status.DesiredNumberScheduled).Should(Equal(daemonSet.Status.CurrentNumberScheduled),
					"daemonset desired count should match currently running")
			} else {
				daemonSet, err := h.Kube().ExtensionsV1beta1().DaemonSets(namespace).Get(daemonSetName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get daemonset %v\n", daemonSetName)
				Expect(daemonSet.Status.DesiredNumberScheduled).Should(Equal(daemonSet.Status.CurrentNumberScheduled),
					"daemonset desired count should match currently running")
			}
		}

		// Ensure services are present
		for _, serviceName := range services {
			service, err := h.Kube().CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get service %v\n", serviceName)
			Expect(service.Spec.ClusterIP).Should(Not(BeNil()),
				"failed to get service cluster ip for %v\n", serviceName)
			Expect(service.Spec.Ports).Should(Not(BeEmpty()),
				"failed to get service cluster ports for %v\n", serviceName)
		}
	}, 300)

})
