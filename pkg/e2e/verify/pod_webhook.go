package verify

import (
	"context"
	"log"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

var podWebhookTestName string = "[Suite: e2e] [OSD] pod validating webhook"

func init() {
	alert.RegisterGinkgoAlert(podWebhookTestName, "SD-SREP", "Alice Hubenko", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(podWebhookTestName, func() {

	h := helper.New()

	ginkgo.Context("pod webhook", func() {

		// for all tests, "manage" is synonymous with "create/update/delete"
		//Dedicated admin can not deploy pod on master on infra nodes in namespaces
		//openshift-operators, openshift-logging namespace or any other namespace that is not a core namespace like openshift-*, redhat-*, default, kube-*.

		ginkgo.It("Webhook will mark pod spec invalid and block deploying", func() {
			name := "osde2e-pod-webhook-test1"
			namespace := "a-namespace"
			_, err := createNamespace(namespace, h)
			defer deleteNamespace(namespace, true, h)
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			_, err = createPod(name, namespace, "node-role.kubernetes.io/master", "toleration-key-value", v1.TaintEffectNoSchedule, "node-role.kubernetes.io/infra", "toleration-key-value2", v1.TaintEffectNoSchedule, h)
			defer deletePod(name, namespace, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})

	})

	ginkgo.Context("pod webhook", func() {
		ginkgo.It("Webhook will mark pod spec invalid and block deploying", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			name := "osde2e-pod-webhook-test2"
			namespace := "openshift-logging"
			defer deletePod(name, namespace, h)
			_, err := createPod(name, namespace, "node-role.kubernetes.io/infra", "toleration-key-value", v1.TaintEffectPreferNoSchedule, "node-role.kubernetes.io/master", "toleration-key-value2", v1.TaintEffectNoExecute, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})
	})

	ginkgo.Context("pod webhook", func() {
		ginkgo.It("Webhook will allow pod to deploy", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			name := "osde2e-pod-webhook-test3"
			namespace := "openshift-apiserver"
			defer deletePod(name, namespace, h)
			_, err := createPod(name, namespace, "node-role.kubernetes.io/infra", "toleration-key-value", v1.TaintEffectNoSchedule, "node-role.kubernetes.io/master", "toleration-key-value2", v1.TaintEffectNoSchedule, h)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	// RBAC will prevent ordinary users from creating pods
	ginkgo.Context("pod webhook", func() {
		ginkgo.It("RBAC will deny deploying pod", func() {
			name := "osde2e-pod-webhook-test4"
			namespace := "openshift-logging"
			user := "alice"
			userGroup := "test-users"
			err := asUser(namespace, user, userGroup, h)
			if err != nil {
				log.Printf("Could not impersonate user, Error %v", err)
				return
			}
			defer h.Impersonate(rest.ImpersonationConfig{})
			_, err = createPod(name, namespace, "node-role.kubernetes.io/master", "toleration-key-value", v1.TaintEffectNoSchedule, "node-role.kubernetes.io/infra", "toleration-key-value2", v1.TaintEffectNoSchedule, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})

	ginkgo.Context("pod webhook", func() {
		ginkgo.It("RBAC will deny deploying pod", func() {
			name := "osde2e-pod-webhook-test4"
			namespace := "random-namespace"
			user := "alice"
			userGroup := "test-users"
			_, err := createNamespace(namespace, h)
			defer deleteNamespace(namespace, true, h)
			err = asUser(namespace, user, userGroup, h)
			if err != nil {
				log.Printf("Could not impersonate user, Error %v", err)
				return
			}
			defer h.Impersonate(rest.ImpersonationConfig{})
			_, err = createPod(name, namespace, "node-role.kubernetes.io/master", "toleration-key-value", v1.TaintEffectNoSchedule, "node-role.kubernetes.io/infra", "toleration-key-value2", v1.TaintEffectNoSchedule, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})

})

func makePod(name string, namespace string, key string, value string, effect v1.TaintEffect, key1 string, value1 string, effect1 v1.TaintEffect) *v1.Pod {

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "test",
					Image: "registry.access.redhat.com/ubi8/ubi-minimal",
				},
			},
			Tolerations: []v1.Toleration{
				{
					Key:      key,
					Operator: v1.TolerationOpEqual,
					Value:    value,
					Effect:   effect,
				},
				{
					Key:      key1,
					Operator: v1.TolerationOpEqual,
					Value:    value1,
					Effect:   effect1,
				},
			},
		},
	}
	return pod
}

func deletePod(name string, namespace string, h *helper.H) error {
	_, err := h.Kube().CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	log.Printf("Check before deleting pod %s, error: %v", name, err)
	if err == nil {
		err = h.Kube().CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		log.Printf("Deleting pod %s, error: %v", name, err)

		// Wait for the pod to delete.
		err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
			if _, err := h.Kube().CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
				return true, nil
			}
			return false, nil
		})

	}
	return err
}

func createPod(name string, namespace string, key string, value string, effect v1.TaintEffect, key1 string, value1 string, effect1 v1.TaintEffect, h *helper.H) (*v1.Pod, error) {

	pod := makePod(name, namespace, key, value, effect, key1, value1, effect1)

	//If pod is already created we delete the pod.
	pd, err := h.Kube().CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if pd != nil && err == nil {
		log.Printf("Pod %s already exists in namespace %s", name, namespace)
		err = deletePod(name, namespace, h)
		return pd, err
	}

	log.Printf("Creating pod for the validation webhook (%s)", pod)
	pd, err = h.Kube().CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	log.Printf("Result of the create command: (%v)", err)
	if err != nil {
		log.Printf("Could not issue create command")
		return pd, err
	}

	// Wait for the pod to create.
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		if _, err := h.Kube().CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})

	return pd, err
}

func asUser(namespace string, user string, userGroup string, h *helper.H) (err error) {
	// reset impersonation at the beginning just-in-case
	h.Impersonate(rest.ImpersonationConfig{})

	// we need to add these groups for impersonation to work
	userGroups := []string{"system:authenticated", "system:authenticated:oauth"}
	if userGroup != "" {
		userGroups = append(userGroups, userGroup)
	}

	h.Impersonate(rest.ImpersonationConfig{
		UserName: user,
		Groups:   userGroups,
	})

	return err
}
