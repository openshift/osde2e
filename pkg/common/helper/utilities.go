package helper

import (
	"context"
	"time"

	kv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitTimeoutForDaemonSetInNamespace waits the given timeout duration for the specified ds.
func (h *H) WaitTimeoutForDaemonSetInNamespace(ctx context.Context, daemonSetName string, namespace string, timeout time.Duration, poll time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, poll, timeout, false, func(ctx context.Context) (bool, error) {
		if _, err := h.Kube().AppsV1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{}); err != nil {
			return false, err
		}
		return true, nil
	})
}

// WaitTimeoutForServiceInNamespace waits the given timeout duration for the specified service.
func (h *H) WaitTimeoutForServiceInNamespace(ctx context.Context, serviceName string, namespace string, timeout time.Duration, poll time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, poll, timeout, false, func(ctx context.Context) (bool, error) {
		svc := &kv1.Service{}
		if err := h.Client.Get(ctx, serviceName, namespace, svc); err != nil {
			return false, err
		}
		return true, nil
	})
}
