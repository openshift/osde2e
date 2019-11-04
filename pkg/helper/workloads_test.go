package helper

import (
	"testing"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamic "k8s.io/client-go/dynamic/fake"
	kubernetes "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestCreateWorkload(t *testing.T) {
	var tests = []struct {
		description string
		file        string
	}{
		{"pod", "/../../test/workloads/tests/pod.yaml"},
		{"pods", "/../../test/workloads/tests/pods.yaml"},
		{"service", "/../../test/workloads/tests/service.yaml"},
	}

	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset()

		obj, err := ReadK8sYaml(test.file)
		if err != nil {
			t.Errorf("%v: Expected a valid runtime.Object (%v)", test.description, err)
		}

		aScheme := scheme.Scheme
		aScheme.AddKnownTypes(kubev1.SchemeGroupVersion)
		dynamicClient := dynamic.NewSimpleDynamicClient(aScheme)

		newObj, err := CreateRuntimeObject(obj, dynamicClient, kubeClient.Discovery())
		if err != nil {
			t.Errorf("%v: Error creating K8s Object (%v)", test.description, err)
			continue
		}

		_, err = dynamicClient.Resource(gvrFromUnstructured(newObj)).Get(newObj.GetName(), metav1.GetOptions{})
		if err != nil {
			t.Errorf("%v: Error retrieving K8s Object (%v)", test.description, err)
		}

	}
}
