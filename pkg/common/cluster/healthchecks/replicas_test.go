package healthchecks

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	kubernetes "k8s.io/client-go/kubernetes/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func testDS(name, namespace string, readyReplicas, totalReplicas int32) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: totalReplicas,
			NumberReady:            readyReplicas,
		},
	}
}

func testRS(name, namespace string, readyReplicas, totalReplicas int32) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.ReplicaSetStatus{
			Replicas:      totalReplicas,
			ReadyReplicas: readyReplicas,
		},
	}
}

func TestCheckReplicaCountForDaemonSets(t *testing.T) {
	const (
		ns = "openshift-test-ns"
	)
	tests := []struct {
		description   string
		expected      bool
		expectedError bool
		objs          []runtime.Object
	}{
		{"no replicas", false, true, nil},
		{"one does not match", false, true, []runtime.Object{testDS("ds1", ns, 4, 3)}},
		{"one matches", true, false, []runtime.Object{testDS("ds1", ns, 6, 6)}},
		{"one of many does not match", false, true, []runtime.Object{testDS("ds1", ns, 4, 4), testDS("ds2", ns, 1, 5)}},
		{"all match", true, false, []runtime.Object{testDS("ds1", ns, 1, 1), testDS("ds2", ns, 2, 2), testDS("ds3", ns, 3, 3)}},
		{"one does not match in customer namespace", true, false, []runtime.Object{testDS("ds1", "default", 1, 3)}},
	}
	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset(test.objs...)
		state, err := CheckReplicaCountForDaemonSets(kubeClient.AppsV1(), nil)
		if err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error: %s", err)
			}
		} else {
			if test.expectedError {
				t.Error("Expected error")
			}
		}
		if state != test.expected {
			t.Errorf("%v: Expected value doesn't match returned value (%v, %v)", test.description, test.expected, state)
		}
	}
}

func TestCheckReplicaCountForReplicaSets(t *testing.T) {
	const (
		ns = "openshift-test-ns"
	)
	tests := []struct {
		description   string
		expected      bool
		expectedError bool
		objs          []runtime.Object
	}{
		{"no replicas", false, true, nil},
		{"one does not match", false, true, []runtime.Object{testRS("rs1", ns, 4, 3)}},
		{"one matches", true, false, []runtime.Object{testRS("rs1", ns, 6, 6)}},
		{"one of many does not match", false, true, []runtime.Object{testRS("rs1", ns, 4, 4), testRS("rs2", ns, 1, 5)}},
		{"all match", true, false, []runtime.Object{testRS("rs1", ns, 1, 1), testRS("rs2", ns, 2, 2), testRS("rs3", ns, 3, 3)}},
		{"one does not match in customer namespace", true, false, []runtime.Object{testRS("rs1", "default", 1, 3)}},
	}
	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset(test.objs...)
		state, err := CheckReplicaCountForReplicaSets(kubeClient.AppsV1(), nil)
		if err != nil {
			if !test.expectedError {
				t.Errorf("Unexpected error: %s", err)
			}
		} else {
			if test.expectedError {
				t.Error("Expected error")
			}
		}
		if state != test.expected {
			t.Errorf("%v: Expected value doesn't match returned value (%v, %v)", test.description, test.expected, state)
		}
	}
}
