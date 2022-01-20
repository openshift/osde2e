package healthchecks

import (
	kubev1 "k8s.io/api/core/v1"
	"time"
)

type PodPredicate func(kubev1.Pod) bool

func IsClusterPod(pod kubev1.Pod) bool {
	return containsPrefixes(pod.Namespace, "openshift-", "kube-", "redhat-")
}

// IsNotReadinessPod ignores the chicken/egg situation where we're running health checks
// from within an ephemeral osd-cluster-ready-* Pod. That Pod would otherwise fail the
// health check it is running because it's in Pending state.
func IsNotReadinessPod(pod kubev1.Pod) bool {
	return !matchingNamePrefix(pod, "osd-cluster-ready-")
}

func MatchesNames(name ...string) PodPredicate {
	return func(p kubev1.Pod) bool {
		return matchingNamePrefix(p, name...)
	}
}

func MatchesNamespace(ns string) PodPredicate {
	return func(p kubev1.Pod) bool {
		return matchingNS(p, ns)
	}
}

func IsNotRunning(pod kubev1.Pod) bool {
	return pod.Status.Phase != kubev1.PodRunning
}

func IsNotCompleted(pod kubev1.Pod) bool {
	return pod.Status.Phase != kubev1.PodSucceeded
}

func matchingNamePrefix(pod kubev1.Pod, name ...string) bool {
	return containsPrefixes(pod.Name, name...)
}

func matchingNS(pod kubev1.Pod, ns string) bool {
	return pod.Namespace == ns
}

func IsNotControlledByJob(pod kubev1.Pod) bool {
	if len(pod.OwnerReferences) > 0 {
		return pod.OwnerReferences[0].Kind != "Job"
	}
	return true
}

func IsOlderThan(d time.Duration) PodPredicate {
	return func(p kubev1.Pod) bool {
		return olderThan(p, d)
	}
}

func olderThan(pod kubev1.Pod, d time.Duration) bool {
	// Let's not just assume that a caller gave a negative duration
	if d > 0 {
		d = -d
	}
	beforeTime := time.Now().Add(d)
	return pod.CreationTimestamp.Time.Before(beforeTime)
}