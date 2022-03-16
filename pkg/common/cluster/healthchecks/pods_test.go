package healthchecks

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetes "k8s.io/client-go/kubernetes/fake"
)

func pod(name, namespace string, label map[string]string, phase v1.PodPhase) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: label},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "scratch",
				},
			},
		},
		Status: v1.PodStatus{
			Phase:   phase,
			Message: "pod message",
			Reason:  "pod reason",
		},
	}
}
func TestCheckPodHealth(t *testing.T) {
	const (
		ns1 = "openshift-1"
		ns2 = "openshift-2"
	)
	var tests = []struct {
		description    string
		expectedLength int
		expectedError  bool
		objs           []runtime.Object
	}{
		{"two pods bad, one pod good with same label in the same namespace", 0, false, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodFailed), pod("b", ns1, map[string]string{"job-name": "image-pruner-124"}, v1.PodFailed), pod("c", ns1, map[string]string{"job-name": "image-pruner-125"}, v1.PodSucceeded)}},
		{"one pod bad, one pod good with same label in the same namespace. One pod bad with different label and namespace", 0, true, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodFailed), pod("b", ns1, map[string]string{"job-name": "image-pruner-124"}, v1.PodSucceeded), pod("c", ns1, map[string]string{"job-name": "new-image-pruner-124"}, v1.PodFailed)}},
		{"no pods", 0, true, nil},
		{"single pod failed", 0, true, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodFailed)}},
		{"single pod without job-name label failed", 0, true, []runtime.Object{pod("a", ns1, map[string]string{}, v1.PodFailed)}},
		{"single pod without job-name label succeeded", 0, false, []runtime.Object{pod("a", ns1, map[string]string{}, v1.PodSucceeded)}},
		{"pod is pending beyond specified threshold", 1, false, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodPending)}},
		{"single pod running", 0, false, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodRunning)}},
		{"single pod succeeded", 0, false, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodSucceeded)}},
		{"single pod failed bad namespace", 0, false, []runtime.Object{pod("a", "foobar", map[string]string{"job-name": "image-pruner-123"}, v1.PodFailed)}},
		{"one pod good one pod bad same namespace", 0, true, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodFailed), pod("b", ns1, map[string]string{"job-name": "image-pruner-124"}, v1.PodRunning)}},
		{"one pod good one pod pending same namespace", 0, false, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodPending), pod("b", ns1, map[string]string{"job-name": "image-pruner-124"}, v1.PodRunning)}},
		{"one pod good one pod bad diff namespace", 0, true, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodFailed), pod("b", ns2, map[string]string{"job-name": "image-pruner-124"}, v1.PodRunning)}},
		{"two succeeded pods diff namespace", 0, false, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodSucceeded), pod("b", ns2, map[string]string{"job-name": "image-pruner-124"}, v1.PodSucceeded)}},
		{"one pod good, one pod bad diff namespace", 0, true, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodSucceeded), pod("b", ns2, map[string]string{"job-name": "image-pruner-124"}, v1.PodFailed)}},
		{"two running pods diff namespace", 0, false, []runtime.Object{pod("a", ns1, map[string]string{"job-name": "image-pruner-123"}, v1.PodRunning), pod("b", ns2, map[string]string{"job-name": "image-pruner-124"}, v1.PodRunning)}},
	}

	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset(test.objs...)
		state, err := CheckClusterPodHealth(kubeClient.CoreV1(), nil)

		// Length of the pending pods list is validated here. The list may have multiple pending pods even if the error is for one pending pod.
		if len(state) < test.expectedLength {
			t.Errorf("%v: Expected length of state doesn't match returned value (%v, %v)", test.description, test.expectedLength, len(state))
		}

		if (err != nil && test.expectedError == false) || (err == nil && test.expectedError == true) {
			t.Errorf("%v: Expected error doesn't match returned value (%v, %v)", test.description, test.expectedError, err)
		}
	}
}
