package healthchecks

import (
	"strconv"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetes "k8s.io/client-go/kubernetes/fake"
)

const (
	ns1 = "openshift-1"
	ns2 = "openshift-2"
)

func pod(name, namespace string, label map[string]string, phase v1.PodPhase) *v1.Pod {
	mockPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    label,
		},
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

	// If we set a job-name label, set the timestamp based on the job number to
	// simulate a cronJob and also populate OwnerReferences
	if val, ok := label["job-name"]; ok {
		jobNumber, err := strconv.Atoi(label["job-name"][strings.LastIndex(label["job-name"], "-")+1:])
		if err != nil {
			jobNumber = 0
		}
		mockPod.ObjectMeta.CreationTimestamp.Time = time.Unix(0, 0).Add(time.Duration(jobNumber) * time.Second)

		mockPod.OwnerReferences = append(mockPod.OwnerReferences, metav1.OwnerReference{
			APIVersion: "batch/v1",
			Kind:       "Job",
			Name:       val,
		})
	}

	return mockPod
}

func TestCheckPodHealth(t *testing.T) {
	var tests = []struct {
		description    string
		expectedLength int
		expectedError  bool
		objs           []runtime.Object
	}{
		{
			description:    "no pods",
			expectedLength: 0,
			expectedError:  true,
			objs:           nil,
		},
		{
			description:    "single pod failed",
			expectedLength: 0,
			expectedError:  true,
			objs: []runtime.Object{
				pod("a", ns1, map[string]string{}, v1.PodFailed),
			},
		},
		{
			description:    "single pod pending",
			expectedLength: 1,
			expectedError:  false,
			objs: []runtime.Object{
				pod("a", ns1, map[string]string{}, v1.PodPending),
			},
		},
		{
			description:    "healthy pods",
			expectedLength: 0,
			expectedError:  false,
			objs: []runtime.Object{
				pod("running", ns1, map[string]string{}, v1.PodRunning),
				pod("completed", ns2, map[string]string{}, v1.PodSucceeded),
				pod("failed-first-run", ns1, map[string]string{"job-name": "test-job-122"}, v1.PodFailed),
				pod("but-completed-second-run", ns1, map[string]string{"job-name": "test-job-123"}, v1.PodSucceeded),
				pod("worked-first-try", ns2, map[string]string{"job-name": "other-job-456"}, v1.PodSucceeded),
			},
		},
		{
			description:    "currently unhealthy cronjob pods",
			expectedLength: 0,
			expectedError:  true,
			objs: []runtime.Object{
				pod("complted-first-run", ns1, map[string]string{"job-name": "test-job-122"}, v1.PodSucceeded),
				pod("but-failed-second-run", ns1, map[string]string{"job-name": "test-job-124"}, v1.PodFailed),
				pod("and-failed-again-run", ns1, map[string]string{"job-name": "test-job-125"}, v1.PodFailed),
			},
		},
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

func TestCheckJobPods(t *testing.T) {
	tests := []struct {
		description         string
		expectError         bool
		expectedPendingPods int
		pods                []v1.Pod
	}{
		{
			description:         "one successful job pod",
			expectError:         false,
			expectedPendingPods: 0,
			pods: []v1.Pod{
				*pod("a", ns1, map[string]string{"job-name": "image-pruner-124"}, v1.PodSucceeded),
			},
		},
		{
			description:         "initially failed, but finally successful cronjob pods",
			expectError:         false,
			expectedPendingPods: 0,
			pods: []v1.Pod{
				*pod("a", ns1, map[string]string{"job-name": "image-pruner-124"}, v1.PodFailed),
				*pod("b", ns1, map[string]string{"job-name": "image-pruner-125"}, v1.PodFailed),
				*pod("c", ns1, map[string]string{"job-name": "image-pruner-126"}, v1.PodSucceeded),
			},
		},
		{
			description:         "completed, but then pending cronjob pods",
			expectError:         false,
			expectedPendingPods: 1,
			pods: []v1.Pod{
				*pod("a", ns1, map[string]string{"job-name": "image-pruner-124"}, v1.PodSucceeded),
				*pod("b", ns1, map[string]string{"job-name": "image-pruner-125"}, v1.PodPending),
			},
		},
		{
			description:         "not a job pod",
			expectError:         true,
			expectedPendingPods: 0,
			pods: []v1.Pod{
				*pod("a", ns1, map[string]string{}, v1.PodSucceeded),
			},
		},
	}

	for _, test := range tests {
		pendingPods, err := checkJobPods(test.pods, nil)

		if len(pendingPods) != test.expectedPendingPods {
			t.Errorf("%s: expected %v pending pods, got %v", test.description, test.expectedPendingPods, len(pendingPods))
		}

		if (err != nil && test.expectError == false) || (err == nil && test.expectError == true) {
			t.Errorf("%s: expected error %v, got %v", test.description, test.expectError, err)
		}
		t.Log(test.description)
	}
}
