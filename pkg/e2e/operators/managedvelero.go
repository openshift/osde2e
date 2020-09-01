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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	checkRole(h, "kube-system", []string{"cluster-config-v1-reader"})
	checkRoleBindings(h,
		"kube-system",
		[]string{"managed-velero-operator-cluster-config-v1-reader"})
	checkBackups(h)
	testDaCRbackups(h)
	testDaCRrestore(h)
	testDaCRdeleteBackupRequests(h)
	testDaCRbackupStorageLocations(h)
	testDaCRdownloadRequests(h)
	testDaCRpodVolumeBackup(h)
	testDaCRpodVolumeRestores(h)
	testDaCRvolumeSnapshotLocation(h)
	testDaCRschedules(h)
	testDaCRserverStatusRequest(h)
	testDaCRrestricRepository(h)
	testCRbackups(h)
	testCRrestore(h)
	testCRdeleteBackupRequests(h)
	testCRbackupStorageLocations(h)
	testCRdownloadRequests(h)
	testCRpodVolumeBackup(h)
	testCRpodVolumeRestores(h)
	testCRvolumeSnapshotLocation(h)
	testCRschedules(h)
	testCRserverStatusRequest(h)
	testCRrestricRepository(h)
	checkClusterRoles(h, clusterRoles, true)
	checkClusterRoleBindings(h, clusterRoleBindings, true)

})

//tests to confirm dedicated-admin user can not edit CRs

func testDaCRbackups(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit Backups", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			backup := velerov1.Backup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-backup-test",
				},
			}
			_, err := h.Velero().VeleroV1().Backups(h.CurrentProject()).Create(context.TODO(), &backup, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})
	})
}

func testDaCRrestore(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit Restore", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			restore := velerov1.Restore{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-restore-test",
				},
			}
			_, err := h.Velero().VeleroV1().Restore(h.CurrentProject()).Create(context.TODO(), &restore, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})
	})
}

func testDaCRdeleteBackupRequests(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit DeleteBackupRequests", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			deleteBackupRequest := velerov1.DeleteBackupRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-delete-backup-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().DeleteBackupRequests(h.CurrentProject()).Create(context.TODO(), &deleteBackupRequest, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRbackupStorageLocations(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit BackupStorageLocations", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			backupStorageLocation := velerov1.BackupStorageLocation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-backup-storage-location-test",
				},
			}
			_, err := h.Velero().VeleroV1().BackupStorageLocations(h.CurrentProject()).Create(context.TODO(), &backupStorageLocation, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRdownloadRequests(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit DownloadRequests", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			downloadRequest := velerov1.DownloadRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-download-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().DownloadRequests(h.CurrentProject()).Create(context.TODO(), &downloadRequest, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRpodVolumeBackup(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit PodVolumeBackups", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			podVolumeBackup := velerov1.PodVolumeBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-pod-volume-backup-test",
				},
			}
			_, err := h.Velero().VeleroV1().PodVolumeBackups(h.CurrentProject()).Create(context.TODO(), &podVolumeBackup, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRpodVolumeRestores(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit PodVolumeRestores", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			podVolumeRestore := velerov1.PodVolumeRestore{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-pod-volume-restore-test",
				},
			}
			_, err := h.Velero().VeleroV1().PodVolumeRestores(h.CurrentProject()).Create(context.TODO(), &podVolumeRestore, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})

	})
}

