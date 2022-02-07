package healthchecks

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/logging"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// PodErrorTracker is the structure that keeps count of each pending pod's threshold
type PodErrorTracker struct {
	Counts                  map[string]int
	MaxPendingPodsThreshold int
}

// NewPodErrorTracker initializes the PodErrorTracker structure with a given pending pod threshold and a new pod counter
func (p *PodErrorTracker) NewPodErrorTracker(threshold int) *PodErrorTracker {
	p.Counts = make(map[string]int)
	p.MaxPendingPodsThreshold = threshold
	return p
}

// CheckClusterPodHealth attempts to look at the state of all internal cluster pods and
// returns the list of pending pods if any exist.
func CheckClusterPodHealth(podClient v1.CoreV1Interface, logger *log.Logger) ([]kubev1.Pod, error) {
	filters := []PodPredicate{
		IsOlderThan(1 * time.Minute),
		IsClusterPod,
		IsNotReadinessPod,
		IsNotRunning,
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
		IsOlderThan(1 * time.Minute),
		MatchesNamespace(ns),
		MatchesNames(podPrefixes...),
		IsNotReadinessPod,
		IsNotRunning,
		IsNotCompleted,
		IsNotControlledByJob,
	}
	podlist, err := checkPods(podClient, logger, filters...)
	if err != nil {
		return false, err
	}
	return !(len(podlist) > 0), err
}

// checkPods looks for pods matching the supplied predicates and returns the list of pods (pending pods) if any are found
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

	// Create a map of pods with their cronJob names under the given namespace. If pod with no associated cronJob is not Pending, simply return the error message.
	filteredPods := map[string]map[string]interface{}{}
	namespace := ""
	jobName := 0
	cronJobName := ""
	for _, pod := range pods.Items {
		if pod.ObjectMeta.Labels["job-name"] == "" {
			if pod.Status.Phase != kubev1.PodPending {
				return nil, fmt.Errorf("pod %s in unexpected phase %s: reason: %s message: %s", pod.GetName(), pod.Status.Phase, pod.Status.Reason, pod.Status.Message)
			}
			logger.Printf("%s is not ready. Phase: %s, Message: %s, Reason: %s", pod.Name, pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
		}
		if namespace == "" || namespace != pod.Namespace {
			filteredPods[pod.Namespace] = map[string]interface{}{}
		}
		jobName = strings.LastIndex(pod.ObjectMeta.Labels["job-name"], "-")
		if jobName == -1 {
			return nil, fmt.Errorf("error parsing 'job-name' label of %s pod", pod.Name)
		}
		if cronJobName == "" || cronJobName != pod.ObjectMeta.Labels["job-name"][:jobName] || namespace != pod.Namespace {
			namespace = pod.Namespace
			cronJobName = pod.ObjectMeta.Labels["job-name"][:jobName]
			filteredPods[namespace][cronJobName] = &kubev1.PodList{}
		}
		if pod.ObjectMeta.Labels["job-name"][:jobName] == cronJobName && pod.Namespace == namespace {
			filteredPods[namespace][cronJobName].(*kubev1.PodList).Items = append(filteredPods[namespace][cronJobName].(*kubev1.PodList).Items, pod)
		}
	}

	// Iterate over the map of pods that was created above. If Phase of the last pod of the cronJob was Successful, do not return an error. Otherwise, return an error message.
	for namespace := range filteredPods {
		for cronJob := range filteredPods[namespace] {
			latestPodPhase := filteredPods[namespace][cronJob].(*kubev1.PodList).Items[len(filteredPods[namespace][cronJob].(*kubev1.PodList).Items)-1].Status.Phase
			for _, pod := range filteredPods[namespace][cronJob].(*kubev1.PodList).Items {
				if latestPodPhase == kubev1.PodSucceeded {
					break
				}
				if latestPodPhase != kubev1.PodPending {
					return nil, fmt.Errorf("pod %s in unexpected phase %s: reason: %s message: %s", pod.GetName(), pod.Status.Phase, pod.Status.Reason, pod.Status.Message)
				}
				logger.Printf("%s is not ready. Phase: %s, Message: %s, Reason: %s", pod.Name, pod.Status.Phase, pod.Status.Message, pod.Status.Reason)
			}
		}
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

// CheckPendingPods checks each pod in the provided list for pending state and updates
// the PodErrorTracker accordingly. It returns nil if no pods were pending more than
// their maximum threshold, and errors if a pod ever exceeds its maximum
// pending threshold.
func (p *PodErrorTracker) CheckPendingPods(podlist []kubev1.Pod) error {
	tempTracker := make(map[string]int)
	for _, pod := range podlist {
		if val, found := p.Counts[string(pod.UID)]; found {
			tempTracker[string(pod.UID)] = val + 1
		} else {
			tempTracker[string(pod.UID)] = 1
		}
		if tempTracker[string(pod.UID)] >= p.MaxPendingPodsThreshold {
			return fmt.Errorf("pod %s in namespace %s is pending beyond normal threshold: %s - %s", pod.GetName(), pod.GetNamespace(), pod.Status.Reason, pod.Status.Message)
		}
	}
	p.Counts = tempTracker

	if podlist != nil {
		return fmt.Errorf("pending pod key-value entries still present in the pending pod counter map")
	}
	return nil
}
