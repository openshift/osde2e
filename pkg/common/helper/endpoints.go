package helper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WaitForEndpointReady until Endpoint for svc is ready, checking n times and sleeping dur between them.
func (h *H) WaitForEndpointReady(svc *kubev1.Service, n int, dur time.Duration) error {
	if svc == nil {
		return errors.New("svc was nil")
	}

	for i := 0; i < n; i++ {
		if endpoints, err := h.Kube().CoreV1().Endpoints(svc.Namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{}); err != nil {
			log.Println(err)
		} else if endpoints != nil {
			for _, subset := range endpoints.Subsets {
				if len(subset.Addresses) > 0 {
					return nil
				}
			}
		}

		log.Printf("Waiting for Endpoint '%s/%s' to be ready...", svc.Namespace, svc.Name)
		time.Sleep(dur)
	}

	return fmt.Errorf("timeout waiting for Endpoint '%s/%s' to be ready", svc.Namespace, svc.Name)
}
