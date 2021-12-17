package cluster

import (
	"context"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ns = "default"
	cm = "osde2e-pr-check-queue"
)

func PrCheckQueue() error {
	var (
		StatusOSDe2eRunning = "osde2e-running"
		interval            = 5 * time.Minute
		timeout             = 2*time.Hour + 30*time.Minute
	)
	h := helper.NewOutsideGinkgo()

	configMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm,
			Namespace: ns,
		},
		Data: map[string]string{StatusOSDe2eRunning: "True"},
	}

	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		//Check that the configmap is not present
		if _, err := h.Kube().CoreV1().ConfigMaps(ns).Get(context.TODO(), cm, metav1.GetOptions{}); err == nil {
			log.Printf("ConfigMap already exists")
			return false, nil
		}

		//Create the configmap
		if _, err := h.Kube().CoreV1().ConfigMaps(ns).Create(context.TODO(), &configMap, metav1.CreateOptions{}); err != nil {
			log.Printf("Couldn't create ConfigMap: %v", err)
			return false, err
		} else {
			return true, nil
		}
	})
	if err != nil {
		log.Printf("Couldn't create ConfigMap: %v", err)
		return err
	}
	return nil
}

func PrCheckQueueCleanup() error {
	h := helper.NewOutsideGinkgo()
	if _, err := h.Kube().CoreV1().ConfigMaps(ns).Get(context.TODO(), cm, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("ConfigMap already deleted: %w", err)
	}
	if err := h.Kube().CoreV1().ConfigMaps(ns).Delete(context.TODO(), cm, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("Couldn't delete ConfigMap: %v", err)
	}
	return nil
}
