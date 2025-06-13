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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/utils/ptr"
)

type Config struct {
	ClusterID           string
	Environment         ocm.Environment
	CloudProviderID     string
	CloudProviderRegion string
	Timeout             time.Duration
}

type executioner struct {
	oc        *openshift.Client
	image     string
	outputDir string
}

// New sets up a new executioner to run a given test suite image
func New(logger logr.Logger, image string) (*executioner, error) {
	oc, err := openshift.New(logger)
	if err != nil {
		return nil, fmt.Errorf("openshift client creation: %w", err)
	}
	return &executioner{oc: oc, image: image}, nil
}

func (e *executioner) Execute(ctx context.Context, cfg *Config) error {
	project := &projectv1.Project{}
	if err := e.oc.Create(ctx, project); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	defer func() {
		if err := e.oc.Delete(ctx, project); err != nil {
			panic(err)
		}
	}()

	sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "cluster-admin", Namespace: project.Name}}
	if err := e.oc.Create(ctx, sa); err != nil {
		return fmt.Errorf("creating cluster-admin serviceaccount: %w", err)
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "osde2e-executioner-cluster-admin-",
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(project, project.GroupVersionKind()),
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

	// TODO: can we do this stuff without viper?
	// create secrets for job

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "osde2e-executioner-",
			Namespace:    project.Name,
		},
		Spec: batchv1.JobSpec{
			Parallelism:           ptr.To[int32](1),
			Completions:           ptr.To[int32](1),
			BackoffLimit:          ptr.To[int32](0),
			ActiveDeadlineSeconds: ptr.To(int64(cfg.Timeout.Seconds())),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: sa.Name,
					Containers: []corev1.Container{
						{
							Name:            "e2e-suite",
							Image:           e.image,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "OCM_CLUSTER_ID",
									Value: cfg.ClusterID,
								},
								{
									Name:  "OCM_ENV",
									Value: string(cfg.Environment),
								},
								{
									Name:  "CLOUD_PROVIDER_ID",
									Value: cfg.CloudProviderID,
								},
								{
									Name:  "CLOUD_PROVIDER_REGION",
									Value: cfg.CloudProviderRegion,
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "ci-secrets",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "artifacts",
									MountPath: "/artifacts",
								},
							},
						},
						{
							Name:  "pause-for-artifacts",
							Image: "registry.k8s.io/pause:latest",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "artifacts",
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
	if err := e.oc.Create(ctx, job); err != nil {
		return fmt.Errorf("creating job: %w", err)
	}

	if err := e.oc.WatchJob(ctx, job.Namespace, job.Name); err != nil {
		// if "job failed" then log and continue
		return fmt.Errorf("watching job: %w", err)
	}

	logs, err := e.oc.GetJobLogs(ctx, job.Name, job.Namespace)
	if err != nil {
		return fmt.Errorf("getting job logs: %w", err)
	}

	if err = e.fetchArtifacts(ctx, job.Name, job.Namespace); err != nil {
		return fmt.Errorf("fetching artifacts: %w", err)
	}

	_ = logs

	return errors.New("unimplemented")
}

func (e *executioner) fetchArtifacts(ctx context.Context, name, namespace string) error {
	clientSet, err := kubernetes.NewForConfig(e.oc.GetConfig())
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}

	execRequest := clientSet.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   []string{"tar", "cf", "-", "-C", "/artifacts", "/artifacts"},
			Container: "e2e-suite",
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

	if err = untarBuffer(&stdout, e.outputDir); err != nil {
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
