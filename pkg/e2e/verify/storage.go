package verify

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var storageTestName string = "[Suite: e2e] Storage"

func init() {
	alert.RegisterGinkgoAlert(storageTestName, "SD-SREP", "Christoph Blecker", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

const (

	// podStartTimeout is how long to wait for the pod to be started.
	podStartTimeout = 3 * time.Minute

	// poll is how often to poll pods, nodes and claims.
	poll = 1 * time.Second
)

var _ = ginkgo.Describe(storageTestName, func() {
	h := helper.New()

	ginkgo.Context("storage", func() {
		util.GinkgoIt("create PVCs", func(ctx context.Context) {
			sc, err := h.Kube().StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred(), "couldn't list StorageClasses")
			list, provider := getScNames(sc, h)
			log.Printf(" PROVIDER:  %v", provider)
			var pvc PersistentVolumeClaimConfig
			namespace := "test-namespace"
			_, err = createNamespace(ctx, namespace, h)
			Expect(err).NotTo(HaveOccurred(), "couldn't create Namespace")
			defer func() {
				deleteNamespace(ctx, namespace, true, h)
			}()
			for _, name := range list {
				var vc []*corev1.PersistentVolumeClaim
				storageClassName := name
				pvcName := "pvc-osde2e-" + name
				pvcYaml := makePersistentVolumeClaim(pvc, namespace, storageClassName, pvcName)
				pvc, err := createPVC(ctx, h, namespace, pvcYaml)
				log.Printf(" CREATING pvc:  %s", name)
				defer func() {
					deletePersistentVolumeClaim(ctx, h, pvcName, namespace)
				}()
				Expect(err).NotTo(HaveOccurred(), "couldn't create PVC")
				vc = append(vc, pvc)
				pod, err := createTestPod(ctx, h, namespace, vc, false)
				Expect(err).NotTo(HaveOccurred())
				podName := pod.GetName()
				defer func() {
					deletePod(ctx, podName, namespace, h)
				}()

			}
		}, podStartTimeout.Seconds()+viper.GetFloat64(config.Tests.PollingTimeout))
	})

	ginkgo.Context("sc-list", func() {
		util.GinkgoIt("should be able to be expanded", func(ctx context.Context) {
			scList, err := h.Kube().StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred(), "couldn't list StorageClasses")
			Expect(scList).NotTo(BeNil())

			for _, sc := range scList.Items {
				Expect(sc.AllowVolumeExpansion).To(Not(BeNil()))
				Expect(*sc.AllowVolumeExpansion).To(BeTrue())
			}
		}, 300)
	})
})

// Get Storage Class names and cloud provider
func getScNames(list *v1.StorageClassList, h *helper.H) ([]string, string) {
	var scs []string
	var provider string
	var provisioner string
	provisioner = ""
	for _, sc := range list.Items {
		name := sc.ObjectMeta.Name
		if provisioner == "" {
			provisioner = sc.Provisioner
			if strings.Contains(provisioner, "aws") {
				provider = "aws"
			} else {
				provider = "gcp"
			}
		}
		scs = append(scs, name)
		log.Printf("\n PROVIDER:  %v \n", provider)
	}
	return scs, provider
}

// PersistentVolumeClaimConfig is consumed by makePersistentVolumeClaim() to
// generate a PVC object.
type PersistentVolumeClaimConfig struct {
	// Name of the PVC. If set, overrides NamePrefix
	Name string
	// NamePrefix defaults to "pvc-" if unspecified
	NamePrefix string
	// ClaimSize must be specified in the Quantity format. Defaults to 2Gi if
	// unspecified
	ClaimSize string
	// AccessModes defaults to RWO if unspecified
	AccessModes      []corev1.PersistentVolumeAccessMode
	Annotations      map[string]string
	Selector         *metav1.LabelSelector
	StorageClassName *string
	// VolumeMode defaults to nil if unspecified or specified as the empty
	// string
	VolumeMode *corev1.PersistentVolumeMode
}

// makePersistentVolumeClaim returns a PVC API Object based on the PersistentVolumeClaimConfig.
func makePersistentVolumeClaim(cfg PersistentVolumeClaimConfig, ns string, storageClass string, name string) *corev1.PersistentVolumeClaim {
	cfg.Name = name

	cfg.StorageClassName = &storageClass

	if len(cfg.AccessModes) == 0 {
		cfg.AccessModes = append(cfg.AccessModes, corev1.ReadWriteOnce)
	}

	if len(cfg.ClaimSize) == 0 {
		cfg.ClaimSize = "2Gi"
	}

	if len(cfg.NamePrefix) == 0 {
		cfg.NamePrefix = "pvc-"
	}

	if cfg.VolumeMode != nil && *cfg.VolumeMode == "" {
		log.Printf("Warning: Making PVC: VolumeMode specified as invalid empty string, treating as nil")
		cfg.VolumeMode = nil
	}

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:         cfg.Name,
			GenerateName: cfg.NamePrefix,
			Namespace:    ns,
			Annotations:  cfg.Annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Selector:    cfg.Selector,
			AccessModes: cfg.AccessModes,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(cfg.ClaimSize),
				},
			},
			StorageClassName: cfg.StorageClassName,
			VolumeMode:       cfg.VolumeMode,
		},
	}
}

