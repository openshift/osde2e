package verify

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	fipsTestName = "[Suite: e2e] FIPS"
	fipsTestPollInterval = 10*time.Second
	fipsTestPollDuration = 5*time.Minute
)

func init() {
	alert.RegisterGinkgoAlert(fipsTestName, "SD-SREP", "Trevor Nierman", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(fipsTestName, func() {
	ginkgo.Context("is enabled", func() {
		if !viper.GetBool(config.Tests.EnableFips) {
			return
		}
		h := helper.New()

		ginkgo.It("for all nodes in a cluster", func() {
			testName := fmt.Sprintf("test-fips-%s-%d-%d", time.Now().Format("20060102-150405"), time.Now().Nanosecond()/1000000, ginkgo.GinkgoParallelNode())

			// Create Daemonset to test FIPS for each node
			// The test consists of mounting '/proc/sys/crypto/fips_enabled' to an init container and checking that the value is 1
			// Ready pods indicate FIPS is enabled for that node, nodes that don't have FIPS enabled will never see their pods become ready
			priv := true
			ds := &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Daemonset",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: testName,
					Namespace: h.CurrentProject(),
				},
				Spec: appsv1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string {"osde2e": "fips-test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-fips",
							Namespace: h.CurrentProject(),
							Labels: map[string]string {"osde2e": "fips-test"},
						},
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{{
								Name: "test-fips",
								Image: "registry.access.redhat.com/ubi8/ubi-minimal",
								Command: []string{ "/bin/sh" },
								// Test that node has FIPS enabled 
								Args: []string{ "-c", "if [[ $(cat /fips_enabled) -eq 1 ]]; then exit 0; else exit 1; fi"},
								VolumeMounts: []corev1.VolumeMount{{
									Name: "fips-enabled",
									ReadOnly: true,
									MountPath: "/fips_enabled",
								}},
								SecurityContext: &corev1.SecurityContext{
									Privileged: &priv,
								},
							}},
							Containers: []corev1.Container{{
								Name: "sleep",
								Image: "registry.access.redhat.com/ubi8/ubi-minimal",
								Command: []string{ "/bin/sh" },
								Args: []string{ "-c", "sleep infinity"},
							}},
							Volumes: []corev1.Volume{{
								Name: "fips-enabled",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource {
										Path: "/proc/sys/crypto/fips_enabled",
									},
								},
							}},
							ServiceAccountName: "cluster-admin",
							Tolerations: []corev1.Toleration{
								{
									Key: "node-role.kubernetes.io/master",
									Operator: corev1.TolerationOpEqual,
									Effect: corev1.TaintEffectNoSchedule,
								},
								{
									Key: "node-role.kubernetes.io/infra",
									Operator: corev1.TolerationOpEqual,
									Effect: corev1.TaintEffectNoSchedule,
								},
							},
						},
					},
				},
			}

			ds, err := h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Create(context.TODO(), ds, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				err := h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Delete(context.TODO(), ds.Name, metav1.DeleteOptions{})
				Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error deleting Daemonset: %v", err))
			}()

			nodes, err := h.Kube().CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error retrieving nodes: %v", err))

			// Poll until the number of ready pods is equal to the number of nodes, indicating FIPS is enabled for the cluster
			err = wait.PollImmediate(fipsTestPollInterval, fipsTestPollDuration, func() (bool, error) {
				ds, err = h.Kube().AppsV1().DaemonSets(h.CurrentProject()).Get(context.TODO(), testName, metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				if ds.Status.NumberReady == int32(len(nodes.Items)){
					return true, err
				}
				return false, err
			})
			Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error waiting for pods to become ready: %v", err))
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
