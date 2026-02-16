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
)

func pod(name string, label map[string]string, phase v1.PodPhase) *v1.Pod {
	mockPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns1,
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
		mockPod.CreationTimestamp.Time = time.Unix(0, 0).Add(time.Duration(jobNumber) * time.Second)

		mockPod.OwnerReferences = append(mockPod.OwnerReferences, metav1.OwnerReference{
			APIVersion: "batch/v1",
			Kind:       "Job",
			Name:       val,
		})
	}

	return mockPod
}

func TestCheckPodHealth(t *testing.T) {
	tests := []struct {
		description   string
		isHealthy     bool
		expectedError bool
		objs          []runtime.Object
	}{
		{
			description:   "no pods",
			isHealthy:     false,
			expectedError: true,
			objs:          nil,
		},
		{
			description:   "single pod failed",
			isHealthy:     false,
			expectedError: true,
			objs: []runtime.Object{
				pod("a", map[string]string{}, v1.PodFailed),
			},
		},
		{
			description:   "healthy pods",
			isHealthy:     true,
			expectedError: false,
			objs: []runtime.Object{
				pod("running", map[string]string{}, v1.PodRunning),
				pod("completed", map[string]string{}, v1.PodSucceeded),
				pod("long-job-name", map[string]string{"job-name": "thisisalongjobnamethatisawhoppingsixtythreecharacterslongtotest"}, v1.PodSucceeded),
				pod("failed-first-run", map[string]string{"job-name": "test-job-122"}, v1.PodFailed),
				pod("but-completed-second-run", map[string]string{"job-name": "test-job-123"}, v1.PodSucceeded),
				pod("worked-first-try", map[string]string{"job-name": "other-job-456"}, v1.PodSucceeded),
			},
		},
		{
			description:   "multiple unhealthy pods",
			isHealthy:     false,
			expectedError: true,
			objs: []runtime.Object{
				pod("running-first-pod", map[string]string{}, v1.PodRunning),
				pod("but-failed-second-pod", map[string]string{}, v1.PodFailed),
				pod("and-failed-again-pod", map[string]string{}, v1.PodFailed),
			},
		},
	}

	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset(test.objs...)
		state, err := CheckPodHealth(kubeClient.CoreV1(), nil, ns1, "")

		// Length of the pending pods list is validated here. The list may have multiple pending pods even if the error is for one pending pod.
		if state != test.isHealthy {
			t.Errorf("%v: Expected health state of pods doesn't match returned value (%v, %v)", test.description, test.isHealthy, state)
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
				*pod("a", map[string]string{"job-name": "image-pruner-124"}, v1.PodSucceeded),
			},
		},
		{
			description:         "initially failed, but finally successful cronjob pods",
			expectError:         false,
			expectedPendingPods: 0,
			pods: []v1.Pod{
				*pod("a", map[string]string{"job-name": "image-pruner-124"}, v1.PodFailed),
				*pod("b", map[string]string{"job-name": "image-pruner-125"}, v1.PodFailed),
				*pod("c", map[string]string{"job-name": "image-pruner-126"}, v1.PodSucceeded),
			},
		},
		{
			description:         "completed, but then pending cronjob pods",
			expectError:         false,
			expectedPendingPods: 1,
			pods: []v1.Pod{
				*pod("a", map[string]string{"job-name": "image-pruner-124"}, v1.PodSucceeded),
				*pod("b", map[string]string{"job-name": "image-pruner-125"}, v1.PodPending),
			},
		},
		{
			description:         "not a job pod",
			expectError:         true,
			expectedPendingPods: 0,
			pods: []v1.Pod{
				*pod("a", map[string]string{}, v1.PodSucceeded),
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
