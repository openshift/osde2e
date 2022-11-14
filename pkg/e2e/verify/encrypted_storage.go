package verify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cloudcredentialv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/util"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	kmsv1 "cloud.google.com/go/kms/apiv1"
	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	computev1 "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	serviceusagev1 "google.golang.org/api/serviceusage/v1"
	kmsprotov1 "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

const (
	encryptedStorageTestName     string = "[Suite: e2e] Encrypted Storage"
	encryptedStoragePollInterval        = 30 * time.Second
	encryptedStoragePollTimeout         = 10 * time.Minute
)

func init() {
	alert.RegisterGinkgoAlert(encryptedStorageTestName, "SD-SREP", "Trevor Nierman", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(encryptedStorageTestName, func() {
	ginkgo.Context("in GCP clusters", func() {
		if viper.GetString(config.CloudProvider.CloudProviderID) != "gcp" {
			return
		}
		h := helper.New()
		var testInstanceName string

		util.GinkgoIt("can be created by dedicated admins", func(ctx context.Context) {
			testInstanceName = "test-" + time.Now().Format("20060102-150405-") + fmt.Sprint(time.Now().Nanosecond()/1000000) + "-" + fmt.Sprint(ginkgo.GinkgoParallelProcess())

			ginkgo.By("Creating an encryption key in the cluster's gcp project")
			serviceAccountJson, err := createGCPServiceAccount(ctx, h, testInstanceName, h.CurrentProject())
			Expect(err).ToNot(HaveOccurred(), "Error creating credentialsrequest")

			testKey, err := createGCPKey(ctx, h, serviceAccountJson, testInstanceName)
			Expect(err).ToNot(HaveOccurred(), "Error creating or retrieving encryption key in GCP KMS.")

			ginkgo.By("Using the key to create an encrypted storageclass as a dedicated-admin")
			// RBAC rules must be set manually if not testing with a ccs cluster
			// (gcp dedicated-admins only have permissions to create/modify storageclasses on ccs clusters)
			if !viper.GetBool(ocmprovider.CCS) {
				_, err := h.Kube().RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterRole",
						APIVersion: "rbac.authorization.k8s.io/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: testInstanceName,
						Labels: map[string]string{
							"managed.openshift.io/aggregate-to-dedicated-admins": "cluster",
						},
					},
					Rules: []rbacv1.PolicyRule{{
						Verbs:     []string{"*"},
						APIGroups: []string{"storage.k8s.io"},
						Resources: []string{"storageclasses"},
					}},
				}, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred(), "Error creating clusterrole '"+testInstanceName+"':")
			}
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})

			// Updated RBAC rules may take time to apply -> poll until we can successfully create a storageclass
			volumeBindingMode := storagev1.VolumeBindingWaitForFirstConsumer
			wait.PollImmediate(encryptedStoragePollInterval, encryptedStoragePollTimeout, func() (bool, error) {
				_, err = h.Kube().StorageV1().StorageClasses().Create(ctx, &storagev1.StorageClass{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StorageClass",
						APIVersion: "storage.k8s.io/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: testInstanceName,
					},
					Provisioner: "pd.csi.storage.gke.io",
					Parameters: map[string]string{
						"type":                    "pd-standard",
						"disk-encryption-kms-key": testKey,
					},
					VolumeBindingMode: &volumeBindingMode,
				}, metav1.CreateOptions{})
				if err != nil {
					return false, nil
				}
				return true, err
			})
			Expect(err).ToNot(HaveOccurred(), "Error creating storageclass '"+testInstanceName+"':")

			ginkgo.By("Verifying the storageclass creates encrypted PVCs")
			// Ensure pods can read/write on encrypted PVs
			pvc, err := h.Kube().CoreV1().PersistentVolumeClaims(h.CurrentProject()).Create(ctx, &corev1.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolumeClaim",
					APIVersion: "core/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testInstanceName,
					Namespace: h.CurrentProject(),
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: &testInstanceName,
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("256Mi"),
						},
					},
				},
			}, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred(), "Error creating pvc '"+pvc.GetName()+"':")

			volumeMountPath := "/mnt/volume"
			pod, err := h.Kube().CoreV1().Pods(h.CurrentProject()).Create(ctx, &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testInstanceName,
					Namespace: h.CurrentProject(),
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{{
						Name:    testInstanceName,
						Image:   "registry.access.redhat.com/ubi8/ubi-minimal",
						Command: []string{"/bin/sh"},
						Args:    []string{"-c", "echo 'Hello world!' > " + volumeMountPath + "/hello-world.txt && sleep 1 && sync && cat " + volumeMountPath + "/hello-world.txt"},
						Stdin:   true,
						VolumeMounts: []corev1.VolumeMount{{
							Name:      pvc.GetName(),
							MountPath: volumeMountPath,
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: pvc.GetName(),
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvc.GetName(),
								ReadOnly:  false,
							},
						},
					}},
				},
			}, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred(), "Error creating pod '"+pod.GetName()+"':")

			wait.PollImmediate(encryptedStoragePollInterval, encryptedStoragePollTimeout, func() (bool, error) {
				pod, err = h.Kube().CoreV1().Pods(pod.GetNamespace()).Get(ctx, pod.GetName(), metav1.GetOptions{})
				if err != nil || pod.Status.Phase != corev1.PodSucceeded {
					return false, err
				}
				return true, err
			})
			Expect(err).ToNot(HaveOccurred(), "Error or timeout waiting for pod '"+pod.GetName()+"' to succeed. This implies the pod had trouble reading/writing to encrypted disk.")

			// Ensure that gcp identifies the disk as customer-managed encryption
			node, err := h.Kube().CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred(), "Error retrieving node '"+pod.Spec.NodeName+"':")

			computeService, err := computev1.NewService(ctx, option.WithCredentialsJSON(serviceAccountJson))
			Expect(err).ToNot(HaveOccurred(), "Error creating computev1 service:")

			clusterInfra, err := h.Cfg().ConfigV1().Infrastructures().Get(ctx, "cluster", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred(), "Error retrieving cluster infrastructure:")

			pvc, err = h.Kube().CoreV1().PersistentVolumeClaims(h.CurrentProject()).Get(ctx, pvc.GetName(), metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred(), "Error retrieving pvc '"+testInstanceName+"':")

			disk, err := computeService.Disks.Get(clusterInfra.Status.PlatformStatus.GCP.ProjectID, node.Labels["topology.kubernetes.io/zone"], pvc.Spec.VolumeName).Context(ctx).Do()
			Expect(err).ToNot(HaveOccurred(), "Error retrieving encrypted disk from gcp: ")
			Expect(strings.Contains(disk.DiskEncryptionKey.KmsKeyName, testKey), "Disk '"+disk.Name+"' not using a customer managed key as expected!")
		}, float64(encryptedStoragePollTimeout*2))

		// Cleanup
		ginkgo.AfterEach(func(ctx context.Context) {
			h.Impersonate(rest.ImpersonationConfig{})

			err := deleteGCPServiceAccount(ctx, h, testInstanceName)
			Expect(err).ToNot(HaveOccurred())

			err = deleteStorageClass(ctx, h, testInstanceName)
			Expect(err).ToNot(HaveOccurred())

			err = deletePersistentVolumeClaim(ctx, h, testInstanceName, h.CurrentProject())
			Expect(err).ToNot(HaveOccurred())

			err = deletePod(ctx, testInstanceName, h.CurrentProject(), h)
			Expect(err).ToNot(HaveOccurred())

			if !viper.GetBool(ocmprovider.CCS) {
				err = deleteClusterRole(ctx, h, testInstanceName)
				Expect(err).ToNot(HaveOccurred())
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})

// Creates a keyring & key using the given service account credentials. Returns the full name of the key in GCP
func createGCPKey(ctx context.Context, h *helper.H, serviceAccountJson []byte, keyName string) (string, error) {
	clusterInfra, err := h.Cfg().ConfigV1().Infrastructures().Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	err = enableGCPKMS(ctx, clusterInfra.Status.PlatformStatus.GCP.ProjectID, serviceAccountJson)
	if err != nil {
		return "", err
	}

	kmsClient, err := kmsv1.NewKeyManagementClient(ctx, option.WithCredentialsJSON(serviceAccountJson))
	if err != nil {
		return "", err
	}
	defer kmsClient.Close()

	// GCP KMS doesn't allow cryptokeys to be deleted. Therefore we have to reuse any key or keyRing that already exists with the same name
	keyRing, err := kmsClient.GetKeyRing(ctx, &kmsprotov1.GetKeyRingRequest{
		Name: "projects/" + clusterInfra.Status.PlatformStatus.GCP.ProjectID + "/locations/" + clusterInfra.Status.PlatformStatus.GCP.Region + "/keyRings/" + keyName,
	})
	if err != nil {
		// keyRing does not exist yet, & KMS was likely just enabled for the project.
		// KMS may not be ready immediately after enabling, poll until we are able to successfully create a keyring, indicating it is fully available for this project
		wait.PollImmediate(encryptedStoragePollInterval, encryptedStoragePollTimeout, func() (bool, error) {
			keyRing, err = kmsClient.CreateKeyRing(ctx, &kmsprotov1.CreateKeyRingRequest{
				Parent:    "projects/" + clusterInfra.Status.PlatformStatus.GCP.ProjectID + "/locations/" + clusterInfra.Status.PlatformStatus.GCP.Region,
				KeyRingId: keyName,
			})
			if keyRing == nil {
				return false, nil
			} else if err != nil {
				return false, err
			}
			return true, err
		})
		if err != nil {
			return "", err
		}
	}

	key, err := kmsClient.GetCryptoKey(ctx, &kmsprotov1.GetCryptoKeyRequest{
		Name: keyRing.GetName() + "/cryptoKeys/" + keyName,
	})
	if key != nil && err == nil {
		return key.GetName(), err
	}

	key, err = kmsClient.CreateCryptoKey(ctx, &kmsprotov1.CreateCryptoKeyRequest{
		Parent:      keyRing.GetName(),
		CryptoKeyId: keyName,
		CryptoKey: &kmsprotov1.CryptoKey{
			Purpose: kmsprotov1.CryptoKey_ENCRYPT_DECRYPT,
			VersionTemplate: &kmsprotov1.CryptoKeyVersionTemplate{
				Algorithm: kmsprotov1.CryptoKeyVersion_GOOGLE_SYMMETRIC_ENCRYPTION,
			},
		},
	})
	if err != nil {
		return "", err
	}
	return key.GetName(), err
}

// Enables KMS in GCP & grants necessary permissions to utilize it for pvc encryption using the given service account
// NOTE: KMS may take up to several minutes to enable for a project (there doesn't appear to be a way to check if it's enabled except to attempt to use it).
func enableGCPKMS(ctx context.Context, projectID string, serviceAccountJson []byte) error {
	// Enable KMS service
	suService, err := serviceusagev1.NewService(ctx, option.WithCredentialsJSON(serviceAccountJson))
	if err != nil {
		return err
	}
	_, err = suService.Services.BatchEnable("projects/"+projectID, &serviceusagev1.BatchEnableServicesRequest{ServiceIds: []string{"cloudkms.googleapis.com"}}).Do()
	if err != nil {
		return err
	}

	// Add necessary permissions to the 'Compute Engine Service Agent' account
	crmService, err := cloudresourcemanagerv1.NewService(ctx, option.WithCredentialsJSON(serviceAccountJson))
	if err != nil {
		return err
	}
	result, err := crmService.Projects.Get(projectID).Context(ctx).Do()
	if err != nil {
		return err
	}
	binding := &cloudresourcemanagerv1.Binding{
		Members: []string{"serviceAccount:service-" + fmt.Sprint(result.ProjectNumber) + "@compute-system.iam.gserviceaccount.com"},
		Role:    "roles/cloudkms.cryptoKeyEncrypterDecrypter",
	}
	policy, err := crmService.Projects.GetIamPolicy(projectID, &cloudresourcemanagerv1.GetIamPolicyRequest{}).Do()
	if err != nil {
		return err
	}
	policy.Bindings = append(policy.Bindings, binding)
	_, err = crmService.Projects.SetIamPolicy(projectID, &cloudresourcemanagerv1.SetIamPolicyRequest{Policy: policy}).Do()
	return err
}

// Creates a unique service account with owner privileges in GCP for the current test instance
// Returns the service-account.json file for the newly created account
func createGCPServiceAccount(ctx context.Context, h *helper.H, saName string, saNamespace string) ([]byte, error) {
	providerBytes := bytes.Buffer{}
	encoder := json.NewEncoder(&providerBytes)
	encoder.Encode(cloudcredentialv1.GCPProviderSpec{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GCPProviderSpec",
			APIVersion: "cloudcredential.openshift.io/v1",
		},
		PredefinedRoles: []string{
			"roles/owner",
		},
		SkipServiceCheck: true,
	})
	saCredentialReq := &cloudcredentialv1.CredentialsRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CredentialsRequest",
			APIVersion: "cloudcredential.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: saNamespace,
		},
		Spec: cloudcredentialv1.CredentialsRequestSpec{
			SecretRef: corev1.ObjectReference{
				Name:      saName,
				Namespace: saNamespace,
			},
			ProviderSpec: &runtime.RawExtension{
				Raw:    providerBytes.Bytes(),
				Object: &cloudcredentialv1.GCPProviderSpec{},
			},
		},
	}

	credentialReqObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(saCredentialReq)
	if err != nil {
		return nil, err
	}

	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "cloudcredential.openshift.io",
		Version:  "v1",
		Resource: "credentialsrequests",
	}).Namespace(saCredentialReq.GetNamespace()).Create(ctx, &unstructured.Unstructured{Object: credentialReqObj}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	wait.PollImmediate(encryptedStoragePollInterval, encryptedStoragePollTimeout, func() (bool, error) {
		unstructCredentialReq, err := h.Dynamic().Resource(schema.GroupVersionResource{
			Group:    "cloudcredential.openshift.io",
			Version:  "v1",
			Resource: "credentialsrequests",
		}).Namespace(saNamespace).Get(ctx, saCredentialReq.GetName(), metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructCredentialReq.UnstructuredContent(), saCredentialReq)
		if err != nil || !saCredentialReq.Status.Provisioned {
			return false, err
		}
		return true, err
	})
	if err != nil {
		return nil, err
	}

	saSecret, err := h.Kube().CoreV1().Secrets(saCredentialReq.Spec.SecretRef.Namespace).Get(ctx, saCredentialReq.Spec.SecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return saSecret.Data["service_account.json"], err
}

func deleteGCPServiceAccount(ctx context.Context, h *helper.H, name string) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "cloudcredential.openshift.io",
		Version:  "v1",
		Resource: "credentialsrequests",
	}).Namespace(h.CurrentProject()).Delete(ctx, name, metav1.DeleteOptions{})
}

func deleteClusterRole(ctx context.Context, h *helper.H, name string) error {
	return h.Kube().RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
}

func deleteStorageClass(ctx context.Context, h *helper.H, name string) error {
	return h.Kube().StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
}
