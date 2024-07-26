package runner

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/pkg/common/util"
	kubev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/pointer"
)

const (
	configMapCreateTimeout = 30 * time.Second
	podCreateTimeout       = 90 * time.Second
	podPendingTimeout      = 20 // 20 Iterations
	fastPoll               = 5 * time.Second
	slowPoll               = 15 * time.Second

	resultsPort     = 8000
	resultsPortName = "results"

	osde2ePayload          = "osde2e-payload"
	osde2ePayloadMountPath = "/osde2e-payload"
	osde2ePayloadScript    = "payload.sh"
)

var fullPayloadScriptPath string

func init() {
	fullPayloadScriptPath = filepath.Join(osde2ePayloadMountPath, osde2ePayloadScript)
}

// DefaultContainer is used by the DefaultRunner to run workloads
var DefaultContainer = kubev1.Container{
	Ports: []kubev1.ContainerPort{
		{
			Name:          resultsPortName,
			ContainerPort: resultsPort,
			Protocol:      kubev1.ProtocolTCP,
		},
	},
	ImagePullPolicy: kubev1.PullAlways,
	ReadinessProbe: &kubev1.Probe{
		ProbeHandler: kubev1.ProbeHandler{
			HTTPGet: &kubev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(resultsPort),
			},
		},
		PeriodSeconds: 7,
	},
	SecurityContext: &kubev1.SecurityContext{
		RunAsUser: pointer.Int64(0),
	},
}

// volumeMounts returns a v1.VolumeMount given a specific name and the static payloadMountPath
func volumeMounts(name string) []kubev1.VolumeMount {
	return []kubev1.VolumeMount{
		{
			Name:      name,
			MountPath: osde2ePayloadMountPath,
		},
	}
}

// volumes returns a v1.Volume given a specific name
func volumes(name string) []kubev1.Volume {
	return []kubev1.Volume{
		{
			Name: name,
			VolumeSource: kubev1.VolumeSource{
				ConfigMap: &kubev1.ConfigMapVolumeSource{
					LocalObjectReference: kubev1.LocalObjectReference{
						Name: name,
					},
					DefaultMode: pointer.Int32(0o755),
				},
			},
		},
	}
}

// createJobPod for creates an osde2e runner Job and returns the singular job pod for further processing.
// Opting for Job creation instead of direct Pod creation to avoid orphan workloads.
func (r *Runner) createJobPod(ctx context.Context) (pod *kubev1.Pod, err error) {
	// configure pod to run workload
	cmName := fmt.Sprintf("%s-%s", osde2ePayload, util.RandomStr(5))
	pod = &kubev1.Pod{
		ObjectMeta: r.meta(),
		Spec:       r.PodSpec,
	}

	if len(r.Cmd) != 0 {
		cmd, err := r.Command()
		if err != nil {
			return nil, fmt.Errorf("couldn't template cmd: %v", err)
		}

		configMap, err := r.Kube.CoreV1().ConfigMaps(r.Namespace).Create(ctx, &kubev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: cmName,
			},
			BinaryData: map[string][]byte{
				osde2ePayloadScript: cmd,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("error creating ConfigMap: %v", err)
		}

		// Verify the configMap has been created before proceeding
		err = wait.PollUntilContextTimeout(ctx, fastPoll, configMapCreateTimeout, false, func(ctx context.Context) (done bool, err error) {
			if configMap, err = r.Kube.CoreV1().ConfigMaps(r.Namespace).Get(ctx, configMap.Name, metav1.GetOptions{}); err != nil {
				r.Error(err, fmt.Sprintf("error creating %s config map", configMap.Name))
			}
			return err == nil, nil
		})
		if err != nil {
			return nil, err
		}
	}

	// A pod can have multiple containers. Create all the necessary mounts per-container.
	for i, container := range pod.Spec.Containers {
		if container.Name == "" || container.Name == r.Name {
			pod.Spec.Containers[i].Name = r.Name
			pod.Spec.Containers[i].Image = r.ImageName
			pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, kubev1.EnvVar{
				Name:  "OSDE2E",
				Value: "true",
			})

			// run command in pod if, present
			if len(r.Cmd) != 0 {
				pod.Spec.Containers[i].Command = []string{
					fullPayloadScriptPath,
				}
				pod.Spec.Containers[i].VolumeMounts = volumeMounts(cmName)
			}
			pod.Spec.Volumes = volumes(cmName)
		}
	}

	// setup git repos to be cloned in init containers
	r.Repos.ConfigurePod(&pod.Spec)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Name,
		},
		Spec: batchv1.JobSpec{
			Template: kubev1.PodTemplateSpec{
				Spec: pod.Spec,
			},
		},
	}
	// retry until Job can be created or timeout occurs
	if err = wait.PollUntilContextTimeout(ctx, fastPoll, podCreateTimeout, false, func(ctx context.Context) (done bool, err error) {
		if job, err = r.Kube.BatchV1().Jobs(r.Namespace).Create(ctx, job, metav1.CreateOptions{}); err != nil {
			r.Error(err, fmt.Sprintf("error creating %s/%s runner Job", r.Namespace, r.Name))
		}
		return err == nil, nil
	}); err != nil {
		return nil, err
	}
	pods := new(corev1.PodList)
	// retry until Pod can be found or timeout occurs
	if err = wait.PollUntilContextTimeout(ctx, fastPoll, podCreateTimeout, false, func(ctx context.Context) (done bool, err error) {
		labelSelector := fmt.Sprintf("%s=%s", "job-name", job.Name)
		pods, err = r.Kube.CoreV1().Pods(r.Namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil || len(pods.Items) < 1 {
			err = fmt.Errorf("failed to list pods for job %s in %s namespace: %w", r.Name, r.Namespace, err)
		}
		return err == nil, nil
	}); err != nil {
		return nil, err
	}
	// return first job pod; test job has single runner pod
	return &pods.Items[0], err
}

// waitForRunningPod, given a v1.Pod, will wait for 3 minutes for a pod to enter the running phase or return an error.
func (r *Runner) waitForPodRunning(ctx context.Context, pod *kubev1.Pod) error {
	pendingCount := 0
	return wait.PollUntilContextTimeout(ctx, fastPoll, 3*time.Minute, false, func(ctx context.Context) (done bool, err error) {
		pod, err = r.Kube.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil && !kerror.IsNotFound(err) {
			return
		} else if pod == nil {
			err = errors.New("pod can't be nil")
		} else if pod.Status.Phase == kubev1.PodFailed || pod.Status.Phase == kubev1.PodUnknown {
			err = fmt.Errorf("failed waiting for Pod: the Pod has a phase of %s", pod.Status.Phase)
		} else if pod.Status.Phase == kubev1.PodRunning {
			done = true
		} else {
			pendingCount++
			if pendingCount > podPendingTimeout {
				err = errors.New("timed out waiting for pod to start")
			}
			r.Info(fmt.Sprintf("Waiting for Pod '%s/%s' to start Running...", pod.Namespace, pod.Name))
		}
		return
	})
}
