package runner

import (
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

func (r *Runner) waitForEndpoints() error {
	var endpoints *kubev1.Endpoints
	endpointsReadyCondition := func() (done bool, err error) {
		endpoints, err = r.Kube.CoreV1().Endpoints(r.svc.Namespace).Get(r.svc.Name, metav1.GetOptions{})
		if err != nil && !kerror.IsNotFound(err) {
			return
		} else if endpoints != nil {
			for _, subset := range endpoints.Subsets {
				if len(subset.Addresses) > 0 {
					return true, nil
				}
			}
		}
		r.Printf("Waiting for test results using Endpoint '%s/%s'...", endpoints.Namespace, endpoints.Name)
		return false, nil
	}
	return wait.PollImmediateUntil(30*time.Second, endpointsReadyCondition, r.stopCh)
}
