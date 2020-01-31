package runner

import (
	"fmt"
	"time"

	kubev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (r *Runner) createService(pod *kubev1.Pod) (svc *kubev1.Service, err error) {
	var ports []kubev1.ServicePort
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			ports = append(ports, kubev1.ServicePort{
				Name:     p.Name,
				Protocol: p.Protocol,
				Port:     p.ContainerPort,
			})
		}
	}

	return r.Kube.CoreV1().Services(r.Namespace).Create(&kubev1.Service{
		ObjectMeta: r.meta(),
		Spec: kubev1.ServiceSpec{
			Selector: pod.Labels,
			Ports:    ports,
		},
	})
}

func (r *Runner) waitForCompletion(timeoutInSeconds int) error {
	var endpoints *kubev1.Endpoints
	var pendingCount int = 0
	return wait.PollImmediate(slowPoll, time.Duration(timeoutInSeconds)*time.Second, func() (done bool, err error) {
		endpoints, err = r.Kube.CoreV1().Endpoints(r.svc.Namespace).Get(r.svc.Name, metav1.GetOptions{})
		if err != nil && !kerror.IsNotFound(err) {
			r.Printf("Encountered error getting endpoint '%s/%s': %v", r.svc.Namespace, r.svc.Name, err)
		} else if endpoints != nil {
			for _, subset := range endpoints.Subsets {
				if len(subset.Addresses) > 0 {
					return true, nil
				}
			}
		}
		pods, err := r.Kube.CoreV1().Pods(r.svc.Namespace).List(metav1.ListOptions{})
		if err != nil {
			r.Printf("Encountered error getting pods: %v", err)
			return false, err
		}
		for _, pod := range pods.Items {
			if pod.Status.Phase == kubev1.PodFailed || pod.Status.Phase == kubev1.PodUnknown {
				r.Printf("Pod entered error state while waiting for endpoint: %+v", pod.Status)
				return false, fmt.Errorf("pod failed while waiting for endpoints")
			} else if pod.Status.Phase == kubev1.PodSucceeded {
				return true, nil
			} else if pod.Status.Phase == kubev1.PodPending {
				pendingCount++
				if pendingCount > podPendingTimeout {
					return false, fmt.Errorf("timed out waiting for pod to start")
				}
			}
		}
		r.Printf("Waiting for test results using Endpoint '%s/%s'...", endpoints.Namespace, endpoints.Name)
		return false, nil
	})
}
