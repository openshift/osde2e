package helper

import (
	"testing"

	kubernetes "k8s.io/client-go/kubernetes/fake"
)

func TestCreateWorkload(t *testing.T) {
	var tests = []struct {
		description string
		file        string
	}{
		{"pod", "workloads/tests/pod.yaml"},
		{"pods", "workloads/tests/pods.yaml"},
		{"service", "workloads/tests/service.yaml"},
	}

	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset()

		obj, err := ReadK8sYaml(test.file)
		if err != nil {
			t.Errorf("%v: Expected a valid runtime.Object (%v)", test.description, err)
		}
		_, err = CreateRuntimeObject(obj, "test", kubeClient)
		if err != nil {
			t.Errorf("%v: Error creating K8s Object (%v)", test.description, err)
			continue
		}

	}
}
