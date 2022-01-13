package verify

import (
	"context"
	"fmt"
	"log"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	namespaceListsTestName        = "[Suite: e2e] Namespace Lists"
	namespaceListTestPollInterval = 30 * time.Second
	namespaceListTestPollTimeout  = 20 * time.Minute
)

func init() {
	alert.RegisterGinkgoAlert(namespaceListsTestName, "SD-SREP", "Trevor Nierman", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(namespaceListsTestName, func() {
	ginkgo.Context("are accurate", func() {
		h := helper.New()

		ginkgo.It("for all managed-object configMaps in the cluster", func() {
			log.Printf("Generating & retrieving configMaps")
			testName := fmt.Sprintf("test-namespace-lists-%s-%d-%d", time.Now().Format("20060102-150405"), time.Now().Nanosecond()/1000000, ginkgo.GinkgoParallelProcess())
			currentConfigMaps, newConfigMaps, err := generateNamespaceConfigMaps(h, testName)
			Expect(err).ToNot(HaveOccurred(), "Error while generating and retrieving configMaps")

			log.Printf("Comparing configMaps")
			Expect(len(currentConfigMaps)).To(Equal(len(newConfigMaps)), fmt.Sprintf("Expected to generate same configmaps as configured in managed-cluster-config.\nNumber of generated ConfigMaps: %d\nNumber of configMaps in managed-cluster-config: %d", len(newConfigMaps), len(currentConfigMaps)))

			for name, newList := range newConfigMaps {
				currentList, present := currentConfigMaps[name]
				Expect(present).To(BeTrue(), "Expected ConfigMap %s to be present in managed-cluster-config.")
				Expect(newList).To(Equal(currentList), fmt.Sprintf("ConfigMap '%s' is out of date and needs to be regenerated in managed-cluster-config (https://github.com/openshift/managed-cluster-config/tree/master/deploy/osd-managed-resources#readme) and managed-cluster-validating-webhooks (https://github.com/openshift/managed-cluster-validating-webhooks#updating-namespace-and-service-account-list and https://github.com/openshift/managed-cluster-validating-webhooks#updating-documenation-files).\n\nmanaged-cluster-config contains:\n%s\nShould be:\n%s", name, currentList, newList))
			}

		}, float64(namespaceListTestPollTimeout.Seconds()*2))
	})
})

// generateNamespaceConfigMaps returns the *-namespaces configMaps that are currently present in managed-cluster-config, as well as new versions of these configmaps generated against the test cluster.
func generateNamespaceConfigMaps(h *helper.H, testName string) (currentConfigMaps map[string]string, newConfigMaps map[string]string, err error) {
	log.Printf("Creating job")

	// Commands used in the following job to generate the namespace configMaps.
	// Returns a set of original and newly generated configmaps inside of parent configmaps named after the job
	const (
		// Clone managed-cluster-config && save a copy of the current configMaps
		cloneInitCmd string = `cd /work-dir &&\ 
git clone https://github.com/openshift/managed-cluster-config.git --branch master --single-branch && \
mkdir /work-dir/original/ && \
mv /work-dir/managed-cluster-config/deploy/osd-managed-resources/*.yaml /work-dir/original/`

		// Add oc binary to the working dir (/work-dir)
		addBinaryInitCmd string = `mkdir /work-dir/bin && \
cp /usr/bin/oc /work-dir/bin`

		// Setup python env, run generation script, & expose results as configMaps
		runCmd string = `python3 -m venv /venv-configmap-generation && \
source /venv-configmap-generation/bin/activate && \
pip3 install oyaml && \
cd /work-dir/managed-cluster-config/scripts/managed-resources && \
PATH=${PATH}:/work-dir/bin ./make-all-managed-lists.sh && \
cd /work-dir/original/ && \
/work-dir/bin/oc create configmap "${TEST_NAME}-original" --from-file addons-namespaces.ConfigMap.yaml --from-file managed-namespaces.ConfigMap.yaml --from-file ocp-namespaces.ConfigMap.yaml && \
cd /work-dir/managed-cluster-config/deploy/osd-managed-resources && \
/work-dir/bin/oc create configmap "${TEST_NAME}-new" --from-file addons-namespaces.ConfigMap.yaml --from-file managed-namespaces.ConfigMap.yaml --from-file ocp-namespaces.ConfigMap.yaml`
	)
	completions := int32(1)

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testName,
			Namespace: h.CurrentProject(),
		},
		Spec: batchv1.JobSpec{
			Completions: &completions,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:    "clone",
							Image:   "quay.io/bitnami/git:latest",
							Command: []string{"/bin/sh"},
							Args:    []string{"-c", cloneInitCmd},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "empty-dir",
									MountPath: "/work-dir",
								},
							},
						},
						{
							Name: "add-binary",
							Image:   "quay.io/openshift/origin-cli:latest",
							Command: []string{"/bin/sh"},
							Args:    []string{"-c", addBinaryInitCmd},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "empty-dir",
									MountPath: "/work-dir",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: "generate",
							// Fedora image is necessary at this time, as generation script requires python >3.7
							Image:   "quay.io/fedora/fedora:latest",
							Command: []string{"/bin/sh"},
							Args:    []string{"-c", runCmd},
							Env: []corev1.EnvVar{
								{
									Name:  "TEST_NAME",
									Value: testName,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "empty-dir",
									MountPath: "/work-dir",
								},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: "cluster-admin",
					Volumes: []corev1.Volume{
						{
							Name: "empty-dir",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	job, err = h.Kube().BatchV1().Jobs(h.CurrentProject()).Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return
	}

	log.Printf("Waiting for job to complete")
	err = wait.PollImmediate(namespaceListTestPollInterval, namespaceListTestPollTimeout, func() (bool, error) {
		job, err := h.Kube().BatchV1().Jobs(h.CurrentProject()).Get(context.TODO(), testName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if job.Status.Succeeded == *job.Spec.Completions {
			return true, nil
		}
		return false, nil

	})
	if err != nil {
		return
	}

	log.Printf("Cleaning up job")
	err = deleteJob(job.Name, h.CurrentProject(), h)
	if err != nil {
		return
	}

	log.Printf("Retrieving configMaps")
	originalName := fmt.Sprintf("%s-original", testName)
	currentConfigMaps, err = getConfigMapData(originalName, h.CurrentProject(), h)
	if err != nil {
		return
	}

	newName := fmt.Sprintf("%s-new", testName)
	newConfigMaps, err = getConfigMapData(newName, h.CurrentProject(), h)
	if err != nil {
		return
	}

	log.Printf("Cleaning up configMaps")
	err = deleteConfigMap(originalName, h.CurrentProject(), h)
	if err != nil {
		return
	}

	err = deleteConfigMap(newName, h.CurrentProject(), h)
	if err != nil {
		return
	}

	return
}

// getConfigMapData returns the data field of a given configMap
func getConfigMapData(name string, namespace string, h *helper.H) (map[string]string, error) {
	configMap, err := h.Kube().CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return map[string]string{}, err
	}

	return configMap.Data, nil
}

// deleteConfigMap deletes a given configMap
func deleteConfigMap(name string, namespace string, h *helper.H) error {
	_, err := h.Kube().CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	log.Printf("Check before deleting configMap %s, error: %v", name, err)
	if err == nil {
		err = h.Kube().CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		log.Printf("Deleting configMap %s, error: %v", name, err)

		// Wait for the configMap to delete.
		err = wait.PollImmediate(namespaceListTestPollInterval, namespaceListTestPollTimeout, func() (bool, error) {
			_, err := h.Kube().CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				return true, nil
			}
			return false, nil
		})

	}
	return err
}

// deleteJob deletes a given job
func deleteJob(name string, namespace string, h *helper.H) error {
	_, err := h.Kube().BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	log.Printf("Check before deleting job %s, error: %v", name, err)
	if err == nil {
		err = h.Kube().BatchV1().Jobs(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		log.Printf("Deleting job %s, error: %v", name, err)

		// Wait for the job to delete.
		err = wait.PollImmediate(namespaceListTestPollInterval, namespaceListTestPollTimeout, func() (bool, error) {
			_, err := h.Kube().BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				return true, nil
			}
			return false, nil
		})

	}
	return err
}
