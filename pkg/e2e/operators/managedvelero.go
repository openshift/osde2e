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

var veleroOperatorTestName string = "[Suite: operators] [OSD] Managed Velero Operator"

func init() {
	alert.RegisterGinkgoAlert(veleroOperatorTestName, "SD-SREP", "Christoph Blecker", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(veleroOperatorTestName, func() {
	var operatorName = "managed-velero-operator"
	var operatorNamespace string = "openshift-velero"
	var operatorLockFile string = "managed-velero-operator-lock"
	var defaultDesiredReplicas int32 = 1
	h := helper.New()

	// NOTE: As this is deployed by OLM now, RBAC objects (ClusterRoles, ClusterRoleBindings, and the primary
	// RoleBinding) have random-ish names like:
	// managed-velero-operator.v0.2.196-2b128e8-5df7c76948
	//
	// Test need to incorporate a regex-like test?

	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkDeployment(h, operatorNamespace, "velero", defaultDesiredReplicas)
	checkRole(h, "kube-system", []string{"cluster-config-v1-reader"})
	checkRoleBindings(h,
		"kube-system",
		[]string{"managed-velero-operator-cluster-config-v1-reader"})
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
