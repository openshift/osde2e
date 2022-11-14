package operators

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

var veleroOperatorTestName string = "[Suite: operators] [OSD] Managed Velero Operator"

func init() {
	alert.RegisterGinkgoAlert(veleroOperatorTestName, "SD-SREP", "@managed-velero-operator", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(veleroOperatorTestName, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool("rosa.STS") {
			ginkgo.Skip("STS does not support MVO")
		}
	})
	operatorName := "managed-velero-operator"
	var operatorNamespace string = "openshift-velero"
	var operatorLockFile string = "managed-velero-operator-lock"
	var defaultDesiredReplicas int32 = 1
	clusterRoles := []string{
		"managed-velero-operator",
	}
	clusterRoleBindings := []string{
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
	checkVeleroBackups(h)
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

// tests to confirm dedicated-admin user can not edit CRs

func testDaCRbackups(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit Backups", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			backup := velerov1.Backup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-backup-test",
				},
			}
			_, err := h.Velero().VeleroV1().Backups(h.CurrentProject()).Create(ctx, &backup, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRrestore(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit Restore", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			restore := velerov1.Restore{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-restore-test",
				},
			}
			_, err := h.Velero().VeleroV1().Restores(h.CurrentProject()).Create(ctx, &restore, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRdeleteBackupRequests(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit DeleteBackupRequests", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			deleteBackupRequest := velerov1.DeleteBackupRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-delete-backup-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().DeleteBackupRequests(h.CurrentProject()).Create(ctx, &deleteBackupRequest, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRbackupStorageLocations(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit BackupStorageLocations", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			backupStorageLocation := velerov1.BackupStorageLocation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-backup-storage-location-test",
				},
				Spec: velerov1.BackupStorageLocationSpec{
					Provider: "aws",
					StorageType: velerov1.StorageType{
						ObjectStorage: &velerov1.ObjectStorageLocation{
							Bucket: "bucket",
							Prefix: "prefix",
						},
					},
				},
			}
			_, err := h.Velero().VeleroV1().BackupStorageLocations(h.CurrentProject()).Create(ctx, &backupStorageLocation, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRdownloadRequests(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit DownloadRequests", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			downloadRequest := velerov1.DownloadRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-download-request-test",
				},
				Spec: velerov1.DownloadRequestSpec{
					Target: velerov1.DownloadTarget{
						Kind: velerov1.DownloadTargetKindBackupContents,
						Name: "targetName",
					},
				},
			}
			_, err := h.Velero().VeleroV1().DownloadRequests(h.CurrentProject()).Create(ctx, &downloadRequest, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRpodVolumeBackup(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit PodVolumeBackups", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			podVolumeBackup := velerov1.PodVolumeBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-pod-volume-backup-test",
				},
			}
			_, err := h.Velero().VeleroV1().PodVolumeBackups(h.CurrentProject()).Create(ctx, &podVolumeBackup, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRpodVolumeRestores(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit PodVolumeRestores", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			podVolumeRestore := velerov1.PodVolumeRestore{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-pod-volume-restore-test",
				},
			}
			_, err := h.Velero().VeleroV1().PodVolumeRestores(h.CurrentProject()).Create(ctx, &podVolumeRestore, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRvolumeSnapshotLocation(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit VolumeSnapshotLocations", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			volumeSnapshotLocation := velerov1.VolumeSnapshotLocation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-volume-snapshot-locations-test",
				},
			}
			_, err := h.Velero().VeleroV1().VolumeSnapshotLocations(h.CurrentProject()).Create(ctx, &volumeSnapshotLocation, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRschedules(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit Schedules", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			schedules := velerov1.Schedule{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-schedules-test",
				},
			}
			_, err := h.Velero().VeleroV1().Schedules(h.CurrentProject()).Create(ctx, &schedules, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRserverStatusRequest(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit ServerStatusRequests", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			serverStatusRequest := velerov1.ServerStatusRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-server-status-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().ServerStatusRequests(h.CurrentProject()).Create(ctx, &serverStatusRequest, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

func testDaCRrestricRepository(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be forbidden to edit RestricRepository", func(ctx context.Context) {
			h.SetServiceAccount(ctx, "system:serviceaccount:%s:dedicated-admin-project")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			resticRepository := velerov1.ResticRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dedicated-admin-restric-repository-test",
				},
			}
			_, err := h.Velero().VeleroV1().ResticRepositories(h.CurrentProject()).Create(ctx, &resticRepository, metav1.CreateOptions{})
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
}

// test to confirm admin user can edit CRs

func testCRbackups(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit Backups", func(ctx context.Context) {
			backup := velerov1.Backup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "admin-backup-test",
				},
			}
			_, err := h.Velero().VeleroV1().Backups(h.CurrentProject()).Create(ctx, &backup, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().Backups(h.CurrentProject()).Delete(ctx, "admin-backup-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRrestore(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit Restore", func(ctx context.Context) {
			restore := velerov1.Restore{
				ObjectMeta: metav1.ObjectMeta{
					Name: "restore-test",
				},
			}
			_, err := h.Velero().VeleroV1().Restores(h.CurrentProject()).Create(ctx, &restore, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().Restores(h.CurrentProject()).Delete(ctx, "restore-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRdeleteBackupRequests(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit DeleteBackupRequests", func(ctx context.Context) {
			deleteBackupRequest := velerov1.DeleteBackupRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "delete-backup-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().DeleteBackupRequests(h.CurrentProject()).Create(ctx, &deleteBackupRequest, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().DeleteBackupRequests(h.CurrentProject()).Delete(ctx, "delete-backup-request-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRbackupStorageLocations(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit BackupStorageLocations", func(ctx context.Context) {
			backupStorageLocation := velerov1.BackupStorageLocation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "backup-storage-location-test",
				},
				Spec: velerov1.BackupStorageLocationSpec{
					Provider: "aws",
					StorageType: velerov1.StorageType{
						ObjectStorage: &velerov1.ObjectStorageLocation{
							Bucket: "bucket",
							Prefix: "prefix",
						},
					},
				},
			}
			_, err := h.Velero().VeleroV1().BackupStorageLocations(h.CurrentProject()).Create(ctx, &backupStorageLocation, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().BackupStorageLocations(h.CurrentProject()).Delete(ctx, "backup-storage-location-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRdownloadRequests(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit DownloadRequests", func(ctx context.Context) {
			downloadRequest := velerov1.DownloadRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "download-request-test",
				},
				Spec: velerov1.DownloadRequestSpec{
					Target: velerov1.DownloadTarget{
						Kind: velerov1.DownloadTargetKindBackupContents,
						Name: "targetName",
					},
				},
			}

			_, err := h.Velero().VeleroV1().DownloadRequests(h.CurrentProject()).Create(ctx, &downloadRequest, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().DownloadRequests(h.CurrentProject()).Delete(ctx, "download-request-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRpodVolumeBackup(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit PodVolumeBackups", func(ctx context.Context) {
			podVolumeBackup := velerov1.PodVolumeBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-volume-backup-test",
				},
			}
			_, err := h.Velero().VeleroV1().PodVolumeBackups(h.CurrentProject()).Create(ctx, &podVolumeBackup, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().PodVolumeBackups(h.CurrentProject()).Delete(ctx, "pod-volume-backup-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRpodVolumeRestores(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit PodVolumeRestores", func(ctx context.Context) {
			podVolumeRestore := velerov1.PodVolumeRestore{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-volume-restore-test",
				},
			}
			_, err := h.Velero().VeleroV1().PodVolumeRestores(h.CurrentProject()).Create(ctx, &podVolumeRestore, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().PodVolumeRestores(h.CurrentProject()).Delete(ctx, "pod-volume-restore-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRvolumeSnapshotLocation(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit VolumeSnapshotLocations", func(ctx context.Context) {
			volumeSnapshotLocation := velerov1.VolumeSnapshotLocation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "volume-snapshot-locations-test",
				},
			}
			_, err := h.Velero().VeleroV1().VolumeSnapshotLocations(h.CurrentProject()).Create(ctx, &volumeSnapshotLocation, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().VolumeSnapshotLocations(h.CurrentProject()).Delete(ctx, "volume-snapshot-locations-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRschedules(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit Schedules", func(ctx context.Context) {
			schedules := velerov1.Schedule{
				ObjectMeta: metav1.ObjectMeta{
					Name: "schedules-test",
				},
			}
			_, err := h.Velero().VeleroV1().Schedules(h.CurrentProject()).Create(ctx, &schedules, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().Schedules(h.CurrentProject()).Delete(ctx, "schedules-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRserverStatusRequest(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit ServerStatusRequests", func(ctx context.Context) {
			serverStatusRequest := velerov1.ServerStatusRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "server-status-request-test",
				},
			}
			_, err := h.Velero().VeleroV1().ServerStatusRequests(h.CurrentProject()).Create(ctx, &serverStatusRequest, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().ServerStatusRequests(h.CurrentProject()).Delete(ctx, "server-status-request-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func testCRrestricRepository(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("Access should be allowed to edit RestricRepository", func(ctx context.Context) {
			resticRepository := velerov1.ResticRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "restric-repository-test",
				},
			}
			_, err := h.Velero().VeleroV1().ResticRepositories(h.CurrentProject()).Create(ctx, &resticRepository, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			h.Velero().VeleroV1().ResticRepositories(h.CurrentProject()).Delete(ctx, "restric-repository-test", metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
}

func checkVeleroBackups(h *helper.H) {
	ginkgo.Context("velero", func() {
		util.GinkgoIt("backups should be complete", func(ctx context.Context) {
			err := wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
				us, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group:    "velero.io",
					Version:  "v1",
					Resource: "backups",
				}).Namespace(operatorNamespace).List(ctx, metav1.ListOptions{})
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
