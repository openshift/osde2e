package helper

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
	var tests = []struct {
		description   string
		expectedState bool
		expectedError bool
		objs          []runtime.Object
	}{
		{"no pods", false, false, nil},
		{"single pod failed", false, true, []runtime.Object{pod("a", "a", v1.PodFailed)}},
		{"one pod good one pod bad same namespace", false, true, []runtime.Object{pod("a", "a", v1.PodFailed), pod("b", "a", v1.PodRunning)}},
		{"one pod good one pod pending same namespace", false, false, []runtime.Object{pod("a", "a", v1.PodPending), pod("b", "a", v1.PodRunning)}},
		{"one pod good one pod bad diff namespace", false, true, []runtime.Object{pod("a", "a", v1.PodFailed), pod("b", "b", v1.PodRunning)}},
		{"single pod running", true, false, []runtime.Object{pod("a", "a", v1.PodRunning)}},
		{"single pod succeeded", true, false, []runtime.Object{pod("a", "a", v1.PodSucceeded)}},
		{"two succeeded pods diff namespace", true, false, []runtime.Object{pod("a", "a", v1.PodSucceeded), pod("b", "b", v1.PodSucceeded)}},
		{"two running pods diff namespace", true, false, []runtime.Object{pod("a", "a", v1.PodRunning), pod("b", "b", v1.PodRunning)}},
	}

	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset(test.objs...)
		state, err := CheckPodHealth(kubeClient.CoreV1())

		if state != test.expectedState {
			t.Errorf("%v: Expected state doesn't match returned value (%v, %v)", test.description, test.expectedState, state)
		}

		if (err != nil && test.expectedError == false) || (err == nil && test.expectedError == true) {
			t.Errorf("%v: Expected error doesn't match returned value (%v, %v)", test.description, test.expectedError, err)
		}
	}
}
