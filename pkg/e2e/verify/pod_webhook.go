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

var podWebhookTestName string = "[Suite: informing] [OSD] pod validating webhook"

func init() {
	alert.RegisterGinkgoAlert(podWebhookTestName, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

const (
	serviceStartTimeout = 1 * time.Minute

	daemonStartTimeout = 1 * time.Minute
)

var _ = ginkgo.Describe(podWebhookTestName, func() {

	h := helper.New()

	ginkgo.Context("pod webhook", func() {

		ginkgo.It("Verify the validation webhook service is running", func() {
			namespace := "openshift-validation-webhook"
			daemonSetName := "validation-webhook"
			serviceName := "validation-webhook"

			err := h.WaitTimeoutForDaemonSetInNamespace(daemonSetName, namespace, daemonStartTimeout, poll)
			Expect(err).NotTo(HaveOccurred(), "No Daemonset named %s found.", daemonSetName)

			err = h.WaitTimeoutForServiceInNamespace(serviceName, namespace, serviceStartTimeout, poll)
			Expect(err).NotTo(HaveOccurred(), "No service named %s found.", serviceName)

		})

		// for all tests, "manage" is synonymous with "create/update/delete"
		//Dedicated admin can not deploy pod on master on infra nodes in namespaces
		//openshift-operators, openshift-logging namespace or any other namespace that is not a core namespace like openshift-*, redhat-*, default, kube-*.

		ginkgo.It("Test 1: Webhook will mark pod spec invalid and block deploying", func() {
			name := "osde2e-pod-webhook-test1"
			namespace := "openshift-logging"
			createNamespace(namespace, h)
			defer deleteNamespace(namespace, true, h)
			//impersonate dedicated-admin
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()
			_, err := createPod(name, namespace, "node-role.kubernetes.io/master", "toleration-key-value", v1.TaintEffectNoSchedule, "node-role.kubernetes.io/infra", "toleration-key-value2", v1.TaintEffectNoSchedule, h)
			defer deletePod(name, namespace, h)
			log.Printf("Create pod error: %v", err)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})

		ginkgo.It("Test 2: Webhook will mark pod spec invalid and block deploying", func() {
			name := "osde2e-pod-webhook-test2"
			namespace := "openshift-logging"
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test-user@redhat.com",
				Groups: []string{
					"dedicated-admins",
					"system:authenticated",
				},
			})
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()

			defer deletePod(name, namespace, h)
			_, err := createPod(name, namespace, "node-role.kubernetes.io/infra", "toleration-key-value", v1.TaintEffectPreferNoSchedule, "node-role.kubernetes.io/master", "toleration-key-value2", v1.TaintEffectNoExecute, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

		})

		// The serviceaccount:dedicated-admin-project is allowed to launch a pod and the pod-webhook will allow it
		ginkgo.It("Webhook will allow pod to deploy", func() {
			name := "osde2e-pod-webhook-test3"
			namespace := "openshift-apiserver"
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			defer deletePod(name, namespace, h)
			_, err := createPod(name, namespace, "node-role.kubernetes.io/infra", "toleration-key-value", v1.TaintEffectNoSchedule, "node-role.kubernetes.io/master", "toleration-key-value2", v1.TaintEffectNoSchedule, h)
			Expect(err).NotTo(HaveOccurred())
		})

		// RBAC blocks dedicated-admins group from creating a pod in openshift-apiserver namespace
		ginkgo.It("Webhook will allow pod to deploy", func() {
			name := "osde2e-pod-webhook-test3"
			namespace := "openshift-apiserver"
			impersonateDedicatedAdmin(h, "test-user")
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
			}()
			defer deletePod(name, namespace, h)
			_, err := createPod(name, namespace, "node-role.kubernetes.io/infra", "toleration-key-value", v1.TaintEffectNoSchedule, "node-role.kubernetes.io/master", "toleration-key-value2", v1.TaintEffectNoSchedule, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})

		// RBAC will prevent ordinary users from creating pods

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
			defer deletePod(name, namespace, h)
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
			createNamespace(namespace, h)
			defer deleteNamespace(namespace, true, h)
			err := asUser(namespace, user, userGroup, h)
			if err != nil {
				log.Printf("Could not impersonate user, Error %v", err)
				return
			}
			defer h.Impersonate(rest.ImpersonationConfig{})
			defer deletePod(name, namespace, h)
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

	//log.Printf("Creating pod for the validation webhook (%s)", pod)
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

func impersonateDedicatedAdmin(h *helper.H, user string) *helper.H {
	h.Impersonate(rest.ImpersonationConfig{
		UserName: user,
		Groups: []string{
			"dedicated-admins",
			"system:authenticated",
		},
	})

	return h
}
