package runner

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"

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
	Env: []kubev1.EnvVar{
		{
			Name:  "KUBECONFIG",
			Value: "~/.kube/config",
		},
	},
	Ports: []kubev1.ContainerPort{
		{
			Name:          resultsPortName,
			ContainerPort: resultsPort,
			Protocol:      kubev1.ProtocolTCP,
		},
	},
	ImagePullPolicy: kubev1.PullAlways,
	ReadinessProbe: &kubev1.Probe{
		Handler: kubev1.Handler{
			HTTPGet: &kubev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(resultsPort),
			},
		},
		PeriodSeconds: 7,
	},
	SecurityContext: &kubev1.SecurityContext{
		RunAsUser: pointer.Int64Ptr(0),
	},
}

var payloadVolumeMounts = []kubev1.VolumeMount{
	{
		Name:      osde2ePayload,
		MountPath: osde2ePayloadMountPath,
	},
}

var payloadVolumes = []kubev1.Volume{
	{
		Name: osde2ePayload,
		VolumeSource: kubev1.VolumeSource{
			ConfigMap: &kubev1.ConfigMapVolumeSource{
				LocalObjectReference: kubev1.LocalObjectReference{
					Name: osde2ePayload,
				},
				DefaultMode: pointer.Int32Ptr(0755),
			},
		},
	},
}

// createPod for running commands
func (r *Runner) createPod() (pod *kubev1.Pod, err error) {
	// configure pod to run workload
	pod = &kubev1.Pod{
		ObjectMeta: r.meta(),
		Spec:       r.PodSpec,
	}

	if len(r.Cmd) != 0 {
		cmd, err := r.Command()
		if err != nil {
			return nil, fmt.Errorf("couldn't template cmd: %v", err)
		}

		configMap, err := r.Kube.CoreV1().ConfigMaps(r.Namespace).Create(&kubev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: osde2ePayload,
			},
			BinaryData: map[string][]byte{
				osde2ePayloadScript: cmd,
			},
		})

		if err != nil {
			return nil, fmt.Errorf("error creating ConfigMap: %v", err)
		}

		err = wait.PollImmediate(5*time.Second, configMapCreateTimeout, func() (done bool, err error) {
			if configMap, err = r.Kube.CoreV1().ConfigMaps(r.Namespace).Get(configMap.Name, metav1.GetOptions{}); err != nil {
				log.Printf("Error creating %s config map: %v", configMap.Name, err)
			}
			return err == nil, nil
		})

		if err != nil {
			return nil, err
		}
	}

	for i, container := range pod.Spec.Containers {
		if container.Name == "" || container.Name == r.Name {
			pod.Spec.Containers[i].Name = r.Name
			pod.Spec.Containers[i].Image = r.ImageName

			// run command in pod if, present
			if len(r.Cmd) != 0 {
				pod.Spec.Containers[i].Command = []string{
					fullPayloadScriptPath,
				}
				pod.Spec.Containers[i].VolumeMounts = payloadVolumeMounts
			}
			pod.Spec.Volumes = payloadVolumes
		}
	}

	// setup git repos to be cloned in init containers
	r.Repos.ConfigurePod(&pod.Spec)

	// retry until Pod can be created or timeout occurs
	var createdPod *kubev1.Pod
	err = wait.PollImmediate(5*time.Second, podCreateTimeout, func() (done bool, err error) {
		if createdPod, err = r.Kube.CoreV1().Pods(r.Namespace).Create(pod); err != nil {
			log.Printf("Error creating %s runner Pod: %v", r.Name, err)
		}
		return err == nil, nil
	})
	return createdPod, err
}

func (r *Runner) waitForPodRunning(pod *kubev1.Pod) error {
	var pendingCount int = 0
	return wait.PollImmediate(10*time.Second, 3*time.Minute, func() (done bool, err error) {
		pod, err = r.Kube.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil && !kerror.IsNotFound(err) {
			return
		} else if pod == nil {
			err = errors.New("pod can't be nil")
		} else if pod.Status.Phase == kubev1.PodFailed || pod.Status.Phase == kubev1.PodUnknown {
			err = errors.New("failed waiting for Pod: the Pod has failed")
		} else if pod.Status.Phase == kubev1.PodRunning {
			done = true
		} else {
			pendingCount++
			if pendingCount > 20 {
				err = errors.New("timed out waiting for pod to start")
			}
			r.Printf("Waiting for Pod '%s/%s' to start Running...", pod.Namespace, pod.Name)
		}
		return
	})
}
