package healthchecks

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/openshift/osde2e/pkg/common/logging"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CheckPodHealth attempts to look at the state of all pods and returns true if things are healthy.
func CheckPodHealth(podClient v1.CoreV1Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	var notReady []kubev1.Pod

	logger.Print("Checking that all Pods are running or completed...")

	listOpts := metav1.ListOptions{}
	list, err := podClient.Pods(metav1.NamespaceAll).List(context.TODO(), listOpts)
	if err != nil {
		return false, fmt.Errorf("error getting pod list: %v", err)
	}

	if len(list.Items) == 0 {
		return false, fmt.Errorf("pod list is empty. this should NOT happen")
	}

	total := 0
	for _, pod := range list.Items {
		// we only care about the openshift, redhat, and osde2e namespaces
		if !containsPrefixes(pod.Namespace, "openshift-", "redhat-", "osde2e-") {
			continue
		}
		total++
		phase := pod.Status.Phase

		if phase != kubev1.PodRunning && phase != kubev1.PodSucceeded {
			if phase != kubev1.PodPending {
				return false, fmt.Errorf("Pod %s errored: %s - %s", pod.GetName(), pod.Status.Reason, pod.Status.Message)
			}
			notReady = append(notReady, pod)
			logger.Printf("%s is not ready. Phase: %s, Message: %s, Reason: %s", pod.Name, pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
		}
	}

	ready := float64(total - len(notReady))
	curRatio := (ready / float64(total)) * 100

	logger.Printf("%v%% of %v pods are currently alive...", curRatio, total)

	return len(notReady) == 0, nil
}

func containsPrefixes(str string, subs ...string) bool {
	for _, sub := range subs {
		if strings.HasPrefix(str, sub) {
			return true
		}
	}
	return false
}
