package dvo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachinerylabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

const (
	suiteName      = "Deployment Validation Operator"
	namespaceName  = "openshift-deployment-validation-operator"
	operatorName   = "deployment-validation-operator"
	deploymentName = "deployment-validation-operator"
	serviceName    = "deployment-validation-operator-metrics"
)

func init() {
	alert.RegisterGinkgoAlert(suiteName, "SD-SREP", "@sd-qe", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(suiteName, ginkgo.Ordered, label.E2E, label.ROSA, label.CCS, label.STS, label.AllCloudProviders(), func() {
	var h *helper.H
	var client *resources.Resources

	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.It("exists and is running", label.Install, func(ctx context.Context) {
		clusterRoles := []string{
			"deployment-validation-operator-og-admin",
			"deployment-validation-operator-og-edit",
			"deployment-validation-operator-og-view",
		}

		err := client.Get(ctx, namespaceName, namespaceName, &v1.Namespace{})
		expect.NoError(err)

		err = client.Get(ctx, serviceName, namespaceName, &v1.Service{})
		expect.NoError(err)

		for _, clusterRoleName := range clusterRoles {
			err = client.Get(ctx, clusterRoleName, "", &rbacv1.ClusterRole{})
			expect.NoError(err)
		}

		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespaceName}}
		err = wait.For(conditions.New(client).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, v1.ConditionTrue), wait.WithTimeout(10*time.Second))
		expect.NoError(err)
	})

	ginkgo.It("flags the new deployment", func(ctx context.Context) {
		validationMsg := fmt.Sprintf("\"msg\":\"Set memory requests and limits for your container based on its requirements. Refer to https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits for details.\",\"request.namespace\":\"%s\"", h.CurrentProject())

		labels := map[string]string{"app": "test"}
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "osde2e-", Namespace: h.CurrentProject()},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: labels},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: labels},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: "pause", Image: "registry.k8s.io/pause:latest"},
						},
					},
				},
			},
		}
		err := client.Create(ctx, deployment)
		expect.NoError(err)

		err = wait.For(conditions.New(client).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, v1.ConditionTrue))
		expect.NoError(err)

		// Wait for the deployment logs to contain the validation message
		err = wait.For(func() (bool, error) {
			clientset, err := kubernetes.NewForConfig(client.GetConfig())
			if err != nil {
				return false, err
			}

			pods := &v1.PodList{}
			err = client.List(ctx, pods, resources.WithLabelSelector(apimachinerylabels.FormatLabels(map[string]string{"app": deploymentName})))
			if err != nil {
				return false, err
			}

			if len(pods.Items) < 1 {
				return false, fmt.Errorf("failed to find pod for deployment %s", deploymentName)
			}

			req := clientset.CoreV1().Pods(namespaceName).GetLogs(pods.Items[0].GetName(), &v1.PodLogOptions{})
			logs, err := req.DoRaw(ctx)
			if err != nil {
				return false, err
			}

			return strings.Contains(string(logs), validationMsg), nil
		})
		expect.NoError(err)

		err = client.Delete(ctx, deployment)
		expect.NoError(err)
	})
})
