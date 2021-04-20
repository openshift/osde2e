package helper

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitTimeoutForDaemonSetInNamespace waits the given timeout duration for the specified ds.
func (h *H) WaitTimeoutForDaemonSetInNamespace(daemonSetName string, namespace string, timeout time.Duration, poll time.Duration) error {

	return (wait.PollImmediate(poll, timeout, func() (bool, error) {

		if _, err := h.Kube().AppsV1().DaemonSets(namespace).Get(context.TODO(), daemonSetName, metav1.GetOptions{}); err != nil {
			return false, err
		}
		return true, nil
	}))
}

// WaitTimeoutForServiceInNamespace waits the given timeout duration for the specified service.
func (h *H) WaitTimeoutForServiceInNamespace(serviceName string, namespace string, timeout time.Duration, poll time.Duration) error {
	return (wait.PollImmediate(poll, timeout, func() (bool, error) {

		if _, err := h.Kube().CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{}); err != nil {
			return false, err
		}
		return true, nil
	}))
}
