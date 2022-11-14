package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/hashicorp/go-multierror"
	kubev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	containerLogs = "containerLogs"
)

// createService returns a v1.Service pointing to a given v1.Pod object
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

	return r.Kube.CoreV1().Services(r.Namespace).Create(context.TODO(), &kubev1.Service{
		ObjectMeta: r.meta(),
		Spec: kubev1.ServiceSpec{
			Selector: pod.Labels,
			Ports:    ports,
		},
	}, metav1.CreateOptions{})
}

// waitForCompletion will wait for a runner's pod to have a valid v1.Endpoint available
func (r *Runner) waitForCompletion(podName string, timeoutInSeconds int) error {
	var endpoints *kubev1.Endpoints
	var pendingCount int = 0
	return wait.PollImmediate(slowPoll, time.Duration(timeoutInSeconds)*time.Second, func() (done bool, err error) {
		endpoints, err = r.Kube.CoreV1().Endpoints(r.svc.Namespace).Get(context.TODO(), r.svc.Name, metav1.GetOptions{})
		if err != nil && !kerror.IsNotFound(err) {
			r.Printf("Encountered error getting endpoint '%s/%s': %v", r.svc.Namespace, r.svc.Name, err)
		} else if endpoints != nil {
			for _, subset := range endpoints.Subsets {
				if len(subset.Addresses) > 0 {
					return true, nil
				}
			}
		}
		pod, err := r.Kube.CoreV1().Pods(r.svc.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			r.Printf("Encountered error getting pod: %v", err)
			return false, err
		}

		if pod.Status.Phase == kubev1.PodFailed || pod.Status.Phase == kubev1.PodUnknown {
			r.Printf("Pod entered error state while waiting for endpoint: %+v", pod.Status)
			return false, fmt.Errorf("pod failed while waiting for endpoints")
		} else if pod.Status.Phase == kubev1.PodSucceeded {
			var err *multierror.Error
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Terminated != nil {
					if containerStatus.State.Terminated.ExitCode != 0 {
						err = multierror.Append(err, fmt.Errorf("container %s failed, please refer to artifacts for results", containerStatus.Name))
					}
				}
			}
			return err == nil, err.ErrorOrNil()
		} else if pod.Status.Phase == kubev1.PodPending {
			pendingCount++
			if pendingCount > podPendingTimeout {
				return false, fmt.Errorf("timed out waiting for pod to start")
			}
		}

		r.Printf("Waiting for test results using Endpoint '%s/%s'...", endpoints.Namespace, endpoints.Name)
		return false, nil
	})
}

func (r *Runner) getAllLogsFromPod(podName string) error {
	pod, err := r.Kube.CoreV1().Pods(r.svc.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var allErrors *multierror.Error
	for _, containerStatus := range pod.Status.ContainerStatuses {
		func() {
			log.Printf("Trying to get logs for %s:%s", podName, containerStatus.Name)
			request := r.Kube.CoreV1().Pods(r.svc.Namespace).GetLogs(podName, &kubev1.PodLogOptions{Container: containerStatus.Name})

			logStream, err := request.Stream(context.TODO())
			if err != nil {
				allErrors = multierror.Append(allErrors, err)
				return
			}

			defer logStream.Close()

			logBytes, err := ioutil.ReadAll(logStream)
			if err != nil {
				allErrors = multierror.Append(allErrors, err)
				return
			}

			configMapDirectory := filepath.Join(viper.GetString(config.ReportDir), viper.GetString(config.Phase), containerLogs)

			if err := os.MkdirAll(configMapDirectory, os.FileMode(0o755)); err != nil {
				allErrors = multierror.Append(allErrors, err)
				return
			}

			logOutput := filepath.Join(configMapDirectory, fmt.Sprintf("%s-%s.log", podName, containerStatus.Name))

			allErrors = multierror.Append(allErrors, ioutil.WriteFile(logOutput, logBytes, os.FileMode(0o644)))
		}()
	}

	return allErrors.ErrorOrNil()
}
