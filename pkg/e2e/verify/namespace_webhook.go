package verify

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/helper"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var namespaceWebhookTestName string = "[Suite: e2e] [OSD] namespace validating webhook"

func init() {
	alert.RegisterGinkgoAlert(namespaceWebhookTestName, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(namespaceWebhookTestName, func() {
	const (
		// Group to use for impersonation
		DUMMY_GROUP = "random-group-name"
		// User to use for impersonation
		DUMMY_USER = "testuser@testdomain"
	)

	PRIVILEGED_USERS := []string{
		"system:admin",
		"backplane-cluster-admin",
	}

	// Map of namespace name and whether it should be created/deleted by the test
	// Should match up with namespaces found in managed-cluster-config:
	// * https://github.com/openshift/managed-cluster-config/blob/master/deploy/osd-managed-resources/ocp-namespaces.ConfigMap.yaml
	// * https://github.com/openshift/managed-cluster-config/blob/master/deploy/osd-managed-resources/managed-namespaces.ConfigMap.yaml
	PRIVILEGED_NAMESPACES := map[string]bool{
		"kube-system":                    false,
		"openshift-apiserver":            false,
		"openshift":                      false,
		"default":                        false,
		"redhat-ocm-addon-test-operator": true,
	}

	// All namespaces in this list will be created/deleted by the test
	NONPRIV_NAMESPACES := []string{
		"mykube-admin",
		"open-shift",
		"oopenshift",
		"ops-health-monitoring-foo",
		"default-user",
		"logger",
	}

	h := helper.New()

	ginkgo.Context("namespace validating webhook", func() {
		// Create all namespaces and groups needed for the tests
		ginkgo.JustBeforeEach(func(ctx context.Context) {
			_, err := createGroup(ctx, DUMMY_GROUP, h)
			Expect(err).NotTo(HaveOccurred())

			for privilegedNamespace, manageNamespace := range PRIVILEGED_NAMESPACES {
				if manageNamespace {
					_, err := createNamespace(ctx, privilegedNamespace, h)
					Expect(err).NotTo(HaveOccurred())
				}
			}
			for _, namespace := range NONPRIV_NAMESPACES {
				_, err := createNamespace(ctx, namespace, h)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		// Clean up all namespaces and groups created for the tests
		ginkgo.JustAfterEach(func(ctx context.Context) {
			h.Impersonate(rest.ImpersonationConfig{})
			err := deleteGroup(ctx, DUMMY_GROUP, h)
			Expect(err).NotTo(HaveOccurred())

			for privilegedNamespace, manageNamespace := range PRIVILEGED_NAMESPACES {
				if manageNamespace {
					h.Impersonate(rest.ImpersonationConfig{})
					err := deleteNamespace(ctx, privilegedNamespace, false, h)
					Expect(err).NotTo(HaveOccurred())
				}
			}
			for _, namespace := range NONPRIV_NAMESPACES {
				err := deleteNamespace(ctx, namespace, false, h)
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait until all namespaces have verified to be deleted
			namespacesToCheck := make([]string, 0)
			for ns, managed := range PRIVILEGED_NAMESPACES {
				if managed {
					namespacesToCheck = append(namespacesToCheck, ns)
				}
			}
			for _, ns := range NONPRIV_NAMESPACES {
				namespacesToCheck = append(namespacesToCheck, ns)
			}

			wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
				for _, ns := range namespacesToCheck {
					namespace, _ := h.Kube().CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
					if namespace != nil && namespace.Status.Phase == "Terminating" {
						return false, nil
					}
				}
				return true, nil
			})
		})

		util.GinkgoIt("dedicated admins cannot manage privileged namespaces", func(ctx context.Context) {
			for privilegedNamespace := range PRIVILEGED_NAMESPACES {
				err := updateNamespace(ctx, privilegedNamespace, DUMMY_USER, "dedicated-admins", h)
				Expect(err).To(HaveOccurred())
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		util.GinkgoIt("Non-privileged users cannot manage privileged namespaces", func(ctx context.Context) {
			for privilegedNamespace := range PRIVILEGED_NAMESPACES {
				err := updateNamespace(ctx, privilegedNamespace, DUMMY_USER, DUMMY_GROUP, h)
				Expect(err).To(HaveOccurred())
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		util.GinkgoIt("Privileged users can manage all namespaces", func(ctx context.Context) {
			for _, privilegedUser := range PRIVILEGED_USERS {
				for privilegedNamespace := range PRIVILEGED_NAMESPACES {
					err := updateNamespace(ctx, privilegedNamespace, privilegedUser, "", h)
					Expect(err).NotTo(HaveOccurred())
				}
				for _, namespace := range NONPRIV_NAMESPACES {
					err := updateNamespace(ctx, namespace, privilegedUser, "", h)
					Expect(err).NotTo(HaveOccurred())
				}
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		util.GinkgoIt("Non-privileged users can manage all non-privileged namespaces", func(ctx context.Context) {
			// Non-privileged users can manage all non-privileged namespaces
			for _, nonPrivilegedNamespace := range NONPRIV_NAMESPACES {
				err := updateNamespace(ctx, nonPrivilegedNamespace, DUMMY_USER, "dedicated-admins", h)
				Expect(err).NotTo(HaveOccurred())
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})
})

func createGroup(ctx context.Context, groupName string, h *helper.H) (*userv1.Group, error) {
	group, err := h.User().UserV1().Groups().Get(ctx, groupName, metav1.GetOptions{})
	if group != nil && err == nil {
		return group, err
	}
	log.Printf("Creating group for namespace validation webhook (%s)", groupName)
	group = &userv1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: groupName,
		},
	}
	return h.User().UserV1().Groups().Create(ctx, group, metav1.CreateOptions{})
}

func deleteGroup(ctx context.Context, groupName string, h *helper.H) error {
	log.Printf("Deleting group for namespace validation webhook (%s)", groupName)
	return h.User().UserV1().Groups().Delete(ctx, groupName, metav1.DeleteOptions{})
}

func updateNamespace(ctx context.Context, namespace string, asUser string, userGroup string, h *helper.H) (err error) {
	// reset impersonation upon return
	defer h.Impersonate(rest.ImpersonationConfig{})

	// reset impersonation at the beginning just-in-case
	h.Impersonate(rest.ImpersonationConfig{})

	// we need to add these groups for impersonation to work
	userGroups := []string{"system:authenticated", "system:authenticated:oauth"}
	if userGroup != "" {
		userGroups = append(userGroups, userGroup)
	}

	// update the namespace as our desired user
	h.Impersonate(rest.ImpersonationConfig{
		UserName: asUser,
		Groups:   userGroups,
	})

	var updatedNamespace *v1.Namespace
	var ns *v1.Namespace

	err = wait.PollImmediate(10*time.Second, 1*time.Minute, func() (bool, error) {
		// Verify the namespace already exists
		ns, err = h.Kube().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to find namespace to update '%s': %v", namespace, err)
		}

		updatedNamespace, err = h.Kube().CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
		if err != nil {
			if apierrors.IsConflict(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	err = wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
		ns, err = h.Kube().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to find updated namespace '%s': %v", namespace, err)
		}

		return updatedNamespace.ResourceVersion == ns.ResourceVersion, nil
	})

	return err
}

func deleteNamespace(ctx context.Context, namespace string, waitForDelete bool, h *helper.H) error {
	log.Printf("Deleting namespace for namespace validation webhook (%s)", namespace)
	err := h.Kube().CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace '%s': %v", namespace, err)
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

func createNamespace(ctx context.Context, namespace string, h *helper.H) (*v1.Namespace, error) {
	// If the namespace already exists, we don't need to create it. Just return.
	ns, err := h.Kube().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if ns != nil && ns.Status.Phase != "Terminating" && err == nil {
		return ns, err
	}

	log.Printf("Creating namespace for namespace validation webhook (%s)", namespace)
	labels := map[string]string{
		"pod-security.kubernetes.io/enforce":             "privileged",
		"pod-security.kubernetes.io/audit":               "privileged",
		"pod-security.kubernetes.io/warn":                "privileged",
		"security.openshift.io/scc.podSecurityLabelSync": "false",
	}
	ns = &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: labels,
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
