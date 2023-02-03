package webhooks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	quotav1 "github.com/openshift/api/quota/v1"
	securityv1 "github.com/openshift/api/security/v1"
	customdomainv1alpha1 "github.com/openshift/custom-domains-operator/api/v1alpha1"
	mustgatherv1alpha1 "github.com/openshift/must-gather-operator/api/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubectl/pkg/util/slice"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	suiteName = "Managed Cluster Validating Webhooks"
)

func init() {
	alert.RegisterGinkgoAlert(suiteName, "SD-SREP", "@sd-qe", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(suiteName, ginkgo.Ordered, label.E2E, label.ROSA, label.CCS, label.STS, label.AllCloudProviders(), func() {
	var h *helper.H

	ginkgo.BeforeAll(func() {
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip(suiteName + " is not deployed to ROSA hosted-cp clusters (SDE-2252)")
		}
		h = helper.New()
	})

	ginkgo.It("exists and is running", label.Install, func(ctx context.Context) {
		const (
			namespaceName = "openshift-validation-webhook"
			serviceName   = "validation-webhook"
			daemonsetName = "validation-webhook"
			configMapName = "webhook-cert"
			secretName    = "webhook-cert"
		)

		client := h.AsUser("")

		ginkgo.By("checking the namespace exists")
		err := client.Get(ctx, namespaceName, namespaceName, &v1.Namespace{})
		expect.NoError(err)

		ginkgo.By("checking the configmaps exist")
		err = client.Get(ctx, configMapName, namespaceName, &v1.ConfigMap{})
		expect.NoError(err)

		ginkgo.By("checking the secret exists")
		err = client.Get(ctx, secretName, namespaceName, &v1.Secret{})
		expect.NoError(err)

		ginkgo.By("checking the service exists")
		err = client.Get(ctx, serviceName, namespaceName, &v1.Service{})
		expect.NoError(err)

		ginkgo.By("checking the daemonset exists")
		ds := &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: daemonsetName, Namespace: namespaceName}}
		err = wait.For(conditions.New(client).ResourceMatch(ds, func(object k8s.Object) bool {
			d := object.(*appsv1.DaemonSet)
			desiredNumScheduled := d.Status.DesiredNumberScheduled
			return d.Status.CurrentNumberScheduled == desiredNumScheduled &&
				d.Status.NumberReady == desiredNumScheduled &&
				d.Status.NumberAvailable == desiredNumScheduled
		}))
		expect.NoError(err)
	})

	ginkgo.Describe("sre-pod-validation", ginkgo.Ordered, func() {
		const (
			privilegedNamespace   = "openshift-backplane"
			unprivilegedNamespace = "openshift-logging"

			deletePodWaitDuration = 5 * time.Minute
			createPodWaitDuration = 1 * time.Minute
		)
		var pod *v1.Pod
		newTestPod := func(name string) *v1.Pod {
			return &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "test",
							Image: "registry.access.redhat.com/ubi8/ubi-minimal",
						},
					},
					Tolerations: []v1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Value:  "toleration-key-value",
							Effect: v1.TaintEffectNoSchedule,
						}, {
							Key:    "node-role.kubernetes.io/infra",
							Value:  "toleration-key-value2",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			}
		}

		withNamespace := func(pod *v1.Pod, namespace string) *v1.Pod {
			pod.SetNamespace(namespace)
			return pod
		}

		ginkgo.BeforeAll(func() {
			name := envconf.RandomName("osde2e", 12)
			pod = newTestPod(name)
		})

		ginkgo.It("blocks pods scheduled onto master/infra nodes", func(ctx context.Context) {
			err := h.AsDedicatedAdmin().Create(ctx, withNamespace(pod, privilegedNamespace))
			expect.Forbidden(err)

			client := h.AsUser("majora")
			err = client.Create(ctx, withNamespace(pod, privilegedNamespace))
			expect.Forbidden(err)

			err = client.Create(ctx, withNamespace(pod, unprivilegedNamespace))
			expect.Forbidden(err)
		}, ginkgo.SpecTimeout(createPodWaitDuration.Seconds()+deletePodWaitDuration.Seconds()))

		ginkgo.It("allows cluster-admin to schedule pods onto master/infra nodes", func(ctx context.Context) {
			client := h.AsServiceAccount(fmt.Sprintf("system:serviceaccount:%s:dedicated-admin-project", h.CurrentProject()))
			pod = withNamespace(pod, privilegedNamespace)
			err := client.Create(ctx, pod)
			expect.NoError(err)
			err = client.Delete(ctx, pod)
			expect.NoError(err)
		}, ginkgo.SpecTimeout(createPodWaitDuration.Seconds()+deletePodWaitDuration.Seconds()))

		ginkgo.It("prevents workloads from being scheduled on worker nodes", func(ctx context.Context) {
			client := h.AsUser("")
			operators := map[string]string{
				"cloud-ingress-operator":          "openshift-cloud-ingress-operator",
				"configure-alertmanager-operator": "openshift-monitoring",
				"custom-domains-operator":         "openshift-custom-domains-operator",
				"managed-upgrade-operator":        "openshift-managed-upgrade-operator",
				"managed-velero-operator":         "openshift-velero",
				"must-gather-operator":            "openshift-must-gather-operator",
				"osd-metrics-exporter":            "openshift-osd-metrics",
				"rbac-permissions-operator":       "openshift-rbac-permissions",
			}

			var podList v1.PodList
			expect.NoError(client.WithNamespace(metav1.NamespaceAll).List(ctx, &podList), "unable to list pods")
			Expect(len(podList.Items)).To(BeNumerically(">", 0), "found no pods")

			var nodeList v1.NodeList
			selectInfraNodes := resources.WithLabelSelector(labels.FormatLabels(map[string]string{"node-role.kubernetes.io": "infra"}))
			expect.NoError(client.List(ctx, &nodeList), selectInfraNodes)

			nodeNames := []string{}
			for _, node := range nodeList.Items {
				nodeNames = append(nodeNames, node.GetName())
			}

			violators := []string{}
			for _, pod := range podList.Items {
				for operator, namespace := range operators {
					if pod.GetNamespace() != namespace {
						continue
					}
					if strings.HasPrefix(pod.GetName(), operator) && !strings.HasPrefix(pod.GetName(), operator+"-registry") {
						if !slice.ContainsString(nodeNames, pod.Spec.NodeName, nil) {
							violators = append(violators, pod.GetNamespace()+"/"+pod.GetName())
						}
					}
				}
			}

			Expect(violators).To(HaveLen(0), "found pods in violation %v", violators)
		})
	})

	ginkgo.Describe("sre-techpreviewnoupgrade-validation", func() {
		ginkgo.It("blocks customers from setting TechPreviewNoUpgrade feature gate", func(ctx context.Context) {
			client := h.AsClusterAdmin()
			clusterFeatureGate := &configv1.FeatureGate{}
			err := client.Get(ctx, "cluster", "", clusterFeatureGate)
			expect.NoError(err)
			clusterFeatureGate.Spec.FeatureSet = "TechPreviewNoUpgrade"
			err = client.Update(ctx, clusterFeatureGate)
			expect.Forbidden(err)
		})
	})

	ginkgo.Describe("sre-regular-user-validation", func() {
		ginkgo.It("blocks unauthenticated users from managing \"managed\" resources", func(ctx context.Context) {
			client := h.AsUser("system:unauthenticated")
			cvo := &configv1.ClusterVersion{ObjectMeta: metav1.ObjectMeta{Name: "osde2e-version"}}
			err := client.Create(ctx, cvo)
			expect.Forbidden(err)
		})

		ginkgo.DescribeTable(
			"allows privileged users to manage \"managed\" resources",
			func(ctx context.Context, user string) {
				client := h.AsUser(user)
				cvo := &configv1.ClusterVersion{ObjectMeta: metav1.ObjectMeta{Name: "osde2e-version"}}
				err := client.Create(ctx, cvo)
				expect.NoError(err)
				err = client.Delete(ctx, cvo)
				expect.NoError(err)
			},
			ginkgo.Entry("as system:admin", "system:admin"),
			ginkgo.Entry("as backplane-cluster-admin", "backplane-cluster-admin"),
		)

		ginkgo.It("only blocks configmap/user-ca-bundle changes", func(ctx context.Context) {
			client := h.AsDedicatedAdmin()
			cm := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "user-ca-bundle", Namespace: "openshift-config"}}
			err := client.Delete(ctx, cm)
			expect.Forbidden(err)

			cm = &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: h.CurrentProject()},
				Data:       map[string]string{"test": "test"},
			}
			err = client.Create(ctx, cm)
			expect.NoError(err)
			err = client.Delete(ctx, cm)
			expect.NoError(err)
		})

		ginkgo.It("blocks modifications to nodes", func(ctx context.Context) {
			client := h.AsDedicatedAdmin()
			var nodes v1.NodeList
			selectInfraNodes := resources.WithLabelSelector(labels.FormatLabels(map[string]string{"node-role.kubernetes.io": "infra"}))
			err := client.List(ctx, &nodes, selectInfraNodes)
			expect.NoError(err)
			Expect(len(nodes.Items)).Should(BeNumerically(">", 0), "failed to find infra nodes")

			node := nodes.Items[0]
			node.SetLabels(map[string]string{"osde2e": ""})
			err = client.Update(ctx, &node)
			expect.Forbidden(err)
		})

		// TODO: test "system:serviceaccounts:openshift-backplane-cee" group can use NetNamespace CR

		ginkgo.It("allows dedicated-admin to manage CustomDomain CRs", func(ctx context.Context) {
			client := h.AsDedicatedAdmin()
			cd := &customdomainv1alpha1.CustomDomain{ObjectMeta: metav1.ObjectMeta{Name: "test-cd", Namespace: h.CurrentProject()}}
			err := client.Create(ctx, cd)
			expect.NoError(err)
			err = client.Delete(ctx, cd)
			expect.NoError(err)
		})

		ginkgo.It("allows backplane-cluster-admin to manage MustGather CRs", func(ctx context.Context) {
			client := h.AsUser("backplane-cluster-admin", "system:serviceaccounts:backplane-cluster-admin")
			mg := &mustgatherv1alpha1.MustGather{ObjectMeta: metav1.ObjectMeta{Name: "test-mg", Namespace: h.CurrentProject()}}
			err := client.Create(ctx, mg)
			expect.NoError(err)
			err = client.Delete(ctx, mg)
			expect.NoError(err)
		})
	})

	ginkgo.Describe("sre-hiveownership-validation", ginkgo.Ordered, func() {
		const quotaName = "-quota"
		var managedCRQ *quotav1.ClusterResourceQuota
		newTestCRQ := func(name string) *quotav1.ClusterResourceQuota {
			managed := strings.HasPrefix(name, "managed")
			return &quotav1.ClusterResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name:   name,
					Labels: map[string]string{"hive.openshift.io/managed": strconv.FormatBool(managed)},
				},
				Spec: quotav1.ClusterResourceQuotaSpec{
					Selector: quotav1.ClusterResourceQuotaSelector{
						AnnotationSelector: map[string]string{"openshift.io/requester": "test"},
					},
				},
			}
		}

		ginkgo.BeforeAll(func(ctx context.Context) {
			asAdmin := h.AsClusterAdmin()
			managedCRQ = newTestCRQ("managed" + quotaName)
			err := asAdmin.Create(ctx, managedCRQ)
			expect.NoError(err)
		})

		ginkgo.AfterAll(func(ctx context.Context) {
			asAdmin := h.AsClusterAdmin()
			err := asAdmin.Delete(ctx, managedCRQ)
			expect.NoError(err)
		})

		ginkgo.It("blocks deletion of managed ClusterResourceQuotas", func(ctx context.Context) {
			client := h.AsDedicatedAdmin()
			err := client.Delete(ctx, managedCRQ)
			expect.Forbidden(err)

			client = h.AsUser("majora")
			err = client.Delete(ctx, managedCRQ)
			expect.Forbidden(err)
		})

		ginkgo.It("allows a member of SRE to update managed ClusterResourceQuotas", func(ctx context.Context) {
			client := h.AsUser("backplane-cluster-admin")
			managedCRQ.SetLabels(map[string]string{"osde2e": ""})
			err := client.Update(ctx, managedCRQ)
			expect.NoError(err)
		})

		ginkgo.It("allows dedicated-admins can manage unmanaged ClusterResourceQuotas", func(ctx context.Context) {
			client := h.AsDedicatedAdmin()
			unmanagedCRQ := newTestCRQ("openshift" + quotaName)

			err := client.Create(ctx, unmanagedCRQ)
			expect.NoError(err)

			unmanagedCRQ.SetLabels(map[string]string{"osde2e": ""})
			err = client.Update(ctx, unmanagedCRQ)
			expect.NoError(err)

			err = client.Delete(ctx, unmanagedCRQ)
			expect.NoError(err)
		})
	})

	ginkgo.Describe("sre-scc-validation", func() {
		ginkgo.It("blocks modifications to default SecurityContextConstraints", func(ctx context.Context) {
			client := h.AsDedicatedAdmin()
			scc := &securityv1.SecurityContextConstraints{ObjectMeta: metav1.ObjectMeta{Name: "privileged"}}
			scc.SetLabels(map[string]string{"osde2e": ""})

			err := client.Update(ctx, scc)
			expect.Forbidden(err)

			err = client.Delete(ctx, scc)
			expect.Forbidden(err)
		})
	})

	ginkgo.Describe("sre-namespace-validation", ginkgo.Ordered, func() {
		const testUser = "testuser@testdomain.com"
		const nonPrivilegedNamespace = "mykube-admin"

		// Map of namespace name and whether it should be created/deleted by the test
		// Should match up with namespaces found in managed-cluster-config:
		// * https://github.com/openshift/managed-cluster-config/blob/master/deploy/osd-managed-resources/ocp-namespaces.ConfigMap.yaml
		// * https://github.com/openshift/managed-cluster-config/blob/master/deploy/osd-managed-resources/managed-namespaces.ConfigMap.yaml
		privilegedNamespaces := map[string]bool{
			"default":                        false,
			"redhat-ocm-addon-test-operator": true,
		}
		privilegedUsers := []string{
			"system:admin",
			"backplane-cluster-admin",
		}

		createNamespace := func(ctx context.Context, name string) {
			client := h.AsClusterAdmin()
			err := client.Get(ctx, name, "", &v1.Namespace{})
			if apierrors.IsNotFound(err) {
				err = client.Create(ctx, &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
						Labels: map[string]string{
							"pod-security.kubernetes.io/enforce":             "privileged",
							"pod-security.kubernetes.io/audit":               "privileged",
							"pod-security.kubernetes.io/warn":                "privileged",
							"security.openshift.io/scc.podSecurityLabelSync": "false",
						},
					},
				})
			}
			expect.NoError(err)
		}

		deleteNamespace := func(ctx context.Context, name string) {
			client := h.AsClusterAdmin()
			err := client.Delete(ctx, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}})
			expect.NoError(err)
		}

		updateNamespace := func(ctx context.Context, h *helper.H, name string, user string, groups ...string) error {
			client := h.AsUser(user, groups...)
			ns := &v1.Namespace{}
			err := client.Get(ctx, name, "", ns)
			if err != nil {
				return err
			}
			return client.Update(ctx, ns)
		}

		ginkgo.BeforeAll(func(ctx context.Context) {
			for namespace, create := range privilegedNamespaces {
				if create {
					createNamespace(ctx, namespace)
				}
			}
			createNamespace(ctx, nonPrivilegedNamespace)
		})

		ginkgo.AfterAll(func(ctx context.Context) {
			for namespace, dlt := range privilegedNamespaces {
				if dlt {
					deleteNamespace(ctx, namespace)
				}
			}
			deleteNamespace(ctx, nonPrivilegedNamespace)
		})

		ginkgo.It("blocks dedicated admins from managing privileged namespaces", func(ctx context.Context) {
			for namespace := range privilegedNamespaces {
				err := updateNamespace(ctx, h, namespace, testUser, "dedicated-admins")
				expect.Forbidden(err)
			}
		})

		ginkgo.It("block non privileged users from managing privileged namespaces", func(ctx context.Context) {
			for namespace := range privilegedNamespaces {
				err := updateNamespace(ctx, h, namespace, testUser)
				expect.Forbidden(err)
			}
		})

		ginkgo.It("allows privileged users to manage all namespaces", func(ctx context.Context) {
			for _, user := range privilegedUsers {
				for namespace := range privilegedNamespaces {
					err := updateNamespace(ctx, h, namespace, user)
					expect.NoError(err)
				}

				err := updateNamespace(ctx, h, nonPrivilegedNamespace, user)
				expect.NoError(err)
			}
		})

		ginkgo.It("allows non privileged users to manage non privileged namespaces", func(ctx context.Context) {
			err := updateNamespace(ctx, h, nonPrivilegedNamespace, testUser, "dedicated-admins")
			expect.NoError(err)
		})
	})

	ginkgo.Describe("sre-prometheusrule-validation", func() {
		const privilegedNamespace = "openshift-backplane"

		newPrometheusRule := func(namespace string) *monitoringv1.PrometheusRule {
			return &monitoringv1.PrometheusRule{
				ObjectMeta: metav1.ObjectMeta{Name: "prometheus-example-app", Namespace: namespace},
				Spec: monitoringv1.PrometheusRuleSpec{
					Groups: []monitoringv1.RuleGroup{
						{
							Name: "example",
							Rules: []monitoringv1.Rule{
								{
									Alert: "VersionAlert",
									Expr:  intstr.FromString("version{job=\"prometheus-example-app\"} == 0"),
								},
							},
						},
					},
				},
			}
		}

		ginkgo.DescribeTable(
			"blocks users from creating PrometheusRules in privileged namespaces",
			func(ctx context.Context, user string) {
				client := h.AsUser(user)
				rule := newPrometheusRule(privilegedNamespace)
				err := client.Create(ctx, rule)
				expect.Forbidden(err)
			},
			ginkgo.Entry("as dedicated-admin", "dedicated-admin"),
			ginkgo.Entry("as random user", "majora"),
		)

		ginkgo.It("allows backplane-cluster-admin to manage PrometheusRules in all namespaces", func(ctx context.Context) {
			client := h.AsUser("backplane-cluster-admin")
			rule := newPrometheusRule(privilegedNamespace)
			err := client.Create(ctx, rule)
			expect.NoError(err)
			err = client.Delete(ctx, rule)
			expect.NoError(err)

			rule = newPrometheusRule(h.CurrentProject())
			err = client.Create(ctx, rule)
			expect.NoError(err)
			err = client.Delete(ctx, rule)
			expect.NoError(err)
		})

		ginkgo.It("allows non privileged users to manage PrometheusRules in non privileged namespaces", func(ctx context.Context) {
			client := h.AsDedicatedAdmin()
			rule := newPrometheusRule(h.CurrentProject())
			err := client.Create(ctx, rule)
			expect.NoError(err)
			err = client.Delete(ctx, rule)
			expect.NoError(err)
		})
	})
})
