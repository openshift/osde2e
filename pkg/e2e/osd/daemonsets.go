package osd

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
)

var daemonSetsTestName string = "[Suite: service-definition] [OSD] DaemonSets"

func init() {
	alert.RegisterGinkgoAlert(daemonSetsTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(daemonSetsTestName, ginkgo.Ordered, label.HyperShift, label.ServiceDefinition, func() {
	var h *helper.H
	var client *resources.Resources

	ginkgo.BeforeAll(func(ctx context.Context) {
		h = helper.New()
		client = h.AsUser("")
	})

	newDaemonSet := func(role string) *appsv1.DaemonSet {
		labels := map[string]string{"app": "test"}
		nodeSelectors := map[string]string{}
		if role != "" {
			nodeSelectors[fmt.Sprintf("node-role.kubernetes.io/%s", role)] = ""
		}
		return &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "osde2e-",
				Namespace:    h.CurrentProject(),
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{MatchLabels: labels},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: labels},
					Spec: v1.PodSpec{
						NodeSelector: nodeSelectors,
						Containers: []v1.Container{
							{Name: "pause", Image: "registry.k8s.io/pause:latest"},
						},
					},
				},
			},
		}
	}

	ginkgo.DescribeTable("should get created", func(ctx context.Context, nodeRole string) {
		if nodeRole == "infra" && viper.GetBool(config.Hypershift) {
			ginkgo.Skip("HyperShift does not have infra nodes")
		}

		daemonset := newDaemonSet(nodeRole)
		err := client.Create(ctx, daemonset)
		expect.NoError(err)

		err = wait.For(func() (bool, error) {
			ds := &appsv1.DaemonSet{}
			err = client.Get(ctx, daemonset.GetName(), daemonset.GetNamespace(), ds)
			if err != nil {
				return false, err
			}
			desired, scheduled, ready := ds.Status.DesiredNumberScheduled, ds.Status.CurrentNumberScheduled, ds.Status.NumberReady
			return desired == scheduled && desired == ready, nil
		})
		expect.NoError(err)

		err = client.Delete(ctx, daemonset)
		expect.NoError(err)
	},
		ginkgo.Entry("with no node role", ""),
		ginkgo.Entry("with worker role", "worker"),
		ginkgo.Entry("with infra role", "infra"),
	)
})
