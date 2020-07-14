package operators

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: operators] [OSD] Managed Velero Operator",
		TeamOwner:        "SD-SREP",
		PrimaryContact:   "Christoph Blecker",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 1,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name, func() {
	var operatorName = "managed-velero-operator"
	var operatorNamespace string = "openshift-velero"
	var operatorLockFile string = "managed-velero-operator-lock"
	var defaultDesiredReplicas int32 = 1
	var clusterRoles = []string{
		"managed-velero-operator",
	}
	var clusterRoleBindings = []string{
		"managed-velero-operator",
		"velero",
	}
	h := helper.New()
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkDeployment(h, operatorNamespace, "velero", defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles)
	checkClusterRoleBindings(h, clusterRoleBindings)
	checkRoleBindings(h,
		operatorNamespace,
		[]string{"managed-velero-operator"})
	checkVeleroBackups(h)
})

func checkVeleroBackups(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("backups should be complete", func() {
			err := wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
				us, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group:    "velero.io",
					Version:  "v1",
					Resource: "backups"}).Namespace(operatorNamespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					return false, fmt.Errorf("Error getting backups: %s", err.Error())
				}
				for _, item := range us.Items {
					var backup velerov1.Backup
					err = runtime.DefaultUnstructuredConverter.
						FromUnstructured(item.UnstructuredContent(), &backup)
					if err != nil {
						return false, fmt.Errorf("Error casting object: %s", err.Error())
					}
					if backup.Status.Phase == velerov1.BackupPhaseFailed {
						return false, fmt.Errorf("Backup failed for %s", backup.Name)
					}
					if backup.Status.Phase != velerov1.BackupPhaseCompleted {
						return false, nil
					}
				}
				return true, nil
			})
			Expect(err).ToNot(HaveOccurred(), "Backups failed to complete.")
		})
	})
}
