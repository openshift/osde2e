package operators

import (
	"context"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/e2e/operators"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	operatorhubv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

var suiteName = "OCM Agent Operator"

func init() {
	alert.RegisterGinkgoAlert(suiteName, "SD_SREP", "@ocm-agent-operator", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

// TODO: Separate PR to remove 'Informing' label from test suite
var _ = ginkgo.Describe(suiteName, ginkgo.Ordered, label.Operators, label.Informing, func() {
	var h *helper.H
	var client *resources.Resources

	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	var (
		configMapName      = "ocm-agent-config"
		clusterRolePrefix  = "ocm-agent-operator"
		deploymentName     = "ocm-agent"
		namespace          = "openshift-ocm-agent-operator"
		networkPolicyName  = "ocm-agent-allow-only-alertmanager"
		operatorName       = "ocm-agent-operator"
		operatorRegistry   = "ocm-agent-operator-registry"
		secretName         = "ocm-access-token"
		serviceMonitorName = "ocm-agent-metrics"
		serviceName        = "ocm-agent"

		deployments = []string{
			deploymentName,
			deploymentName + "-operator",
		}
	)

	ginkgo.It("cluster service version exists", label.Install, func(ctx context.Context) {
		var csvs operatorhubv1.ClusterServiceVersionList
		err := client.List(ctx, &csvs, resources.WithFieldSelector(labels.FormatLabels(map[string]string{"metadata.namespace": namespace})))
		expect.NoError(err)

		for _, csv := range csvs.Items {
			if csv.Spec.DisplayName == operatorName {
				csv := &operatorhubv1.ClusterServiceVersion{
					ObjectMeta: metav1.ObjectMeta{
						Name:      csv.Name,
						Namespace: namespace,
					},
				}
				err = wait.For(conditions.New(client).ResourceMatch(
					csv, func(object k8s.Object) bool {
						obj := object.(*operatorhubv1.ClusterServiceVersion)
						return obj.Status.Phase == "Succeeded"
					}))
				expect.NoError(err)
				break
			}
		}
	})

	ginkgo.It("deployments exist", label.Install, func(ctx context.Context) {
		for _, deploymentName := range deployments {
			deployment := &appsv1.Deployment{}
			err := client.Get(ctx, deploymentName, namespace, deployment)
			expect.NoError(err)
			Expect(deployment.Status.Replicas).To(BeNumerically("==", 1))
			Expect(deployment.Status.ReadyReplicas).To(BeNumerically("==", 1))
			Expect(deployment.Status.AvailableReplicas).To(BeNumerically("==", 1))
		}
	})

	ginkgo.It("cluster roles exist", label.Install, func(ctx context.Context) {
		var clusterRoles rbacv1.ClusterRoleList
		err := client.List(ctx, &clusterRoles)
		expect.NoError(err)
		found := false
		for _, clusterRole := range clusterRoles.Items {
			if strings.HasPrefix(clusterRole.Name, clusterRolePrefix) {
				found = true
			}
		}
		Expect(found).To(BeTrue())
	})

	ginkgo.It("cluster role bindings exist", label.Install, func(ctx context.Context) {
		var clusterRoleBindings rbacv1.ClusterRoleBindingList
		err := client.List(ctx, &clusterRoleBindings)
		expect.NoError(err)
		found := false
		for _, clusterRoleBinding := range clusterRoleBindings.Items {
			if strings.HasPrefix(clusterRoleBinding.Name, clusterRolePrefix) {
				found = true
			}
		}
		Expect(found).To(BeTrue())
	})

	ginkgo.It("deployment is restored when removed", func(ctx context.Context) {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: namespace,
			},
		}

		err := client.Delete(ctx, deployment)
		expect.NoError(err)

		err = wait.For(conditions.New(client).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, v1.ConditionTrue))
		expect.NoError(err)
	})

	ginkgo.It("config map is restored when removed", func(ctx context.Context) {
		configMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: namespace,
			},
		}

		err := client.Delete(ctx, configMap)
		expect.NoError(err)

		err = wait.For(conditions.New(client).ResourceMatch(
			configMap, func(object k8s.Object) bool {
				obj := object.(*v1.ConfigMap)
				return len(obj.Data) >= 1
			}))
		expect.NoError(err)
	})

	ginkgo.It("secret is restored when removed", func(ctx context.Context) {
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
		}
		err := client.Delete(ctx, secret)
		expect.NoError(err)

		err = wait.For(conditions.New(client).ResourceMatch(
			secret, func(object k8s.Object) bool {
				obj := object.(*v1.Secret)
				return len(obj.Data) >= 1
			}))
		expect.NoError(err)
	})

	ginkgo.It("service monitor is restored when removed", func(ctx context.Context) {
		serviceMonitor := &monitoringv1.ServiceMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceMonitorName,
				Namespace: namespace,
			},
		}
		err := client.Delete(ctx, serviceMonitor)
		expect.NoError(err)

		err = wait.For(conditions.New(client).ResourceMatch(
			serviceMonitor, func(object k8s.Object) bool {
				obj := object.(*monitoringv1.ServiceMonitor)
				return obj.ObjectMeta.Generation == 1
			}))
		expect.NoError(err)
	})

	ginkgo.It("service is restored when removed", func(ctx context.Context) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: namespace,
			},
		}
		err := client.Delete(ctx, service)
		expect.NoError(err)

		err = wait.For(conditions.New(client).ResourceMatch(
			service, func(object k8s.Object) bool {
				obj := object.(*v1.Service)
				return obj.ObjectMeta.Name == serviceName
			}))
		expect.NoError(err)
	})

	ginkgo.It("network policy is restored when removed", func(ctx context.Context) {
		networkPolicy := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      networkPolicyName,
				Namespace: namespace,
			},
		}
		err := client.Delete(ctx, networkPolicy)
		expect.NoError(err)

		err = wait.For(conditions.New(client).ResourceMatch(
			networkPolicy, func(object k8s.Object) bool {
				obj := object.(*networkingv1.NetworkPolicy)
				return obj.ObjectMeta.Name == networkPolicyName
			}))
		expect.NoError(err)
	})

	ginkgo.It("can be upgraded from previous version", label.Upgrade, func(ctx context.Context) {
		operators.PerformUpgrade(ctx, h, namespace, operatorName, operatorName, operatorRegistry, 5, 30)
	})
})