func testDaCRvolumeSnapshotLocation(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit VolumeSnapshotLocations", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			volumeSnapshotLocation := velerov1.VolumeSnapshotLocation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-volume-snapshot-locations-test",
				},
			}
			_, err := h.Velero().VeleroV1().VolumeSnapshotLocation(h.CurrentProject()).Create(context.TODO(), &volumeSnapshotLocation, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRschedules(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit Schedules", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			schedules := velerov1.Schedules{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-schedules-test",
				},
			}
			_, err := h.Velero().VeleroV1().Schedules(h.CurrentProject()).Create(context.TODO(), &schedules, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRserverStatusRequest(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit ServerStatusRequests", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			serverStatusRequest := velerov1.ServerStatusRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-server-status-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().ServerStatusRequest(h.CurrentProject()).Create(context.TODO(), &serverStatusRequest, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRrestricRepository(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be forbidden to edit RestricRepository", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")

			resticRepository := velerov1.ResticRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-restric-repository-test",
				},
			}
			_, err := h.Velero().VeleroV1().ResticRepository(h.CurrentProject()).Create(context.TODO(), &resticRepository, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

//test to confirm admin user can edit CRs

func testCRbackups(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit Backups", func() {
			backup := velerov1.Backup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "admin-backup-test",
				},
			}
			_, err := h.Velero().VeleroV1().Backups(h.CurrentProject()).Create(context.TODO(), &backup, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

		})
	})
}

func testCRrestore(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit Restore", func() {

			restore := velerov1.Restore{
				ObjectMeta: metav1.ObjectMeta{
					Name: "restore-test",
				},
			}
			_, err := h.Velero().VeleroV1().Restore(h.CurrentProject()).Create(context.TODO(), &restore, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

		})
	})
}

func testCRdeleteBackupRequests(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit DeleteBackupRequests", func() {

			deleteBackupRequest := velerov1.DeleteBackupRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "delete-backup-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().DeleteBackupRequests(h.CurrentProject()).Create(context.TODO(), &deleteBackupRequest, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRbackupStorageLocations(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit BackupStorageLocations", func() {

			backupStorageLocation := velerov1.BackupStorageLocation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup-storage-location-test",
				},
			}
			_, err := h.Velero().VeleroV1().BackupStorageLocations(h.CurrentProject()).Create(context.TODO(), &backupStorageLocation, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRdownloadRequests(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit DownloadRequests", func() {

			downloadRequest := velerov1.DownloadRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "download-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().DownloadRequests(h.CurrentProject()).Create(context.TODO(), &downloadRequest, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRpodVolumeBackup(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit PodVolumeBackups", func() {

			podVolumeBackup := velerov1.PodVolumeBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-volume-backup-test",
				},
			}
			_, err := h.Velero().VeleroV1().PodVolumeBackups(h.CurrentProject()).Create(context.TODO(), &podVolumeBackup, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRpodVolumeRestores(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit PodVolumeRestores", func() {

			podVolumeRestore := velerov1.PodVolumeRestore{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-volume-restore-test",
				},
			}
			_, err := h.Velero().VeleroV1().PodVolumeRestores(h.CurrentProject()).Create(context.TODO(), &podVolumeRestore, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

	})
}

func testCRvolumeSnapshotLocation(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit VolumeSnapshotLocations", func() {

			volumeSnapshotLocation := velerov1.VolumeSnapshotLocation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "volume-snapshot-locations-test",
				},
			}
			_, err := h.Velero().VeleroV1().VolumeSnapshotLocation(h.CurrentProject()).Create(context.TODO(), &volumeSnapshotLocation, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRschedules(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit Schedules", func() {

			schedules := velerov1.Schedules{
				ObjectMeta: metav1.ObjectMeta{
					Name: "schedules-test",
				},
			}
			_, err := h.Velero().VeleroV1().Schedules(h.CurrentProject()).Create(context.TODO(), &schedules, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRserverStatusRequest(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit ServerStatusRequests", func() {

			serverStatusRequest := velerov1.ServerStatusRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "server-status-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().ServerStatusRequest(h.CurrentProject()).Create(context.TODO(), &serverStatusRequest, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRrestricRepository(h *helper.H) {
	ginkgo.Context("velero", func() {
		ginkgo.It("Access should be allowed to edit RestricRepository", func() {

			resticRepository := velerov1.ResticRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "restric-repository-test",
				},
			}
			_, err := h.Velero().VeleroV1().ResticRepository(h.CurrentProject()).Create(context.TODO(), &resticRepository, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

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
