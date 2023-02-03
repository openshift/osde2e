package verify

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/expect"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

var storageTestName string = "[Suite: e2e] Storage"

func init() {
	alert.RegisterGinkgoAlert(storageTestName, "SD-SREP", "Christoph Blecker", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(storageTestName, ginkgo.Ordered, label.HyperShift, label.E2E, func() {
	var h *helper.H
	var client *resources.Resources
	ginkgo.BeforeAll(func() {
		h = helper.New()
		client = h.AsUser("")
	})

	ginkgo.It("PVCs can be managed", func(ctx context.Context) {
		var storageClassList storagev1.StorageClassList
		expect.NoError(client.List(ctx, &storageClassList))

		for _, sc := range storageClassList.Items {
			Expect(sc.AllowVolumeExpansion).To(Equal(pointer.Bool(true)))
		}

		for _, storageClass := range storageClassList.Items {
			pvc := &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "osde2e-",
					Namespace:    h.CurrentProject(),
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: resource.MustParse("2Gi"),
						},
					},
					StorageClassName: pointer.String(storageClass.GetName()),
				},
			}
			expect.NoError(client.Create(ctx, pvc), "failed to create PVC")

			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "osde2e-pvc-tester-",
					Namespace:    h.CurrentProject(),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "test",
							Image:   "registry.access.redhat.com/ubi8/ubi-minimal",
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{"echo 'hello' > /mnt/volume/hello.txt && sleep 1 && sync && grep hello /mnt/volume/hello.txt && sleep 10"},
							Stdin:   true,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "volume",
									MountPath: "/mnt/volume",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "volume",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvc.GetName(),
									ReadOnly:  false,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			}
			expect.NoError(client.Create(ctx, pod), "failed to create pod")

			err := wait.For(conditions.New(client).PodPhaseMatch(pod, v1.PodSucceeded), wait.WithTimeout(2*time.Minute))
			expect.NoError(err, "pod %q never succeeded", pod.GetName())

			expect.NoError(client.Delete(ctx, pod), "unable to delete pod")
			expect.NoError(client.Delete(ctx, pvc), "unable to delete PVC")
		}
	})
})
