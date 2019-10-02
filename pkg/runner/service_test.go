package runner

import (
	"fmt"
	"testing"
	"time"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResultsService(t *testing.T) {
	// setup mock client
	client := fake.NewSimpleClientset()

	// setup runner
	def := *DefaultRunner
	r := &def
	r.Kube = client

	// create results service
	var err error
	r.svc, err = r.createService(new(kubev1.Pod))
	if err != nil {
		t.Fatalf("Failed to create example service: %v", err)
	}

	// start waiting for endpoint Ready
	done := make(chan struct{})
	errs := make(chan error, 1)
	go func() {
		err := r.waitForEndpoints()
		if err != nil {
			errs <- fmt.Errorf("Failed waiting for endpoints: %v", err)
		} else {
			done <- struct{}{}
		}
	}()

	// create endpoint
	endpoints, err := client.CoreV1().Endpoints(r.Namespace).Create(&kubev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.svc.Name,
		},
	})
	if err != nil {
		t.Fatalf("Failed getting created endpoint: %v", err)
	}

	// wait some time and add address
	time.Sleep(10 * time.Second)
	address := kubev1.EndpointAddress{
		IP:       "127.0.0.1",
		Hostname: "localhost",
	}
	endpoints.Subsets = []kubev1.EndpointSubset{
		{
			Addresses: []kubev1.EndpointAddress{address},
		},
	}
	_, err = r.Kube.CoreV1().Endpoints(r.Namespace).Update(endpoints)
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
