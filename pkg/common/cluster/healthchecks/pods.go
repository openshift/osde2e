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

// CheckClusterPodHealth attempts to look at the state of all internal cluster pods and
// returns true if things are healthy.
func CheckClusterPodHealth(podClient v1.CoreV1Interface, logger *log.Logger) ([]kubev1.Pod, error) {
	filters := []PodPredicate{
		IsClusterPod,
		IsNotReadinessPod,
		IsNotRunning,
		IsNotCompleted,
	}
	podlist, err := checkPods(podClient, logger, filters...)
	if err != nil {
		return nil, err
	}

	return podlist, err
}

// CheckPodHealth attempts to look at the state of all pods and returns true if things are healthy.
func CheckPodHealth(podClient v1.CoreV1Interface, logger *log.Logger, ns string, podPrefixes ...string) (bool, error) {
	filters := []PodPredicate{
		MatchesNamespace(ns),
		MatchesNames(podPrefixes...),
		IsNotReadinessPod,
		IsNotRunning,
		IsNotCompleted,
	}
	podlist, err := checkPods(podClient, logger, filters...)
	if err != nil {
		return false, err
	}
	return !(len(podlist) > 0), err
}

// checkPods looks for pods matching the supplied predicates and returns true if any are found
func checkPods(podClient v1.CoreV1Interface, logger *log.Logger, filters ...PodPredicate) ([]kubev1.Pod, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	logger.Print("Checking that all Pods are running or completed...")

	listOpts := metav1.ListOptions{}
	list, err := podClient.Pods(metav1.NamespaceAll).List(context.TODO(), listOpts)
	if err != nil {
		return nil, fmt.Errorf("error getting pod list: %v", err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("pod list is empty. this should NOT happen")
	}

	pods := filterPods(list, filters...)

	logger.Printf("%v pods are currently not running or complete:", len(pods.Items))

	for _, pod := range pods.Items {
		if pod.Status.Phase != kubev1.PodPending {
			return nil, fmt.Errorf("Pod %s errored: %s - %s", pod.GetName(), pod.Status.Reason, pod.Status.Message)
		}
		logger.Printf("%s is not ready. Phase: %s, Message: %s, Reason: %s", pod.Name, pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
	}
	return pods.Items, nil
}

func containsPrefixes(str string, subs ...string) bool {
	for _, sub := range subs {
		if strings.HasPrefix(str, sub) {
			return true
		}
	}
	return false
}

func filterPods(podList *kubev1.PodList, predicates ...PodPredicate) *kubev1.PodList {
	filteredPods := &kubev1.PodList{}
	for _, pod := range podList.Items {
		var match = true
		for _, p := range predicates {
			if !p(pod) {
				match = false
				break
			}
		}
		if match {
			filteredPods.Items = append(filteredPods.Items, pod)
		}
	}
	return filteredPods
}

// CheckPendingPods looks for a pod that still remains under pending state for a given number of times in a row
func CheckPendingPods(podlist []kubev1.Pod, podErrorCount map[string]int, pendingPodThreshold int) (bool, error) {
	for _, pod := range podlist {
		if val, found := podErrorCount[pod.Name]; found {
			podErrorCount[pod.Name]++
			if val >= pendingPodThreshold {
				return false, fmt.Errorf("Pod %s is pending beyond normal threshold: %s - %s", pod.GetName(), pod.Status.Reason, pod.Status.Message)
			}
		} else {
			podErrorCount[pod.Name] = 1
		}
	}

	return true, nil
}
