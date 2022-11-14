package helper

import (
	"context"
	"fmt"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"

	routev1 "github.com/openshift/api/route/v1"
	userv1 "github.com/openshift/api/user/v1"
	routemonitorv1alpha1 "github.com/openshift/route-monitor-operator/api/v1alpha1"
	kv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RouteMonitorResource(h *H) dynamic.NamespaceableResourceInterface {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "monitoring.openshift.io", Version: "v1alpha1", Resource: "routemonitors",
	})
}

func GetRouteMonitor(ctx context.Context, name string, ns string, h *H) (*routemonitorv1alpha1.RouteMonitor, error) {
	ucObj, err := RouteMonitorResource(h).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error retrieving RouteMonitor: %w", err)
	}
	var routeMonitor routemonitorv1alpha1.RouteMonitor
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ucObj.UnstructuredContent(), &routeMonitor)
	if err != nil {
		// This, however, is probably error-worthy because it means our RouteMonitor
		// has been messed with or something odd's occurred
		return nil, fmt.Errorf("error parsing RouteMonitor into object: %w", err)
	}

	return &routeMonitor, nil
}

func UpdateRouteMonitor(ctx context.Context, routeMonitor *routemonitorv1alpha1.RouteMonitor, namespace string, h *H) error {
	rawObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(routeMonitor.DeepCopy())
	obj := &unstructured.Unstructured{rawObj} // warning about unjeyed fields here
	if err != nil {
		return fmt.Errorf("can't convert RouteMonitor to unstructured resource: %w", err)
	}
	_, err = RouteMonitorResource(h).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Could not issue create command")
		return err
	}
	return nil
}

func CreateRouteMonitor(ctx context.Context, routeMonitor *routemonitorv1alpha1.RouteMonitor, namespace string, h *H) error {
	rawObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(routeMonitor.DeepCopy())
	if err != nil {
		return fmt.Errorf("can't convert UpgradeConfig to unstructured resource: %w", err)
	}
	obj := &unstructured.Unstructured{rawObj} // warning about unjeyed fields here

	newObj, err := RouteMonitorResource(h).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("can't create RouteMonitor, returned obj %v: %w", newObj, err)
	}

	// Wait for the pod to create.
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		if _, err := RouteMonitorResource(h).Namespace(namespace).Get(ctx, routeMonitor.Name, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})
	return err
}

func DeleteRouteMonitor(ctx context.Context, nsName types.NamespacedName, waitForDelete bool, h *H) error {
	namespace, name := nsName.Namespace, nsName.Name
	err := RouteMonitorResource(h).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace '%s': %w", namespace, err)
	}

	// Deleting a namespace can take a while. If desired, wait for the namespace to delete before returning.
	if waitForDelete {
		err = wait.PollImmediate(2*time.Second, 1*time.Minute, func() (bool, error) {
			rmo, err := RouteMonitorResource(h).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
			// not sure on that
			if rmo != nil && err == nil {
				return false, nil
			}
			return true, nil
		})
	}

	return err
}

func SampleRouteMonitor(name string, ns string, h *H) *routemonitorv1alpha1.RouteMonitor {
	routeMonitor := routemonitorv1alpha1.RouteMonitor{
		// This is required for ToUnstructured to work correctly (test and see output)
		TypeMeta: metav1.TypeMeta{
			Kind:       "RouteMonitor",
			APIVersion: "monitoring.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: routemonitorv1alpha1.RouteMonitorSpec{
			Slo: routemonitorv1alpha1.SloSpec{
				TargetAvailabilityPercent: "99.95",
			},
			Route: routemonitorv1alpha1.RouteMonitorRouteSpec{
				Namespace: ns,
				Name:      name,
			},
		},
	}
	return &routeMonitor
}

func ClusterRouteMonitorResource(h *H) dynamic.NamespaceableResourceInterface {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "monitoring.openshift.io", Version: "v1alpha1", Resource: "clusterurlmonitors",
	})
}

