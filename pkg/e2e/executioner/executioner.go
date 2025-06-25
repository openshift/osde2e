package executioner

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	projectv1 "github.com/openshift/api/project/v1"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	"github.com/openshift/osde2e/pkg/common/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

type Config struct {
	Image               string
	Environment         ocm.Environment
	ClusterID           string
	CloudProviderID     string
	CloudProviderRegion string
	PassthruSecrets     map[string]string
	Timeout             time.Duration
	OutputDir           string
}

type executioner struct {
	oc     *openshift.Client
	cfg    *Config
	logger logr.Logger
}

// New sets up a new executioner to run a given test suite image
func New(logger logr.Logger, cfg *Config) (*executioner, error) {
	oc, err := openshift.New(logger)
	if err != nil {
		return nil, fmt.Errorf("openshift client creation: %w", err)
	}
	return &executioner{oc: oc, cfg: cfg, logger: logger.WithName("executioner")}, nil
}

func (e *executioner) Execute(ctx context.Context) error {
	// TODO: why does GenerateName not work?
	project := &projectv1.Project{ObjectMeta: metav1.ObjectMeta{Name: "osde2e-executioner-" + util.RandomStr(5)}}
	if err := e.oc.Create(ctx, project); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}
	e.logger.Info("created namespace", "name", project.Name)

	defer func() {
		if err := e.oc.Delete(ctx, project); err != nil {
			panic(err)
		}
	}()

	sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "cluster-admin", Namespace: project.Name}}
	if err := e.oc.Create(ctx, sa); err != nil {
		return fmt.Errorf("creating cluster-admin serviceaccount: %w", err)
	}
	e.logger.Info("created service account", "name", sa.Name)

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "osde2e-executioner-cluster-admin-",
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(project.DeepCopy(), schema.FromAPIVersionAndKind("project.openshift.io/v1", "Project")),
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}

	if err := e.oc.Create(ctx, crb); err != nil {
		return fmt.Errorf("creating cluster-admin clusterrolebinding: %w", err)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "executioner-",
			Namespace:    project.Name,
		},
		Spec: batchv1.JobSpec{
			Parallelism:           ptr.To[int32](1),
			Completions:           ptr.To[int32](1),
			BackoffLimit:          ptr.To[int32](0),
			ActiveDeadlineSeconds: ptr.To(int64(e.cfg.Timeout.Seconds())),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: sa.Name,
					Containers: []corev1.Container{
						{
							Name:            "e2e-suite",
							Image:           e.cfg.Image,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "OCM_CLUSTER_ID",
									Value: e.cfg.ClusterID,
								},
								{
									Name:  "OCM_ENV",
									Value: string(e.cfg.Environment),
								},
								{
									Name:  "CLOUD_PROVIDER_ID",
									Value: e.cfg.CloudProviderID,
								},
								{
									Name:  "CLOUD_PROVIDER_REGION",
									Value: e.cfg.CloudProviderRegion,
								},
								{
									Name:  "GINKGO_NO_COLOR",
									Value: "true",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "results",
									MountPath: "/test-run-results",
								},
							},
						},
						{
							Name:    "pause-for-artifacts",
							Image:   "busybox:latest",
							Command: []string{"tail", "-f", "/dev/null"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "results",
									MountPath: "/test-run-results",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "results",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	if len(e.cfg.PassthruSecrets) > 0 {
		passthruSercret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ci-secrets",
				Namespace: project.Name,
			},
			StringData: e.cfg.PassthruSecrets,
		}

		if err := e.oc.Create(ctx, passthruSercret); err != nil {
			return fmt.Errorf("creating passthru secrets: %w", err)
		}

		job.Spec.Template.Spec.Containers[0].EnvFrom = append(job.Spec.Template.Spec.Containers[0].EnvFrom,
			corev1.EnvFromSource{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "ci-secrets"},
				},
			})
	}

	if err := e.oc.Create(ctx, job); err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	e.logger.Info("created job", "name", job.Name)

	// Wait for the e2e-suite container to complete (succeed/fail/stop)
	// We can't wait for the job because the pause container keeps it running for artifact collection
	if err := wait.PollUntilContextTimeout(ctx, 10*time.Second, e.cfg.Timeout, false, func(ctx context.Context) (bool, error) {
		// Get the pod created by the job
		pods := new(corev1.PodList)
		if err := e.oc.WithNamespace(project.Name).List(ctx, pods, resources.WithLabelSelector(labels.FormatLabels(map[string]string{"job-name": job.Name}))); err != nil {
			return false, nil // retry on error
		}
		if len(pods.Items) == 0 {
			return false, nil // pod not created yet
		}
		pod := &pods.Items[0]
		// Check the status of the e2e-suite container specifically
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name == "e2e-suite" {
				// Return true if container has terminated (succeeded or failed)
				if containerStatus.State.Terminated != nil {
					e.logger.Info("e2e-suite has terminated", "state", containerStatus.State)
					return true, nil
				}
				// Return false if container is still running or waiting
				return false, nil
			}
		}
		// Container status not found yet, keep waiting
		return false, nil
	}); err != nil {
		return fmt.Errorf("waiting for e2e-suite container to complete: %w", err)
	}

	e.logger.Info("fetching artifacts")
	if err := e.fetchArtifacts(ctx, job.Name, job.Namespace); err != nil {
		return fmt.Errorf("fetching artifacts: %w", err)
	}

	return nil
}

func (e *executioner) fetchArtifacts(ctx context.Context, name, namespace string) error {
	// TODO: this is broken, for some reason the api server rejects these requests
	/*
		logs, err := e.oc.GetJobLogs(ctx, name, namespace)
		if err != nil {
			return fmt.Errorf("getting pod logs: %w", err)
		}

		if err = os.WriteFile(filepath.Join(e.cfg.OutputDir, name+".log"), []byte(logs), os.ModePerm); err != nil {
			return fmt.Errorf("writing pod logs: %w", err)
		}
	*/

	pods := new(corev1.PodList)
	if err := e.oc.WithNamespace(namespace).List(ctx, pods, resources.WithLabelSelector(labels.FormatLabels(map[string]string{"job-name": name}))); err != nil {
		return fmt.Errorf("listing pods for job: %w", err)
	}

	if len(pods.Items) == 0 {
		return errors.New("pod for job not found")
	}

	pod := &pods.Items[0]

	clientSet, err := kubernetes.NewForConfig(e.oc.GetConfig())
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}

	execRequest := clientSet.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   []string{"tar", "cf", "-", "-C", "/test-run-results", "/test-run-results"},
			Container: "pause-for-artifacts",
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(e.oc.GetConfig(), "POST", execRequest.URL())
	if err != nil {
		return fmt.Errorf("new remote SPDY executor: %w", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		return fmt.Errorf("streaming executor: %w", err)
	}

	if err = untarBuffer(&stdout, e.cfg.OutputDir); err != nil {
		return fmt.Errorf("untarring buffer: %w", err)
	}

	return nil
}

func untarBuffer(r io.Reader, outputDir string) error {
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // end of stream
		}
		if err != nil {
			return fmt.Errorf("reading tar header: %w", err)
		}
		outputPath := filepath.Join(outputDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(outputPath, os.ModePerm); err != nil {
				return fmt.Errorf("unable to mkdir: %w", err)
			}
		case tar.TypeReg:
			outputFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}
			if _, err = io.Copy(outputFile, tarReader); err != nil {
				_ = outputFile.Close()
				return fmt.Errorf("writing file: %w", err)
			}
			_ = outputFile.Close()
		default:
		}
	}
	return nil
}
