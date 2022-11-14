package healthchecks

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetes "k8s.io/client-go/kubernetes/fake"
)

func node(name string, conditions []v1.NodeCondition) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       v1.NodeSpec{}, Status: v1.NodeStatus{
			Conditions: conditions,
		},
	}
}

func TestCheckNodeHealth(t *testing.T) {
	tests := []struct {
		description   string
		expected      bool
		expectedError bool
		objs          []runtime.Object
	}{
		{"no nodes", false, true, nil},
		{"node ready false", false, false, []runtime.Object{
			node("node-ready-false", []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: "False",
				},
			}),
		}},
		{"node ready unknown", false, false, []runtime.Object{
			node("node-ready-unknown", []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: "Unknown",
				},
			}),
		}},
		{"node ready true", true, false, []runtime.Object{
			node("node-ready-true", []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
			}),
		}},
		{"out-of-disk", false, false, []runtime.Object{
			node("correct-namespace", []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
				{
					Type:   "OutOfDisk",
					Status: "True",
				},
			}),
		}},
	}

	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset(test.objs...)
		state, err := CheckNodeHealth(kubeClient.CoreV1(), nil)

		if err != nil && !test.expectedError {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		if state != test.expected {
			t.Errorf("%v: Expected value doesn't match returned value (%v, %v)", test.description, test.expected, state)
		}
	}
}
