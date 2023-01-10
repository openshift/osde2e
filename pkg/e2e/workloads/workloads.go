package workloads

import (
	"context"
	"io/fs"
	"os"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

const suiteName = "[Suite: e2e] Workloads"

func init() {
	alert.RegisterGinkgoAlert(suiteName, "SD-SREP", "@sd-qe", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(suiteName, ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var h *helper.H

	ginkgo.BeforeAll(func() {
		h = helper.New()
	})

	ginkgo.DescribeTable("can be created, used, and deleted", func(ctx context.Context, workloadFS fs.FS, deploymentName, serviceName string) {
		client := h.AsUser("")

		err := decoder.DecodeEachFile(ctx, workloadFS, "*", decoder.CreateHandler(client), decoder.MutateNamespace(h.CurrentProject()))
		expect.NoError(err)

		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: h.CurrentProject()}}
		err = wait.For(conditions.New(client).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, v1.ConditionTrue), wait.WithTimeout(60*time.Second))
		expect.NoError(err)

		clientset, err := kubernetes.NewForConfig(client.GetConfig())
		expect.NoError(err)

		req := clientset.CoreV1().Services(h.CurrentProject()).ProxyGet("http", serviceName, "3000", "/", nil)
		_, err = req.DoRaw(ctx)
		expect.Error(err)

		err = decoder.DecodeEachFile(ctx, workloadFS, "*", decoder.DeleteHandler(client), decoder.MutateNamespace(h.CurrentProject()))
		expect.NoError(err)
	},
		ginkgo.Entry("Redmine", os.DirFS("assets/workloads/e2e/redmine"), "redmine", "redmine-frontend"),
		ginkgo.Entry("Guestbook", os.DirFS("assets/workloads/e2e/guestbook"), "frontend", "frontend"),
	)
})
