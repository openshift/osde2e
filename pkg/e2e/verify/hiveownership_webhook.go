package verify

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ov1 "github.com/openshift/api/quota/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/util/wait"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var hiveownershipWebhookTestName string = "[Suite: informing] [OSD] hive ownership validating webhook"

func init() {
	alert.RegisterGinkgoAlert(hiveownershipWebhookTestName, "SD-SREP", "Boran Seref", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(hiveownershipWebhookTestName, func() {
	// const
	const (
		// Group to use for impersonation
		DUMMY_GROUP = "random-group-name"
		// User to use for impersonation
		DUMMY_USER = "testuser@testdomain"
	)

	// Map of CRQS name and whether it should be created/deleted by the test
	PRIVILEGED_CRQs := map[string]bool{
		"managed-first-quota":  true,
		"managed-second-quota": true,
		"openshift-quota":      false,
		"default":              false,
	}

	PRIVILEGED_USER := "system:admin"

	ELEVATED_SRE_USER := "backplane-cluster-admin"

	h := helper.New()
	ginkgo.Context("hiveownership validating webhook", func() {
		// Create all crqs needed for the tests
		ginkgo.JustBeforeEach(func(ctx context.Context) {
			_, err := createGroup(ctx, DUMMY_GROUP, h)
			Expect(err).NotTo(HaveOccurred())

			for privileged, managed := range PRIVILEGED_CRQs {
				err := createClusterResourceQuota(ctx, h, produceCRQ(privileged, managed), PRIVILEGED_USER, "")
				Expect(err).NotTo(HaveOccurred())
			}
		})

		// Clean up all clusterresourcequotas and groups generated for the tests
		ginkgo.JustAfterEach(func(ctx context.Context) {
			err := deleteGroup(ctx, DUMMY_GROUP, h)
			Expect(err).NotTo(HaveOccurred())

			for privileged := range PRIVILEGED_CRQs {
				err := deleteClusterResourceQuota(ctx, h, privileged, PRIVILEGED_USER, "")
				Expect(err).NotTo(HaveOccurred())
			}
		})

		// TESTS BEGIN

		// though https://github.com/openshift/managed-cluster-config/pull/626, we expect dedicated-admins cannot delete managed resources by protection of the hook.
		util.GinkgoIt("dedicated admins cannot delete managed CRQs", func(ctx context.Context) {
			for item, managed := range PRIVILEGED_CRQs {
				if managed {
					err := deleteClusterResourceQuota(ctx, h, item, DUMMY_USER, "dedicated-admins")
					Expect(err).To(HaveOccurred())
				}
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		// Passing constantly.
		util.GinkgoIt("a random user cannot delete managed CRQs", func(ctx context.Context) {
			for item, managed := range PRIVILEGED_CRQs {
				if managed {
					err := deleteClusterResourceQuota(ctx, h, item, DUMMY_USER, DUMMY_GROUP)
					Expect(err).To(HaveOccurred())
				}
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		// Passsing Constantly.
		util.GinkgoIt("Members of SRE can update a managed quota object", func(ctx context.Context) {
			for item, managed := range PRIVILEGED_CRQs {
				if managed {
					err := updateClusterResourceQuota(ctx, h, item, ELEVATED_SRE_USER, "")
					Expect(err).NotTo(HaveOccurred())
				}
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		// MCC TESTS(dedicated-admin changes - https://github.com/openshift/managed-cluster-config/pull/626)
		util.GinkgoIt("as dedicated admin can update crqs inside the cluster that are non managed.", func(ctx context.Context) {
			for item, managed := range PRIVILEGED_CRQs {
				if !managed {
					err := updateClusterResourceQuota(ctx, h, item, DUMMY_USER, "dedicated-admins")
					Expect(err).NotTo(HaveOccurred())
				}
			}
		}, viper.GetFloat64(config.Tests.PollingTimeout))

		util.GinkgoIt("as dedicated admin can create a crq inside the cluster that is non managed.", func(ctx context.Context) {
			cuQuota := "quota-customer"
			err := createClusterResourceQuota(ctx, h, produceCRQ(cuQuota, false), DUMMY_USER, "dedicated-admins")
			Expect(err).NotTo(HaveOccurred())
			err = deleteClusterResourceQuota(ctx, h, cuQuota, DUMMY_USER, "dedicated-admins")
			Expect(err).NotTo(HaveOccurred())
		}, viper.GetFloat64(config.Tests.PollingTimeout))
		// ENDS
	})
})

// CRUD OPERATIONS

// createClusterResourceQuota creates a clusterResourceQuota obj in the cluster as given user
func createClusterResourceQuota(ctx context.Context, h *helper.H, crq *ov1.ClusterResourceQuota, asUser, userGroup string) (err error) {
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

	// set up cli
	cli, err := h.Quota()
	if err != nil {
		return fmt.Errorf("cannot create quota client")
	}

	err = wait.PollImmediate(5*time.Second, 10*time.Second, func() (bool, error) {
		// create the CRQ
		quotas, err := cli.QuotaV1().ClusterResourceQuotas().Create(ctx, crq, metav1.CreateOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to create ClusterResourceQuota: '%s': %v", quotas, err)
		}

		return true, nil
	})
	return err
}

// deleteClusterResourceQuota deletes a given clusterResourceQuota obj
func deleteClusterResourceQuota(ctx context.Context, h *helper.H, name, asUser, userGroup string) (err error) {
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

	// set up cli
	cli, err := h.Quota()
	if err != nil {
		return fmt.Errorf("cannot create quota client")
	}

	err = wait.PollImmediate(5*time.Second, 10*time.Second, func() (bool, error) {
		// delete the CRQ
		err := cli.QuotaV1().ClusterResourceQuotas().Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to delete ClusterResourceQuota: '%s': %v", name, err)
		}

		return true, nil
	})
	return err
}

// updateClusterResourceQuota updates the label with a random text for a given clusterResourceQuota obj
func updateClusterResourceQuota(ctx context.Context, h *helper.H, name, asUser, userGroup string) (err error) {
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

	// set up cli
	cli, err := h.Quota()
	if err != nil {
		return fmt.Errorf("cannot create quota client")
	}

	crq, err := cli.QuotaV1().ClusterResourceQuotas().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ClusterResourceQuota: '%s': %v", name, err)
	}

	// update the CRQ
	updated, err := cli.QuotaV1().ClusterResourceQuotas().Update(ctx, updateCRQ(crq), metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ClusterResourceQuota: '%s': %v", updated, err)
	}

	return err
}

// ClusterResourceQuota Producer/Updater
func produceCRQ(name string, managed bool) *ov1.ClusterResourceQuota {
	labels := map[string]string{"dummy": "true"}
	annotation := map[string]string{"openshift.io/requester": "test"}

	if managed {
		labels = map[string]string{
			"hive.openshift.io/managed": "true",
		}
	}

	return &ov1.ClusterResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            labels,
		},
		Spec: ov1.ClusterResourceQuotaSpec{
			Selector: ov1.ClusterResourceQuotaSelector{
				AnnotationSelector: annotation,
			},
		},
		Status: ov1.ClusterResourceQuotaStatus{
			Total: v1.ResourceQuotaStatus{},
		},
	}
}

func updateCRQ(crq *ov1.ClusterResourceQuota) *ov1.ClusterResourceQuota {
	labels := crq.GetLabels()
	labels[util.RandomStr(5)] = util.RandomStr(5)
	crq.SetLabels(labels)
	return crq
}
