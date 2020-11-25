package healthchecks

import kubev1 "k8s.io/api/core/v1"

type PodPredicate func(kubev1.Pod) bool

func IsClusterPod(pod kubev1.Pod) bool {
	return containsPrefixes(pod.Namespace, "openshift-", "kube-", "redhat-")
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

