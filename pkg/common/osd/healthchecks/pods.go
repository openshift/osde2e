package healthchecks

import (
	"fmt"
	"log"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CheckPodHealth attempts to look at the state of all pods and returns true if things are healthy.
func CheckPodHealth(podClient v1.CoreV1Interface) (bool, error) {
	var notReady []kubev1.Pod

	log.Print("Checking that all Pods are running or completed...")

	listOpts := metav1.ListOptions{}
	list, err := podClient.Pods(metav1.NamespaceAll).List(listOpts)
	if err != nil {
		return false, err
	}

	if len(list.Items) == 0 {
		return false, err
	}

	for _, pod := range list.Items {
		phase := pod.Status.Phase

		if phase != kubev1.PodRunning && phase != kubev1.PodSucceeded {
			if phase != kubev1.PodPending {
				return false, fmt.Errorf("Pod %s errored: %s - %s", pod.GetName(), pod.Status.Reason, pod.Status.Message)
			}
			notReady = append(notReady, pod)
			log.Printf("%s is not ready. Phase: %s, Message: %s, Reason: %s", pod.Name, pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
		}
	}

	total := len(list.Items)
	ready := float64(total - len(notReady))
	curRatio := (ready / float64(total)) * 100

	log.Printf("%v%% of pods are currently alive: ", curRatio)

	return len(notReady) == 0, nil
}
