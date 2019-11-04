package helper

import (
	"fmt"
	"io/ioutil"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
)

// ReadK8sYaml reads a file at the specified path and attempts to decode it into a runtime.Object
func ReadK8sYaml(file string) (runtime.Object, error) {
	pwd, _ := os.Getwd()
	f, err := ioutil.ReadFile(pwd + file)
	if err != nil {
		return nil, err
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(f), nil, nil)

	if err != nil {
		return nil, fmt.Errorf("Error while decoding YAML object. Err was: %s", err)
	}

	return obj, nil
}

// CreateRuntimeObject takes a runtime.Object and attempts to create an object in K8s with it
func CreateRuntimeObject(obj runtime.Object, dynamicClient dynamic.Interface, discovery discovery.DiscoveryInterface) (unstructured.Unstructured, error) {
	if obj == nil {
		return unstructured.Unstructured{}, fmt.Errorf("Nil object passed in")
	}

	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		fmt.Printf("Line 59: %v", err)
		return unstructured.Unstructured{}, err
	}
	thing := unstructured.Unstructured{unstructuredObj}

	newObj, err := dynamicClient.Resource(gvrFromObject(obj)).Create(&thing, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Line 65: %v", err)
		return unstructured.Unstructured{}, err
	}

	return *newObj, nil
}

func gvrFromObject(obj runtime.Object) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    obj.GetObjectKind().GroupVersionKind().Group,
		Version:  obj.GetObjectKind().GroupVersionKind().Version,
		Resource: obj.GetObjectKind().GroupVersionKind().Kind,
	}
}
func gvrFromUnstructured(obj unstructured.Unstructured) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    obj.GetObjectKind().GroupVersionKind().Group,
		Version:  obj.GetObjectKind().GroupVersionKind().Version,
		Resource: obj.GetObjectKind().GroupVersionKind().Kind,
	}
}