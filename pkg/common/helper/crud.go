package helper

import (
	"context"
	"fmt"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	userv1 "github.com/openshift/api/user/v1"
	kv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreatePod(ctx context.Context, pod *kv1.Pod, namespace string, h *H) error {
	uwm, err := h.Kube().CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Could not issue create command")
		return err
	}

	// Wait for the pod to create.
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 1*time.Minute, false, func(ctx context.Context) (bool, error) {
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
	err := h.Client.WithNamespace(svc.Namespace).Create(ctx, svc)
	if err != nil {
		log.Printf("Could not issue create command")
		return err
	}

	// Wait for the pod to create.
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 1*time.Minute, false, func(ctx context.Context) (bool, error) {
		tmpSvc := &kv1.Service{}
		if err := h.Client.Get(ctx, svc.Name, svc.Namespace, tmpSvc); err != nil {
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
	ns, err = h.Kube().CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// Wait for the namespace to create. This is usually pretty quick.
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (bool, error) {
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
		err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 1*time.Minute, false, func(ctx context.Context) (bool, error) {
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