func CreateRoute(ctx context.Context, route *routev1.Route, namespace string, h *H) error {
	uwm, err := h.Route().RouteV1().Routes(namespace).Create(ctx, route, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Could not issue create command")
		return err
	}

	// Wait for the pod to create.
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		if _, err := h.Route().RouteV1().Routes(namespace).Get(ctx, uwm.Name, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})
	return err
}

func SampleRoute(name, ns string) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Name: name,
			},
			TLS: &routev1.TLSConfig{Termination: "edge"},
		},
		Status: routev1.RouteStatus{},
	}
}

func CreatePod(ctx context.Context, pod *kv1.Pod, namespace string, h *H) error {
	uwm, err := h.Kube().CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Could not issue create command")
		return err
	}

	// Wait for the pod to create.
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		if _, err := h.Kube().CoreV1().Pods(namespace).Get(ctx, uwm.Name, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})
	return err
}

func SamplePod(name, namespace, imageName string) *kv1.Pod {
	return &kv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kv1.PodSpec{
			Containers: []kv1.Container{
				{
					Name:  "test",
					Image: imageName,
				},
			},
		},
	}
}

func SampleService(port int32, targetPort int, serviceName, serviceNamespace string, prometheusName string) *kv1.Service {
	service := &kv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      prometheusName,
			Namespace: serviceNamespace,
			Labels:    map[string]string{prometheusName: serviceName},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		Spec: kv1.ServiceSpec{
			Ports: []kv1.ServicePort{
				{
					Port:       port,
					Protocol:   kv1.ProtocolTCP,
					TargetPort: intstr.FromInt(targetPort),
					Name:       "web",
				},
			},
			Selector: map[string]string{prometheusName: serviceName},
		},
	}
	return service
}

func CreateService(ctx context.Context, svc *kv1.Service, h *H) error {
	uwm, err := h.Kube().CoreV1().Services(svc.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Could not issue create command")
		return err
	}

	// Wait for the pod to create.
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		if _, err := h.Kube().CoreV1().Services(uwm.Namespace).Get(ctx, uwm.Name, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})
	return err
}

func CreateNamespace(ctx context.Context, namespace string, h *H) (*kv1.Namespace, error) {
	// If the namespace already exists, we don't need to create it. Just return.
	ns, err := h.Kube().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if ns != nil && ns.Status.Phase != "Terminating" && err == nil {
		return ns, err
	}

	log.Printf("Creating namespace for namespace validation webhook (%s)", namespace)
	ns = &kv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	h.Kube().CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})

	// Wait for the namespace to create. This is usually pretty quick.
	err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		if _, err := h.Kube().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})

	return ns, err
}

func DeleteNamespace(ctx context.Context, namespace string, waitForDelete bool, h *H) error {
	log.Printf("Deleting namespace (%s)", namespace)
	err := h.Kube().CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace '%s': %w", namespace, err)
	}

	// Deleting a namespace can take a while. If desired, wait for the namespace to delete before returning.
	if waitForDelete {
		err = wait.PollImmediate(2*time.Second, 1*time.Minute, func() (bool, error) {
			ns, _ := h.Kube().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
			if ns != nil && ns.Status.Phase == "Terminating" {
				return false, nil
			}
			return true, nil
		})
	}

	return err
}

func CreateUser(ctx context.Context, userName string, identities []string, groups []string, h *H) (*userv1.User, error) {
	user := &userv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName,
		},
		Identities: identities,
		Groups:     groups,
	}
	return h.User().UserV1().Users().Create(ctx, user, metav1.CreateOptions{})
}

func AddUserToGroup(ctx context.Context, userName string, groupName string, h *H) (result *userv1.Group, err error) {
	group, err := h.User().UserV1().Groups().Get(ctx, groupName, metav1.GetOptions{})
	if err != nil {
		return &userv1.Group{}, err
	}

	group.Users = append(group.Users, userName)
	return h.User().UserV1().Groups().Update(ctx, group, metav1.UpdateOptions{})
}
