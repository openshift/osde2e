package configurealertmanager

import (
	"context"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/e2e/operators"

	operatorhubv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

var suiteName = "Configure AlertManager Operator"

func init() {
	alert.RegisterGinkgoAlert(suiteName, "SD-SREP", "@sd-srep-team-thor", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(suiteName, ginkgo.Ordered, label.Operators, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("Configure AlertManager Operator is not supported on HyperShift")
		}
	})

	var h *helper.H
	var client *resources.Resources

	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	var (
		configMapLockFile = "configure-alertmanager-operator-lock"
		namespaceName     = "openshift-monitoring"
		operatorName      = "configure-alertmanager-operator"
		operatorRegistry  = "configure-alertmanager-operator-registry"
		secrets           = []string{"pd-secret", "dms-secret"}
		serviceAccounts   = []string{"configure-alertmanager-operator"}
	)

	ginkgo.It("cluster service version exists", label.Install, func(ctx context.Context) {
		var csvs operatorhubv1.ClusterServiceVersionList
		err := client.List(ctx, &csvs, resources.WithFieldSelector(
			labels.FormatLabels(map[string]string{"metadata.namespace": namespaceName})))
		expect.NoError(err, "failed to get cluster service versions")

		for _, csv := range csvs.Items {
			if csv.Spec.DisplayName == operatorName {
				err = wait.For(conditions.New(client).ResourceMatch(
					&operatorhubv1.ClusterServiceVersion{
						ObjectMeta: metav1.ObjectMeta{
							Name:      csv.Name,
							Namespace: namespaceName,
						},
					}, func(object k8s.Object) bool {
						obj := object.(*operatorhubv1.ClusterServiceVersion)
						return obj.Status.Phase == "Succeeded"
					}))
				expect.NoError(err, "csv %s not in a succeeded phase", csv.Spec.DisplayName)
				break
			}
		}
	})

	ginkgo.It("service accounts exist", label.Install, func(ctx context.Context) {
		for _, serviceAccount := range serviceAccounts {
			err := client.Get(ctx, serviceAccount, namespaceName, &v1.ServiceAccount{})
			expect.NoError(err, "service account %s not found", serviceAccount)
		}
	})

	ginkgo.It("deployment exist", label.Install, func(ctx context.Context) {
		err := wait.For(conditions.New(client).DeploymentConditionMatch(
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      operatorName,
					Namespace: namespaceName,
				},
			}, appsv1.DeploymentAvailable, v1.ConditionTrue))
		expect.NoError(err, "deployment %s not available", operatorName)

	})

	ginkgo.It("roles exist", label.Install, func(ctx context.Context) {
		var roles rbacv1.RoleList
		err := client.List(ctx, &roles)
		expect.NoError(err, "failed to get roles")
		found := false
		for _, role := range roles.Items {
			if strings.HasPrefix(role.Name, operatorName) {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "roles not found")
	})

	ginkgo.It("role bindings exist", label.Install, func(ctx context.Context) {
		var roleBindings rbacv1.RoleBindingList
		err := client.List(ctx, &roleBindings)
		expect.NoError(err, "failed to get role bindings")
		found := false
		for _, roleBinding := range roleBindings.Items {
			if strings.HasPrefix(roleBinding.Name, operatorName) {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "role bindings not found")
	})

	ginkgo.It("cluster roles exist", label.Install, func(ctx context.Context) {
		var clusterRoles rbacv1.ClusterRoleList
		err := client.List(ctx, &clusterRoles)
		expect.NoError(err, "failed to get cluster roles")
		found := false
		for _, clusterRole := range clusterRoles.Items {
			if strings.HasPrefix(clusterRole.Name, operatorName) {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "cluster roles not found")
	})

	ginkgo.It("cluster role bindings exist", label.Install, func(ctx context.Context) {
		var clusterRoleBindings rbacv1.ClusterRoleBindingList
		err := client.List(ctx, &clusterRoleBindings)
		expect.NoError(err, "failed to get cluster role bindings")
		found := false
		for _, clusterRoleBinding := range clusterRoleBindings.Items {
			if strings.HasPrefix(clusterRoleBinding.Name, operatorName) {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "cluster role bindings not found")
	})

	ginkgo.It("config map exists", label.Install, func(ctx context.Context) {
		err := client.Get(ctx, configMapLockFile, namespaceName, &v1.ConfigMap{})
		expect.NoError(err, "failed to get config map %s", configMapLockFile)
	})

	ginkgo.It("secrets exist", label.Install, func(ctx context.Context) {
		for _, secret := range secrets {
			err := client.Get(ctx, secret, namespaceName, &v1.Secret{})
			expect.NoError(err, "secret %s not found", secret)
		}
	})

	ginkgo.It("can be upgraded from previous version", label.Upgrade, func(ctx context.Context) {
		errorMsg, err := operators.PerformUpgrade(ctx, h, namespaceName, operatorName, operatorName, operatorRegistry, 5, 30)
		expect.NoError(err, errorMsg)
	})
})
