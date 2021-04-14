package verify

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	userv1 "github.com/openshift/api/user/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	"github.com/openshift/osde2e/pkg/common/helper"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
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

	var PRIVILEGED_USERS = []string{
		"system:admin",
	}

	var SRE_GROUPS = []string{
		"osd-sre-cluster-admins",
		"osd-sre-admins",
	}

	// Map of namespace name and whether it should be created/deleted by the test
	var PRIVILEGED_NAMESPACES = map[string]bool{
		"kube-admin":    true,
		"kube-foo":      true,
		"openshifter":   true,
		"openshift-foo": true,
		"openshift":     false,
		"default":       false,
	}

	// All namespaces in this list will be created/deleted by the test
	var REDHAT_NAMESPACES = []string{
		"redhat-user",
		"redhatuser",
	}

	// All namespaces in this list will be created/deleted by the test
	var NONPRIV_NAMESPACES = []string{
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
		ginkgo.JustBeforeEach(func() {

			_, err := createGroup(DUMMY_GROUP, h)
			Expect(err).NotTo(HaveOccurred())

			for privilegedNamespace, manageNamespace := range PRIVILEGED_NAMESPACES {
				if manageNamespace {
					_, err := createNamespace(privilegedNamespace, h)
					Expect(err).NotTo(HaveOccurred())
				}
			}
			for _, namespace := range append(REDHAT_NAMESPACES, NONPRIV_NAMESPACES...) {
				_, err := createNamespace(namespace, h)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		// Clean up all namespaces and groups created for the tests
		ginkgo.JustAfterEach(func() {
			h.Impersonate(rest.ImpersonationConfig{})
			err := deleteGroup(DUMMY_GROUP, h)
			Expect(err).NotTo(HaveOccurred())

			for privilegedNamespace, manageNamespace := range PRIVILEGED_NAMESPACES {
				if manageNamespace {
					h.Impersonate(rest.ImpersonationConfig{})
					err := deleteNamespace(privilegedNamespace, false, h)
					Expect(err).NotTo(HaveOccurred())
				}
			}
			for _, namespace := range append(REDHAT_NAMESPACES, NONPRIV_NAMESPACES...) {
				err := deleteNamespace(namespace, false, h)
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait until all namespaces have verified to be deleted

			namespacesToCheck := make([]string, 0)
			for ns, managed := range PRIVILEGED_NAMESPACES {
				if managed {
					namespacesToCheck = append(namespacesToCheck, ns)
				}
			}
			for _, ns := range append(REDHAT_NAMESPACES, NONPRIV_NAMESPACES...) {
				namespacesToCheck = append(namespacesToCheck, ns)
			}

			wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
				for _, ns := range namespacesToCheck {
					namespace, _ := h.Kube().CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
					if namespace != nil && namespace.Status.Phase == "Terminating" {
						return false, nil
					}
				}
				return true, nil
			})
		})

		ginkgo.It("dedicated admins cannot manage privileged namespaces", func() {
			for privilegedNamespace := range PRIVILEGED_NAMESPACES {
				err := updateNamespace(privilegedNamespace, DUMMY_USER, "dedicated-admins", h)
				Expect(err).To(HaveOccurred())
			}
			for _, namespace := range REDHAT_NAMESPACES {
				err := updateNamespace(namespace, DUMMY_USER, "dedicated-admins", h)
				Expect(err).To(HaveOccurred())
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		ginkgo.It("Non-privileged users cannot manage privileged namespaces", func() {
			for privilegedNamespace := range PRIVILEGED_NAMESPACES {
				err := updateNamespace(privilegedNamespace, DUMMY_USER, DUMMY_GROUP, h)
				Expect(err).To(HaveOccurred())
			}
			for _, namespace := range REDHAT_NAMESPACES {
				err := updateNamespace(namespace, DUMMY_USER, DUMMY_GROUP, h)
				Expect(err).To(HaveOccurred())
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		ginkgo.It("Members of SRE groups can manage all namespaces", func() {
			for _, sreGroup := range SRE_GROUPS {
				for privilegedNamespace := range PRIVILEGED_NAMESPACES {
					err := updateNamespace(privilegedNamespace, DUMMY_USER, sreGroup, h)
					Expect(err).NotTo(HaveOccurred())
				}
				for _, namespace := range append(REDHAT_NAMESPACES, NONPRIV_NAMESPACES...) {
					err := updateNamespace(namespace, DUMMY_USER, sreGroup, h)
					Expect(err).NotTo(HaveOccurred())
				}
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		ginkgo.It("Privileged users can manage all namespaces", func() {
			for _, privilegedUser := range PRIVILEGED_USERS {
				for privilegedNamespace := range PRIVILEGED_NAMESPACES {
					err := updateNamespace(privilegedNamespace, privilegedUser, "", h)
					Expect(err).NotTo(HaveOccurred())
				}
				for _, namespace := range append(REDHAT_NAMESPACES, NONPRIV_NAMESPACES...) {
					err := updateNamespace(namespace, privilegedUser, "", h)
					Expect(err).NotTo(HaveOccurred())
				}
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		ginkgo.It("Non-privileged users can manage all non-privileged namespaces", func() {
			// Non-privileged users can manage all non-privileged namespaces
			for _, nonPrivilegedNamespace := range NONPRIV_NAMESPACES {
				err := updateNamespace(nonPrivilegedNamespace, DUMMY_USER, "dedicated-admins", h)
				Expect(err).NotTo(HaveOccurred())
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})
})

func createGroup(groupName string, h *helper.H) (*userv1.Group, error) {
	group, err := h.User().UserV1().Groups().Get(context.TODO(), groupName, metav1.GetOptions{})
	if group != nil && err == nil {
		return group, err
	}
	log.Printf("Creating group for namespace validation webhook (%s)", groupName)
	group = &userv1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: groupName,
		},
	}
	return h.User().UserV1().Groups().Create(context.TODO(), group, metav1.CreateOptions{})
}

func deleteGroup(groupName string, h *helper.H) error {
	log.Printf("Deleting group for namespace validation webhook (%s)", groupName)
	return h.User().UserV1().Groups().Delete(context.TODO(), groupName, metav1.DeleteOptions{})
}

func updateNamespace(namespace string, asUser string, userGroup string, h *helper.H) (err error) {
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
		ns, err = h.Kube().CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to find namespace to update '%s': %v", namespace, err)
		}

		updatedNamespace, err = h.Kube().CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
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
		ns, err = h.Kube().CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to find updated namespace '%s': %v", namespace, err)
		}

		return updatedNamespace.ResourceVersion == ns.ResourceVersion, nil
	})

	return err
}

func deleteNamespace(namespace string, waitForDelete bool, h *helper.H) error {
	log.Printf("Deleting namespace for namespace validation webhook (%s)", namespace)
	err := h.Kube().CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace '%s': %v", namespace, err)
	}

	// Deleting a namespace can take a while. If desired, wait for the namespace to delete before returning.
	if waitForDelete {
		err = wait.PollImmediate(2*time.Second, 1*time.Minute, func() (bool, error) {
			ns, _ := h.Kube().CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
			if ns != nil && ns.Status.Phase == "Terminating" {
				return false, nil
			}
			return true, nil
		})
	}

	return err
}

func createNamespace(namespace string, h *helper.H) (*v1.Namespace, error) {

	// If the namespace already exists, we don't need to create it. Just return.
	ns, err := h.Kube().CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if ns != nil && ns.Status.Phase != "Terminating" && err == nil {
		return ns, err
	}

	log.Printf("Creating namespace for namespace validation webhook (%s)", namespace)
	ns = &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	h.Kube().CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})

	// Wait for the namespace to create. This is usually pretty quick.
	err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		if _, err := h.Kube().CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{}); err != nil {
			return false, nil
		}
		return true, nil
	})

	return ns, err
}
