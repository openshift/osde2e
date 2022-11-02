package helper

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/openshift/osde2e/assets"
)

// ApplyYamlInFolder reads a folder and attempts to create objects in K8s with the yaml
func ApplyYamlInFolder(folder, namespace string, kube kubernetes.Interface) ([]runtime.Object, error) {
	var (
		objects []runtime.Object
		obj     runtime.Object
		files   []string
		err     error
	)
	err = fs.WalkDir(assets.FS, folder, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info != nil && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return objects, err
	}

	for _, file := range files {
		if obj, err = ReadK8sYaml(file); err != nil {
			return objects, err
		}
		if obj, err = CreateRuntimeObject(obj, namespace, kube); err != nil {
			return objects, err
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// ReadK8sYaml reads a file at the specified path and attempts to decode it into a runtime.Object
func ReadK8sYaml(file string) (runtime.Object, error) {
	var (
		fileReader fs.File
		err        error
	)

	if fileReader, err = assets.FS.Open(file); err != nil {
		return nil, err
	}

	f, err := ioutil.ReadAll(fileReader)
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
func CreateRuntimeObject(obj runtime.Object, ns string, kube kubernetes.Interface) (runtime.Object, error) {
	var (
		newObj runtime.Object
		ok     bool
		err    error
	)
	if obj == nil {
		return nil, fmt.Errorf("Nil object passed in")
	}

	if _, err := kube.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{}); err != nil {
		_, err := kube.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "namespace",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
			Spec: corev1.NamespaceSpec{},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Error creating namespace: %s", err.Error())
		}

		// Need to wait till the namespace is actually created/active before proceeding
		// Namespace creation is fairly swift, so these numbers are arbitrary.
		// If this times out, there's something wrong.
		wait.PollImmediate(2*time.Second, 1*time.Minute, func() (bool, error) {
			if _, err := kube.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{}); err != nil {
				return false, nil
			}
			return true, nil
		})
	}

	// TODO: As new object types need to be created, add support for them here
	switch obj.GetObjectKind().GroupVersionKind().Kind {
	case "Pod":
		if _, ok = obj.(*corev1.Pod); !ok {
			return nil, fmt.Errorf("Error casting object to pod")
		}

		if newObj, err = kube.CoreV1().Pods(ns).Create(context.TODO(), obj.(*corev1.Pod), metav1.CreateOptions{}); err != nil {
			return nil, err
		}
		return newObj, nil
	case "Service":
		if _, ok = obj.(*corev1.Service); !ok {
			return nil, fmt.Errorf("Error casting object to service")
		}
		if newObj, err = kube.CoreV1().Services(ns).Create(context.TODO(), obj.(*corev1.Service), metav1.CreateOptions{}); err != nil {
			return nil, err
		}
		return newObj, nil
	case "Deployment":
		dep := &appsv1.Deployment{}
		obj.(*appsv1.Deployment).DeepCopyInto(dep)

		if newObj, err = kube.AppsV1().Deployments(ns).Create(context.TODO(), dep, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
		return newObj, nil
	case "StatefulSet":
		ss := &appsv1.StatefulSet{}
		obj.(*appsv1.StatefulSet).DeepCopyInto(ss)

		if newObj, err = kube.AppsV1().StatefulSets(ns).Create(context.TODO(), ss, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
		return newObj, nil
	case "PersistentVolume":
		if _, ok = obj.(*corev1.PersistentVolume); !ok {
			return nil, fmt.Errorf("Error casting object to PersistentVolume")
		}

		if newObj, err = kube.CoreV1().PersistentVolumes().Create(context.TODO(), obj.(*corev1.PersistentVolume), metav1.CreateOptions{}); err != nil {
			return nil, err
		}
		return newObj, nil
	case "PersistentVolumeClaim":
		if _, ok = obj.(*corev1.PersistentVolumeClaim); !ok {
			return nil, fmt.Errorf("Error casting object to PersistentVolumeClaim")
		}

		if newObj, err = kube.CoreV1().PersistentVolumeClaims(ns).Create(context.TODO(), obj.(*corev1.PersistentVolumeClaim), metav1.CreateOptions{}); err != nil {
			return nil, err
		}
		return newObj, nil
	case "Secret":
		if _, ok = obj.(*corev1.Secret); !ok {
			return nil, fmt.Errorf("Error casting object to Secret")
		}

		if newObj, err = kube.CoreV1().Secrets(ns).Create(context.TODO(), obj.(*corev1.Secret), metav1.CreateOptions{}); err != nil {
			return nil, err
		}
		return newObj, nil
	case "PodDisruptionBudget":
		if _, ok = obj.(*policyv1beta1.PodDisruptionBudget); !ok {
			return nil, fmt.Errorf("Error casting object to PodDisruptionBudget")
		}
		if newObj, err = kube.PolicyV1beta1().PodDisruptionBudgets(ns).Create(context.TODO(), obj.(*policyv1beta1.PodDisruptionBudget), metav1.CreateOptions{}); err != nil {
			return nil, err
		}
		return newObj, nil

	default:
		return nil, fmt.Errorf("Unable to handle object type %s", obj.GetObjectKind().GroupVersionKind().Kind)
	}
}
