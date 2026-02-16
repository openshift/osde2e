package runner

import (
	"context"
	"fmt"
	"io"
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
func (r *Runner) createService(ctx context.Context, pod *kubev1.Pod) (svc *kubev1.Service, err error) {
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

	return r.Kube.CoreV1().Services(r.Namespace).Create(ctx, &kubev1.Service{
		ObjectMeta: r.meta(),
		Spec: kubev1.ServiceSpec{
			Selector: pod.Labels,
			Ports:    ports,
		},
	}, metav1.CreateOptions{})
}

// waitForCompletion will wait for a runner's pod to have a valid v1.Endpoint available
// otherwise waits for pod to be running
func (r *Runner) waitForCompletion(ctx context.Context, podName string, timeoutInSeconds int) error {
	var endpoints *kubev1.Endpoints
	pendingCount := 0
	return wait.PollUntilContextTimeout(ctx, slowPoll, time.Duration(timeoutInSeconds)*time.Second, false, func(ctx context.Context) (done bool, err error) {
		endpoints, err = r.Kube.CoreV1().Endpoints(r.svc.Namespace).Get(ctx, r.svc.Name, metav1.GetOptions{})
		if err != nil && !kerror.IsNotFound(err) {
			r.Error(err, fmt.Sprintf("unable to get endpoint '%s/%s'", r.svc.Namespace, r.svc.Name))
		} else if endpoints != nil {
			for _, subset := range endpoints.Subsets {
				if len(subset.Addresses) > 0 {
					return true, nil
				}
			}
		}
		pod, err := r.Kube.CoreV1().Pods(r.svc.Namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			r.Error(err, fmt.Sprintf("unable to get pod %s/%s", r.svc.Namespace, podName))
			return false, err
		}

		switch pod.Status.Phase {
		case kubev1.PodFailed:
			r.Info(fmt.Sprintf("Pod entered error state while waiting for endpoint: %+v", pod.Status))
			return false, fmt.Errorf("pod failed while waiting for endpoints")
		case kubev1.PodSucceeded:
			var err *multierror.Error
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Terminated != nil {
					if containerStatus.State.Terminated.ExitCode != 0 {
						err = multierror.Append(err, fmt.Errorf("container %s failed, please refer to artifacts for results", containerStatus.Name))
					}
				}
			}
			return err == nil, err.ErrorOrNil()
		case kubev1.PodPending:
			pendingCount++
			if pendingCount > podPendingTimeout {
				return false, fmt.Errorf("timed out waiting for pod to start")
			}
		}

		r.Info(fmt.Sprintf("Pod state: %s; polling endpoint '%s'...", pod.Status.Phase, endpoints.Name))
		return false, nil
	})
}

func (r *Runner) getAllLogsFromPod(ctx context.Context, podName string) error {
	pod, err := r.Kube.CoreV1().Pods(r.svc.Namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var allErrors *multierror.Error
	for _, containerStatus := range pod.Status.ContainerStatuses {
		func() {
			r.Info(fmt.Sprintf("Trying to get logs for %s:%s", podName, containerStatus.Name))
			request := r.Kube.CoreV1().Pods(r.svc.Namespace).GetLogs(podName, &kubev1.PodLogOptions{Container: containerStatus.Name})

			logStream, err := request.Stream(ctx)
			if err != nil {
				allErrors = multierror.Append(allErrors, err)
				return
			}

			defer logStream.Close()

			logBytes, err := io.ReadAll(logStream)
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

			allErrors = multierror.Append(allErrors, os.WriteFile(logOutput, logBytes, os.FileMode(0o644)))
		}()
	}

	return allErrors.ErrorOrNil()
}
