package runner

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResultsService(t *testing.T) {
	// setup mock client
	client := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "osde2e-runner",
			},
		},
		Status: corev1.PodStatus{
			PodIP: "172.1.0.3",
		},
	})

	// setup runner
	def := *DefaultRunner
	r := &def
	r.Kube = client
	ctx := context.Background()

	// create pod
	pod, err := r.createJobPod(ctx)
	if err != nil {
		t.Fatalf("Failed to create example  pod: %v", err)
	}

	// create results service
	r.svc, err = r.createService(ctx, pod)
	if err != nil {
		t.Fatalf("Failed to create example service: %v", err)
	}

	// start waiting for endpoint Ready
	done := make(chan struct{})
	errs := make(chan error, 1)
	go func() {
		err := r.waitForCompletion(ctx, pod.Name, 1800)
		if err != nil {
			errs <- fmt.Errorf("Failed waiting for endpoints: %v", err)
		} else {
			done <- struct{}{}
		}
	}()

	// create endpoint
	endpoints, err := client.CoreV1().Endpoints(r.Namespace).Create(context.TODO(), &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.svc.Name,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed getting created endpoint: %v", err)
	}

	// wait some time and add address
	time.Sleep(10 * time.Second)
	address := corev1.EndpointAddress{
		IP:       "127.0.0.1",
		Hostname: "localhost",
	}
	endpoints.Subsets = []corev1.EndpointSubset{
		{
			Addresses: []corev1.EndpointAddress{address},
		},
	}
	_, err = r.Kube.CoreV1().Endpoints(r.Namespace).Update(context.TODO(), endpoints, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update endpoint: %v", err)
	}

	select {
	case err := <-errs:
		t.Fatalf("Failed waiting for endpoint: %v", err)
	case <-done:
		// test passes
	case <-time.After(21 * time.Second):
		t.Fatal("timeout waiting for endpoints")
	}
}
