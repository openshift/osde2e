package healthchecks

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetes "k8s.io/client-go/kubernetes/fake"
)

func pod(name, namespace string, phase v1.PodPhase) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
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
		{"no pods", 0, true, nil},
		{"single pod failed", 0, true, []runtime.Object{pod("a", ns1, v1.PodFailed)}},
		{"pod is pending beyond specified threshold", 1, false, []runtime.Object{pod("a", ns1, v1.PodPending)}},
		{"single pod running", 0, false, []runtime.Object{pod("a", ns1, v1.PodRunning)}},
		{"single pod succeeded", 0, false, []runtime.Object{pod("a", ns1, v1.PodSucceeded)}},
		{"single pod failed bad namespace", 0, false, []runtime.Object{pod("a", "foobar", v1.PodFailed)}},
		{"one pod good one pod bad same namespace", 0, true, []runtime.Object{pod("a", ns1, v1.PodFailed), pod("b", ns1, v1.PodRunning)}},
		{"one pod good one pod pending same namespace", 0, false, []runtime.Object{pod("a", ns1, v1.PodPending), pod("b", ns1, v1.PodRunning)}},
		{"one pod good one pod bad diff namespace", 0, true, []runtime.Object{pod("a", ns1, v1.PodFailed), pod("b", ns2, v1.PodRunning)}},
		{"two succeeded pods diff namespace", 0, false, []runtime.Object{pod("a", ns1, v1.PodSucceeded), pod("b", ns2, v1.PodSucceeded)}},
		{"two running pods diff namespace", 0, false, []runtime.Object{pod("a", ns1, v1.PodRunning), pod("b", ns2, v1.PodRunning)}},
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
