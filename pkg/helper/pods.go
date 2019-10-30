package helper

import (
	"log"
	"time"

	. "github.com/onsi/gomega"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// WaitForPodPhase until in target, checking n times and sleeping dur between them. Last known phase is returned.
func (h *H) WaitForPodPhase(pod *kubev1.Pod, target kubev1.PodPhase, n int, dur time.Duration) (phase kubev1.PodPhase) {
	var err error
	for i := 0; i < n; i++ {
		if pod, err = h.Kube().CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{}); err != nil {
			log.Println(err)
		} else if pod != nil {
			phase = pod.Status.Phase

			// stop checking if Pod has reached state or failed
			if phase == target || phase == kubev1.PodFailed {
				return
			}
		}

		log.Printf("Waiting for Pod '%s/%s' to be %s, currently %s...", pod.Namespace, pod.Name, target, phase)
		time.Sleep(dur)
	}

	Expect(phase).NotTo(BeEmpty())
	return
}

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
