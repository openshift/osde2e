package mcsimulation

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// RestartOperator deletes pods matching labelSelector in the given namespace,
// then waits for the owning Deployment to reach ready state. This is used
// after CRD installation so the operator detects the newly registered CRDs.
//
// The Deployment's own replica controller recreates deleted pods automatically,
// making this equivalent to "kubectl rollout restart" without mutating the
// Deployment spec.
func RestartOperator(ctx context.Context, clientset kubernetes.Interface, namespace, labelSelector string) error {
	// Find the Deployment that owns these pods so we can wait on its readiness.
	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("listing deployments in %s: %w", namespace, err)
	}
	if len(deployments.Items) == 0 {
		return fmt.Errorf("no deployment found in %s matching %q", namespace, labelSelector)
	}
	if len(deployments.Items) > 1 {
		names := make([]string, len(deployments.Items))
		for i := range deployments.Items {
			names[i] = deployments.Items[i].Name
		}
		return fmt.Errorf("multiple deployments found in %s matching %q, expected exactly one: %v", namespace, labelSelector, names)
	}
	deploy := &deployments.Items[0]

	// Resolve the Deployment's pod selector to find matching pods.
	// Use LabelSelectorAsSelector to honour both matchLabels and matchExpressions.
	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		return fmt.Errorf("converting deployment selector: %w", err)
	}
	podSelector := selector.String()
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: podSelector,
	})
	if err != nil {
		return fmt.Errorf("listing pods in %s: %w", namespace, err)
	}
	if len(pods.Items) == 0 {
		klog.InfoS("No pods found to restart", "namespace", namespace, "selector", podSelector)
		return nil
	}

	for i := range pods.Items {
		name := pods.Items[i].Name
		klog.V(2).InfoS("Deleting pod to trigger restart", "pod", name, "namespace", namespace)
		if err := clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				klog.V(2).InfoS("Pod already deleted", "pod", name, "namespace", namespace)
				continue
			}
			return fmt.Errorf("deleting pod %s: %w", name, err)
		}
	}

	// Wait for the Deployment to have at least one ready replica.
	klog.V(2).InfoS("Waiting for deployment readiness after restart", "deployment", deploy.Name, "namespace", namespace)
	if err := waitForDeploymentReady(ctx, clientset, namespace, deploy.Name, 2*time.Minute); err != nil {
		return fmt.Errorf("deployment %s did not become ready: %w", deploy.Name, err)
	}
	klog.V(2).InfoS("Deployment restarted successfully", "deployment", deploy.Name)
	return nil
}

// waitForDeploymentReady polls until the named Deployment has at least one
// ready replica or the timeout expires.
func waitForDeploymentReady(ctx context.Context, clientset kubernetes.Interface, namespace, name string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		deploy, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return deploy.Status.ReadyReplicas > 0, nil
	})
}
