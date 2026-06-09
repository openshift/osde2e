package mcsimulation

import (
	"context"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const testNS = "openshift-test-operator"

const testDeploymentName = "controller-manager"

func newDeployment(ready int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testDeploymentName,
			Namespace: testNS,
			Labels:    map[string]string{"app": "test-operator"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test-operator"},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: ready,
		},
	}
}

func newPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNS,
			Labels:    map[string]string{"app": "test-operator"},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
}

func TestRestartOperator_DeletesPodsAndWaitsReady(t *testing.T) {
	deploy := newDeployment(1)
	pod := newPod("controller-manager-abc12")
	client := fake.NewSimpleClientset(deploy, pod)

	err := RestartOperator(context.Background(), client, testNS, "app=test-operator")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Pod should have been deleted.
	pods, _ := client.CoreV1().Pods(testNS).List(context.Background(), metav1.ListOptions{})
	if len(pods.Items) != 0 {
		t.Errorf("expected pod to be deleted, got %d pods", len(pods.Items))
	}
}

func TestRestartOperator_NoDeployment(t *testing.T) {
	client := fake.NewSimpleClientset()

	err := RestartOperator(context.Background(), client, testNS, "app=nonexistent")
	if err == nil || !strings.Contains(err.Error(), "no deployment found") {
		t.Fatalf("expected 'no deployment found' error, got: %v", err)
	}
}

func TestRestartOperator_MultipleDeployments(t *testing.T) {
	deploy1 := newDeployment(1)
	deploy2 := newDeployment(1)
	deploy2.Name = "controller-manager-2"
	client := fake.NewSimpleClientset(deploy1, deploy2)

	err := RestartOperator(context.Background(), client, testNS, "app=test-operator")
	if err == nil || !strings.Contains(err.Error(), "multiple deployments") {
		t.Fatalf("expected 'multiple deployments' error, got: %v", err)
	}
}

func TestRestartOperator_NoPods(t *testing.T) {
	deploy := newDeployment(1)
	client := fake.NewSimpleClientset(deploy)

	// No pods exist — should succeed without error (nothing to restart).
	err := RestartOperator(context.Background(), client, testNS, "app=test-operator")
	if err != nil {
		t.Fatalf("unexpected error when no pods: %v", err)
	}
}

func TestRestartOperator_ReadinessTimeout(t *testing.T) {
	deploy := newDeployment(0) // never ready
	pod := newPod("controller-manager-abc12")
	client := fake.NewSimpleClientset(deploy, pod)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := RestartOperator(ctx, client, testNS, "app=test-operator")
	if err == nil || !strings.Contains(err.Error(), "did not become ready") {
		t.Fatalf("expected 'did not become ready' error, got: %v", err)
	}
}

func TestWaitForDeploymentReady_AlreadyReady(t *testing.T) {
	deploy := newDeployment(1)
	client := fake.NewSimpleClientset(deploy)

	err := waitForDeploymentReady(context.Background(), client, testNS, deploy.Name, 5*time.Second)
	if err != nil {
		t.Fatalf("expected no error for already-ready deployment, got: %v", err)
	}
}
