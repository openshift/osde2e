package healthchecks

import (
	"context"
	"fmt"
	"log"
	"sort"
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

	list, err := podClient.Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting pod list: %v", err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("pod list is empty. this should NOT happen")
	}

	pods := filterPods(list, filters...)

	// Keep track of all pending pods that are not associated with a job
	// and store all pods associated with a job for further analysis
	pendingPods := []kubev1.Pod{}
	jobPods := []kubev1.Pod{}
	for _, pod := range pods.Items {
		if IsNotControlledByJob(pod) {
			// Completed pod not associated with a job, e.g. a standalone pod
			if pod.Status.Phase == kubev1.PodSucceeded {
				continue
			}

			if pod.Status.Phase != kubev1.PodPending {
				return nil, fmt.Errorf("pod %s/%s in unexpected phase %s: reason: %s message: %s", pod.Namespace, pod.Name, pod.Status.Phase, pod.Status.Reason, pod.Status.Message)
			}
			logger.Printf("pod %s/%s is not ready. Phase: %s, Reason: %s, Message: %s", pod.Namespace, pod.Name, pod.Status.Phase, pod.Status.Reason, pod.Status.Message)
			pendingPods = append(pendingPods, pod)
		} else {
			jobPods = append(jobPods, pod)
		}
	}

	pendingJobPods, err := checkJobPods(jobPods, logger)
	if err != nil {
		return nil, err
	}
	logger.Printf("%v pods are currently not running or complete:", len(pendingPods)+len(pendingJobPods))

	return append(pendingPods, pendingJobPods...), nil
}

func checkJobPods(pods []kubev1.Pod, logger *log.Logger) ([]kubev1.Pod, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	filteredPods := make(map[string]map[string][]kubev1.Pod)
	pendingPods := []kubev1.Pod{}

	for _, pod := range pods {
		if _, ok := filteredPods[pod.Namespace]; !ok {
			filteredPods[pod.Namespace] = make(map[string][]kubev1.Pod)
		}

		jobOrCronJobName, ok := pod.ObjectMeta.Labels["job-name"]
		if !ok || len(jobOrCronJobName) == 0 {
			return nil, fmt.Errorf("expected 'job-name' label to be non-empty for pod %s/%s", pod.Namespace, pod.Name)
		}

		// The maximum length of a CronJob name is 52 characters, resulting in a job-name of
		// cronJobName-timestamp, where timestamp is a numeric string.
		// If it's a standalone job instead, it can be up to 63 characters
		if len(jobOrCronJobName) <= 52 {
			jobNameCutoff := strings.LastIndex(pod.ObjectMeta.Labels["job-name"], "-")
			if jobNameCutoff != -1 {
				jobOrCronJobName = pod.ObjectMeta.Labels["job-name"][:jobNameCutoff]
			}
		}

		if _, ok := filteredPods[pod.Namespace][jobOrCronJobName]; !ok {
			filteredPods[pod.Namespace][jobOrCronJobName] = []kubev1.Pod{}
		}
		filteredPods[pod.Namespace][jobOrCronJobName] = append(filteredPods[pod.Namespace][jobOrCronJobName], pod)
	}

	for namespace := range filteredPods {
		for cronJob := range filteredPods[namespace] {
			// Sort the cronJob pods by creationTimestamp, most recent will be last
			sort.SliceStable(filteredPods[namespace][cronJob], func(i, j int) bool {
				first := filteredPods[namespace][cronJob][i].GetObjectMeta().GetCreationTimestamp()
				second := filteredPods[namespace][cronJob][j].GetObjectMeta().GetCreationTimestamp()
				return first.Before(&second)
			})
			latestPod := filteredPods[namespace][cronJob][len(filteredPods[namespace][cronJob])-1]
			if latestPod.Status.Phase == kubev1.PodSucceeded {
				continue
			}
			if latestPod.Status.Phase != kubev1.PodPending {
				return nil, fmt.Errorf("pod %s/%s in unexpected phase %s", latestPod.Namespace, latestPod.Name, latestPod.Status.Phase)
			}

			for _, pod := range filteredPods[namespace][cronJob] {
				logger.Printf("pod %s/%s is not ready. Phase: %s", pod.Namespace, pod.Name, pod.Status.Phase)
				if pod.Status.Phase != kubev1.PodSucceeded {
					pendingPods = append(pendingPods, pod)
				}
			}
		}
	}

	return pendingPods, nil
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
		match := true
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
		tempTracker[string(pod.UID)]++
		if tempTracker[string(pod.UID)] >= p.MaxPendingPodsThreshold {
			return fmt.Errorf("pod %s/%s is pending beyond normal threshold: %s - %s", pod.Namespace, pod.Name, pod.Status.Reason, pod.Status.Message)
		}
	}
	p.Counts = tempTracker

	return nil
}