// createPVC creates the PVC resource. Fails test on error.
func createPVC(ctx context.Context, h *helper.H, ns string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := h.Kube().CoreV1().PersistentVolumeClaims(ns).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("PVC Create API error: %v", err)
	}
	return pvc, nil
}

// deletePersistentVolumeClaim deletes the PVC.
func deletePersistentVolumeClaim(ctx context.Context, h *helper.H, pvcName string, ns string) error {
	if h != nil && len(pvcName) > 0 {
		log.Printf("Deleting PersistentVolumeClaim %q", pvcName)
		err := h.Kube().CoreV1().PersistentVolumeClaims(ns).Delete(ctx, pvcName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("PVC Delete API error: %v", err)
		}
	}
	return nil
}

// Config is a struct containing all arguments for creating a pod.
// SELinux testing requires to pass HostIPC and HostPID as boolean arguments.
type Config struct {
	NS                     string
	PVCs                   []*corev1.PersistentVolumeClaim
	PVCsReadOnly           bool
	InlineVolumeSources    []*corev1.VolumeSource
	IsPrivileged           bool
	Command                string
	HostIPC                bool
	HostPID                bool
	SeLinuxLabel           *corev1.SELinuxOptions
	FsGroup                *int64
	ImageID                int
	PodFSGroupChangePolicy *corev1.PodFSGroupChangePolicy
}

// makeTestPod returns a pod definition based on the namespace. The pod references the PVC's
// name.
func makeTestPod(ns string, pvclaims []*corev1.PersistentVolumeClaim, isPrivileged bool) *corev1.Pod {
	podSpec := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-tester-",
			Namespace:    ns,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "dummy",
					Image:   "registry.access.redhat.com/ubi8/ubi-minimal",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", "echo 'hello' > /mnt/volume1/hello.txt && sleep 1 && sync && grep hello /mnt/volume1/hello.txt; sleep 3600"},
					Stdin:   true,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &isPrivileged,
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
	volumeMounts := make([]corev1.VolumeMount, len(pvclaims))
	volumes := make([]corev1.Volume, len(pvclaims))
	for index, pvclaim := range pvclaims {
		volumename := fmt.Sprintf("volume%v", index+1)
		volumeMounts[index] = corev1.VolumeMount{Name: volumename, MountPath: "/mnt/" + volumename}
		volumes[index] = corev1.Volume{Name: volumename, VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: pvclaim.Name, ReadOnly: false}}}
	}
	podSpec.Spec.Containers[0].VolumeMounts = volumeMounts
	podSpec.Spec.Volumes = volumes
	return podSpec
}

// createTestPod creates pod with given pvc claims
func createTestPod(ctx context.Context, h *helper.H, namespace string, pvclaims []*corev1.PersistentVolumeClaim, isPrivileged bool) (*corev1.Pod, error) {
	pod := makeTestPod(namespace, pvclaims, isPrivileged)
	pod, err := h.Kube().CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("pod Create API error: %v", err)
	}
	// Waiting for pod to be running
	err = waitForPodNameRunningInNamespace(ctx, h, pod.Name, namespace)
	if err != nil {
		return pod, fmt.Errorf("pod %q is not Running: %v", pod.Name, err)
	}
	// get fresh pod info
	pod, err = h.Kube().CoreV1().Pods(namespace).Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		return pod, fmt.Errorf("pod Get API error: %v", err)
	}
	return pod, nil
}

// errPodCompleted is returned by PodRunning to indicate that
// the pod has already reached completed state.
var errPodCompleted = fmt.Errorf("pod ran to completion")

// podRunning checks if pod is running
func podRunning(ctx context.Context, h *helper.H, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := h.Kube().CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, errPodCompleted
		}
		return false, nil
	}
}

// waitForPodNameRunningInNamespace waits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func waitForPodNameRunningInNamespace(ctx context.Context, h *helper.H, podName, namespace string) error {
	return waitTimeoutForPodRunningInNamespace(ctx, h, podName, namespace, podStartTimeout)
}

// waitTimeoutForPodRunningInNamespace waits the given timeout duration for the specified pod to become running.
func waitTimeoutForPodRunningInNamespace(ctx context.Context, h *helper.H, podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(poll, timeout, podRunning(ctx, h, podName, namespace))
}
