package verify

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/util/slice"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

var dedicatedAdminSccTestName = "[Suite: e2e] [OSD] RBAC Dedicated Admins SCC permissions"

func init() {
	alert.RegisterGinkgoAlert(dedicatedAdminSccTestName, "SD-CICD", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(dedicatedAdminSccTestName, ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.DescribeTable("should include", func(ctx context.Context, scc string) {
		clusterRole := &rbacv1.ClusterRole{}
		expect.NoError(client.Get(ctx, "dedicated-admins-cluster", "", clusterRole))
		for _, rule := range clusterRole.Rules {
			if slice.ContainsString(rule.Resources, "securitycontextconstraints", nil) && slice.ContainsString(rule.Verbs, "use", nil) {
				Expect(slice.ContainsString(rule.ResourceNames, scc, nil)).To(BeTrue(), "ClusterRole resource did not contain %s", scc)
			}
		}
	},
		ginkgo.Entry("anyuid", "anyuid"),
		ginkgo.Entry("nonroot", "nonroot"),
	)

	ginkgo.It("allow a pod to be created with a SecurityContext", func(ctx context.Context) {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "osde2e-anyuid-",
				Namespace:    h.CurrentProject(),
			},
			Spec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					RunAsUser:    pointer.Int64(987654321),
					RunAsNonRoot: pointer.Bool(false),
				},
				Containers: []v1.Container{
					{
						Name:  "test",
						Image: "openshift/hello-openshift",
						SecurityContext: &v1.SecurityContext{
							AllowPrivilegeEscalation: pointer.Bool(false),
							Capabilities:             &v1.Capabilities{Drop: []v1.Capability{"ALL"}},
							SeccompProfile:           &v1.SeccompProfile{Type: v1.SeccompProfileTypeRuntimeDefault},
						},
					},
				},
			},
		}
		expect.NoError(client.Create(ctx, pod))
		expect.NoError(client.Delete(ctx, pod))
	})

	ginkgo.It("new SCCs do not break existing workloads", func(ctx context.Context) {
		deploymentName := "prometheus-operator"
		namespace := "openshift-monitoring"

		err := wait.For(func() (bool, error) {
			deployment := &appsv1.Deployment{}
			err := client.Get(ctx, deploymentName, namespace, deployment)
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			return true, nil
		})
		expect.NoError(err)

		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace}}
		err = wait.For(conditions.New(client).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, v1.ConditionTrue), wait.WithTimeout(60*time.Second))
		expect.NoError(err)

		scc := &securityv1.SecurityContextConstraints{
			ObjectMeta:         metav1.ObjectMeta{GenerateName: "osde2e-scc-"},
			Groups:             []string{"system:authenticated"},
			SELinuxContext:     securityv1.SELinuxContextStrategyOptions{Type: securityv1.SELinuxStrategyRunAsAny},
			RunAsUser:          securityv1.RunAsUserStrategyOptions{Type: securityv1.RunAsUserStrategyRunAsAny},
			FSGroup:            securityv1.FSGroupStrategyOptions{Type: securityv1.FSGroupStrategyRunAsAny},
			SupplementalGroups: securityv1.SupplementalGroupsStrategyOptions{Type: securityv1.SupplementalGroupsStrategyRunAsAny},
		}
		expect.NoError(client.Create(ctx, scc))

		deployment = &appsv1.Deployment{}
		expect.NoError(client.Get(ctx, deploymentName, namespace, deployment))

		originalReplicaCount := deployment.Spec.DeepCopy().Replicas
		deployment.Spec.Replicas = pointer.Int32(0)
		expect.NoError(client.Update(ctx, deployment))

		err = wait.For(conditions.New(client).ResourceScaled(deployment, func(object k8s.Object) int32 {
			return object.(*appsv1.Deployment).Status.ReadyReplicas
		}, 0))
		expect.NoError(err, "deployment never scaled to 0")

		deployment.Spec.Replicas = originalReplicaCount
		expect.NoError(client.Update(ctx, deployment))

		err = wait.For(conditions.New(client).ResourceScaled(deployment, func(object k8s.Object) int32 {
			return object.(*appsv1.Deployment).Status.ReadyReplicas
		}, *originalReplicaCount))
		expect.NoError(err, "deployment never scaled back up")

		expect.NoError(client.Delete(ctx, scc))
	})
})
