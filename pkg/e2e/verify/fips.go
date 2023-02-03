package verify

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
)

const (
	fipsTestName = "[Suite: e2e] FIPS"
)

func init() {
	alert.RegisterGinkgoAlert(fipsTestName, "SD-SREP", "Trevor Nierman", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(fipsTestName, ginkgo.Ordered, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources

	ginkgo.BeforeAll(func() {
		if !viper.GetBool(config.Cluster.EnableFips) {
			ginkgo.Skip("FIPs is not enabled")
		}
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("FIPs is not currently supported on HyperShift")
		}
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.It("for all nodes in a cluster", func(ctx context.Context) {
		// Create Daemonset to test FIPS for each node
		// The test consists of mounting '/proc/sys/crypto/fips_enabled' to an
		// init container and checking that the value is 1. Ready pods indicate
		// FIPS is enabled for that node, nodes that don't have FIPS enabled
		// will never see their pods become ready
		ds := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-fips-",
				Namespace:    h.CurrentProject(),
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"osde2e": "fips-test"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"osde2e": "fips-test"},
					},
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{{
							Name:    "test-fips",
							Image:   "registry.access.redhat.com/ubi8/ubi-minimal",
							Command: []string{"/bin/sh"},
							// Test that node has FIPS enabled
							Args: []string{"-c", "if [[ $(cat /fips_enabled) -eq 1 ]]; then exit 0; else exit 1; fi"},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "fips-enabled",
								ReadOnly:  true,
								MountPath: "/fips_enabled",
							}},
							SecurityContext: &corev1.SecurityContext{
								Privileged: pointer.Bool(true),
							},
						}},
						Containers: []corev1.Container{{
							Name:  "pause",
							Image: "registry.k8s.io/pause:latest",
						}},
						Volumes: []corev1.Volume{{
							Name: "fips-enabled",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/proc/sys/crypto/fips_enabled",
								},
							},
						}},
						ServiceAccountName: "cluster-admin",
						Tolerations: []corev1.Toleration{
							{
								Key:      "node-role.kubernetes.io/master",
								Operator: corev1.TolerationOpEqual,
								Effect:   corev1.TaintEffectNoSchedule,
							},
							{
								Key:      "node-role.kubernetes.io/infra",
								Operator: corev1.TolerationOpEqual,
								Effect:   corev1.TaintEffectNoSchedule,
							},
						},
					},
				},
			},
		}

		expect.NoError(client.Create(ctx, ds))

		err := wait.For(func() (bool, error) {
			daemonset := &appsv1.DaemonSet{}
			err := client.Get(ctx, ds.GetName(), ds.GetNamespace(), daemonset)
			if err != nil {
				return false, err
			}
			desired, scheduled, ready := daemonset.Status.DesiredNumberScheduled, daemonset.Status.CurrentNumberScheduled, daemonset.Status.NumberReady
			return desired == scheduled && desired == ready, nil
		})
		expect.NoError(err, "daemonset never became available")

		expect.NoError(client.Delete(ctx, ds))
	})
})
