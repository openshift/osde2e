package helper

import (
	"log"
	"time"

	. "github.com/onsi/gomega"

	kubev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitForPodPhase until in target, checking n times and sleeping dur between them. Last known phase is returned.
func (h *H) WaitForPodPhase(pod *kubev1.Pod, target kubev1.PodPhase, n int, dur time.Duration) (phase kubev1.PodPhase) {
	var err error
	for i := 0; i < n; i++ {
		if pod, err = h.Kube().CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{}); err != nil {
			log.Println(err.Error())
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

// PollForHealthyPods will check for 100% healthy pods every interval(seconds) for a timeout(minutes) or error
func (h *H) PollForHealthyPods(interval, timeout int) (err error) {
	var (
		requiredRatio float64 = 100
		curRatio      float64
		notReady      []v1.Pod
	)
	parsedInterval := time.Duration(interval) * time.Second
	parsedTimeout := time.Duration(timeout) * time.Minute

	err = wait.Poll(parsedInterval, parsedTimeout, func() (done bool, err error) {
		list, err := h.Kube().CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		notReady = nil
		for _, pod := range list.Items {
			phase := pod.Status.Phase
			if phase != v1.PodRunning && phase != v1.PodSucceeded {
				notReady = append(notReady, pod)
			}
		}

		total := len(list.Items)
		ready := float64(total - len(notReady))
		curRatio = (ready / float64(total)) * 100

		if curRatio != 0 {
			log.Printf("Current status of pods that are running/completed: (%f%%)...", curRatio)
		}

		return len(notReady) == 0, nil
	})

	if err != nil {
		log.Printf("only %f%% of Pods ready, need %f%%. Not ready: %v", curRatio, requiredRatio, len(notReady))
		return err
	}

	return nil
}
