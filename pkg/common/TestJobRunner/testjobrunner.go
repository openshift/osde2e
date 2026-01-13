package TestJobRunner

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

type Config struct {
	Environment         ocm.Environment
	ClusterID           string
	CloudProviderID     string
	CloudProviderRegion string
	PassthruSecrets     map[string]string
	Timeout             time.Duration
	OutputDir           string
	SkipCleanup         bool
	RestConfig          *rest.Config
}

type TestJobRunner struct {
	oc     *openshift.Client
	cfg    *Config
	logger logr.Logger
}

// New sets up a new executor to run a given test suite image
func New(logger logr.Logger, cfg *Config) (*TestJobRunner, error) {
	var oc *openshift.Client
	var err error
	if cfg.RestConfig != nil {
		oc, err = openshift.NewFromRestConfig(cfg.RestConfig, logger)
	} else {
		oc, err = openshift.New(logger)
	}
	if err != nil {
		return nil, fmt.Errorf("openshift client creation: %w", err)
	}
	return &TestJobRunner{oc: oc, cfg: cfg, logger: logger.WithName("executor")}, nil
}

func (e *TestJobRunner) Execute(ctx context.Context, image string) (*testResults, error) {
	if err := os.MkdirAll(e.cfg.OutputDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	project, err := e.setupProject(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating namespace: %w", err)
	}

	defer func() {
		if e.cfg.SkipCleanup {
			e.logger.Info("Skipping cleanup")
			return
		}
		if err := e.oc.Delete(ctx, project); err != nil {
			e.logger.Info("Failed to delete project", "name", project.Name)
		}
	}()

	job, err := e.createJob(ctx, project.Name, image)
	if err != nil {
		return nil, fmt.Errorf("creating job: %w", err)
	}

	e.logger.Info("waiting for suite to complete")
	if err := e.waitForSuite(ctx, job.Name, job.Namespace, image); err != nil {
		return nil, fmt.Errorf("waiting for suite to finish: %w", err)
	}

	e.logger.Info("fetching artifacts")
	if err := e.fetchArtifacts(ctx, job.Name, job.Namespace); err != nil {
		return nil, fmt.Errorf("fetching artifacts: %w", err)
	}

	e.logger.Info("processing test results")
	results, err := processJUnitResults(e.logger, e.cfg.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("processing junit results: %w", err)
	}

	return results, nil
}

func (e *TestJobRunner) setupProject(ctx context.Context) (*projectv1.Project, error) {
	// TODO: why does GenerateName not work?
	project := &projectv1.Project{ObjectMeta: metav1.ObjectMeta{Name: "osde2e-executor-" + util.RandomStr(5)}}
	if err := e.oc.Create(ctx, project); err != nil {
		return nil, err
	}
	e.logger.Info("created namespace", "name", project.Name)

	sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "cluster-admin", Namespace: project.Name}}
	if err := e.oc.Create(ctx, sa); err != nil {
		return nil, fmt.Errorf("creating cluster-admin serviceaccount: %w", err)
	}
	e.logger.Info("created service account", "name", sa.Name)

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "osde2e-executor-cluster-admin-",
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
		return nil, fmt.Errorf("creating cluster role binding: %w", err)
	}

	return project, nil
}

func (e *TestJobRunner) createJob(ctx context.Context, namespace string, image string) (*batchv1.Job, error) {
	job := e.buildJobSpec(namespace, image)

	if len(e.cfg.PassthruSecrets) > 0 {
		passthruSercret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ci-secrets",
				Namespace: namespace,
			},
			StringData: e.cfg.PassthruSecrets,
		}

		if err := e.oc.Create(ctx, passthruSercret); err != nil {
			return nil, fmt.Errorf("creating passthru secrets: %w", err)
		}

		job.Spec.Template.Spec.Containers[0].EnvFrom = append(job.Spec.Template.Spec.Containers[0].EnvFrom,
			corev1.EnvFromSource{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "ci-secrets"},
				},
			})
	}

	if err := e.oc.Create(ctx, job); err != nil {
		return nil, err
	}
	e.logger.Info("created job", "name", job.Name)
	return job, nil
}

func (e *TestJobRunner) buildJobSpec(namespace string, image string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "executor-",
			Namespace:    namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism:           ptr.To[int32](1),
			Completions:           ptr.To[int32](1),
			BackoffLimit:          ptr.To[int32](0),
			ActiveDeadlineSeconds: ptr.To(int64(e.cfg.Timeout.Seconds())),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"openshift.io/required-scc": "restricted-v2",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "cluster-admin",
					Containers: []corev1.Container{
						{
							Name:            "e2e-suite",
							Image:           image,
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
									Value: "TRUE",
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
}

// Wait for the e2e-suite container to complete (succeed/fail/stop)
// We can't wait for the job because the pause container keeps it running for artifact collection
func (e *TestJobRunner) waitForSuite(ctx context.Context, name, namespace, image string) error {
	return wait.PollUntilContextTimeout(ctx, 10*time.Second, e.cfg.Timeout, false, func(ctx context.Context) (bool, error) {
		pod, err := e.findJobPod(ctx, name, namespace)
		if err != nil {
			return false, nil // pod not created yet
		}
		// Check the status of the e2e-suite container specifically
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name == "e2e-suite" {
				// Check for image pull failures first
				if containerStatus.State.Waiting != nil {
					reason := containerStatus.State.Waiting.Reason
					if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
						return false, fmt.Errorf("failed to pull image: %s", image)
					}
				}
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
	})
}

func (e *TestJobRunner) fetchArtifacts(ctx context.Context, name, namespace string) error {
	clientSet, err := kubernetes.NewForConfig(e.oc.GetConfig())
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}

	pod, err := e.findJobPod(ctx, name, namespace)
	if err != nil {
		return fmt.Errorf("finding job pod: %w", err)
	}

	if err := e.fetchPodLogs(ctx, clientSet, pod, name); err != nil {
		return fmt.Errorf("fetching pod logs: %w", err)
	}

	if err := e.fetchArtifactFiles(ctx, clientSet, pod); err != nil {
		return fmt.Errorf("fetching artifact files: %w", err)
	}

	return nil
}

func (e *TestJobRunner) findJobPod(ctx context.Context, jobName, namespace string) (*corev1.Pod, error) {
	pods := new(corev1.PodList)
	if err := e.oc.WithNamespace(namespace).List(ctx, pods, resources.WithLabelSelector(labels.FormatLabels(map[string]string{"job-name": jobName}))); err != nil {
		return nil, fmt.Errorf("listing pods for job: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, errors.New("pod for job not found")
	}

	return &pods.Items[0], nil
}

func (e *TestJobRunner) fetchPodLogs(ctx context.Context, clientSet *kubernetes.Clientset, pod *corev1.Pod, jobName string) error {
	var logs strings.Builder

	req := clientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: "e2e-suite"})
	logStream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("getting logs: %w", err)
	}
	defer logStream.Close()

	logBytes, err := io.ReadAll(logStream)
	if err != nil {
		return fmt.Errorf("reading logs: %w", err)
	}

	logs.Write(logBytes)
	logs.WriteString("\n")

	if err = os.WriteFile(filepath.Join(e.cfg.OutputDir, jobName+".log"), []byte(logs.String()), os.ModePerm); err != nil {
		return fmt.Errorf("writing pod logs: %w", err)
	}

	return nil
}

func (e *TestJobRunner) fetchArtifactFiles(ctx context.Context, clientSet *kubernetes.Clientset, pod *corev1.Pod) error {
	execRequest := clientSet.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
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
