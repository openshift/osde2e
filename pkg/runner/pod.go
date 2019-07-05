package runner

import (
	"errors"
	"fmt"
	"time"

	kubev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	resultsPort     = 8000
	resultsPortName = "results"
)

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
}

// createPod for openshift-tests
func (r *Runner) createPod() (*kubev1.Pod, error) {
	// configure pod to run workload
	pod := &kubev1.Pod{
		ObjectMeta: r.meta(),
		Spec:       r.PodSpec,
	}

	for i, container := range pod.Spec.Containers {
		if container.Name == "" || container.Name == r.Name {
			pod.Spec.Containers[i].Name = r.Name
			pod.Spec.Containers[i].Image = r.ImageName

			// run command in pod if, present
			if len(r.Cmd) != 0 {
				cmd, err := r.Command()
				if err != nil {
					return nil, fmt.Errorf("couldn't template Cmd: %v", err)
				}

				pod.Spec.Containers[i].Args = []string{
					"/bin/bash",
					"-c",
					cmd,
				}
			}
		}
	}

	// setup git repos to be cloned in init containers
	r.Repos.ConfigurePod(pod)

	// create runner pod
	return r.Kube.CoreV1().Pods(r.Namespace).Create(pod)
}

func (r *Runner) waitForPodRunning(pod *kubev1.Pod) error {
	runningCondition := func() (done bool, err error) {
		pod, err = r.Kube.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil && !kerror.IsNotFound(err) {
			return
		} else if pod == nil {
			err = errors.New("pod can't be nil")
		} else if pod.Status.Phase == kubev1.PodFailed {
			err = errors.New("failed waiting for Pod: the Pod has failed")
		} else if pod.Status.Phase == kubev1.PodRunning {
			done = true
		} else {
			r.Printf("Waiting for Pod '%s/%s' to start Running...", pod.Namespace, pod.Name)
		}
		return
	}
	return wait.PollImmediateUntil(10*time.Second, runningCondition, r.stopCh)
}
