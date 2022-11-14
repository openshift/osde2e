package operators

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/api/config/v1"
	upgradev1alpha1 "github.com/openshift/managed-upgrade-operator/pkg/apis/upgrade/v1alpha1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/prometheus"
	"github.com/openshift/osde2e/pkg/common/util"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

var managedUpgradeOperatorTestName string = "[Suite: informing] [OSD] Managed Upgrade Operator"

func init() {
	alert.RegisterGinkgoAlert(managedUpgradeOperatorTestName, "SD-SREP", "@managed-upgrade-operator", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(managedUpgradeOperatorTestName, func() {
	operatorName := "managed-upgrade-operator"
	var operatorNamespace string = "openshift-managed-upgrade-operator"
	var operatorLockFile string = "managed-upgrade-operator-lock"
	var upgradeConfigResourceName string = "managed-upgrade-config"
	var upgradeConfigForDedicatedAdminTestName string = "osde2e-da-upgrade-config"
	var defaultDesiredReplicas int32 = 1
	clusterRoles := []string{
		"managed-upgrade-operator",
	}
	clusterRoleBindings := []string{
		"managed-upgrade-operator",
	}
	h := helper.New()
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkRoleBindings(h,
		operatorNamespace,
		[]string{"prometheus-k8s"})

	// this operator's clusterroles have a version suffix, so only check the prefix
	checkClusterRoles(h, clusterRoles, true)
	checkClusterRoleBindings(h, clusterRoleBindings, true)

	ginkgo.Context("when an upgrade config is received", func() {
		var (
			clusterVersion *v1.ClusterVersion = nil
			err            error
		)

		ginkgo.BeforeEach(func(ctx context.Context) {
			clusterVersion, err = getClusterVersion(ctx, h)
		})

		util.GinkgoIt("should not upgrade if the upgrade time is in the future", func(ctx context.Context) {
			// Validate clusterversion
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterVersion).NotTo(BeNil())
			if len(clusterVersion.Status.AvailableUpdates) == 0 {
				// We can't do this test if the cluster has no updates available
				return
			}

			// If there is an existing upgrade config, we must be in a post-upgrade
			// state, so no need to re-run the tests
			existingUc, _ := getUpgradeConfig(ctx, upgradeConfigResourceName, operatorNamespace, h)
			if existingUc != nil {
				ginkgo.Skip("skipping due to existing UpgradeConfig")
			}

			// Pick a version to upgrade to, we don't care which as we're not actually upgrading
			targetVersion := clusterVersion.Status.AvailableUpdates[0].Version
			targetChannel := clusterVersion.Spec.Channel

			startTime := time.Now().UTC().Add(12 * time.Hour)

			// Add the upgradeconfig to the cluster
			uc := makeUpgradeConfig(upgradeConfigResourceName, operatorNamespace, startTime.Format(time.RFC3339), targetVersion, targetChannel)
			err = addUpgradeConfig(ctx, uc, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())
			// Delete the upgradeconfig after the test
			defer func() {
				err := deleteUpgradeConfig(ctx, upgradeConfigResourceName, operatorNamespace, h)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Wait a minute and see whether the upgradeconfig history phase changes
			err = wait.Poll(1*time.Minute, 2*time.Minute, func() (bool, error) {
				ucObj, err := h.Dynamic().Resource(schema.GroupVersionResource{
					Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
				}).Namespace(operatorNamespace).Get(ctx, upgradeConfigResourceName, metav1.GetOptions{})
				if err != nil {
					return false, fmt.Errorf("unable to retrieve upgradeconfig")
				}

				var upgradeConfig upgradev1alpha1.UpgradeConfig
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(ucObj.UnstructuredContent(), &upgradeConfig)
				if err != nil {
					return false, fmt.Errorf("error parsing upgradeconfig into object")
				}

				upgradeHistory := upgradeConfig.Status.History.GetHistory(targetVersion)
				// If the operator hasn't processed the upgradeconfig yet, wait a bit longer
				if upgradeHistory == nil {
					return false, nil
				}
				// If the phase changes to anything other than 'New', something unexpected has happened
				if upgradeHistory.Phase != upgradev1alpha1.UpgradePhasePending {
					return false, fmt.Errorf("upgradeconfig phase is not 'New', is: %v", upgradeHistory.Phase)
				}
				// The status remains New after a minute, that's fine
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		})

		util.GinkgoIt("should error if the upgrade time is too far in the past", func(ctx context.Context) {
			// Validate clusterversion
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterVersion).NotTo(BeNil())
			if len(clusterVersion.Status.AvailableUpdates) == 0 {
				// We can't do this test if the cluster has no updates available
				return
			}

			// If there is an existing upgrade config, we must be in a post-upgrade
			// state, so no need to re-run the tests
			existingUc, _ := getUpgradeConfig(ctx, upgradeConfigResourceName, operatorNamespace, h)
			if existingUc != nil {
				ginkgo.Skip("skipping due to existing UpgradeConfig")
			}

			targetVersion := clusterVersion.Status.AvailableUpdates[0].Version
			targetChannel := clusterVersion.Spec.Channel

			// Set a start time of 12 hours ago
			startTime := time.Now().UTC().Add(-12 * time.Hour)

			// Add the upgradeconfig to the cluster
			uc := makeUpgradeConfig(upgradeConfigResourceName, operatorNamespace, startTime.Format(time.RFC3339), targetVersion, targetChannel)
			err = addUpgradeConfig(ctx, uc, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())
			// Delete the upgradeconfig after the test
			defer func() {
				err := deleteUpgradeConfig(ctx, upgradeConfigResourceName, operatorNamespace, h)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Get a Prom API connection
			promClient, err := prometheus.CreateClusterClient(h)
			Expect(err).NotTo(HaveOccurred())
			promAPI := prometheusv1.NewAPI(promClient)

			// Wait a minute for the operator to flag this as a problem
			err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
				query := fmt.Sprintf("upgradeoperator_upgrade_window_breached{upgradeconfig_name=\"%s\"} == 1", "osde2e-upgrade-config")
				context, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				value, _, err := promAPI.Query(context, query, time.Now())
				if err != nil {
					return false, fmt.Errorf("Unable to query prom API")
				}
				vector, _ := value.(model.Vector)
				if vector.Len() != 1 {
					return false, nil
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		})

		util.GinkgoIt("should error if provided an invalid start time", func(ctx context.Context) {
			// Validate clusterversion
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterVersion).NotTo(BeNil())
			if len(clusterVersion.Status.AvailableUpdates) == 0 {
				// We can't do this test if the cluster has no updates available
				return
			}

			// If there is an existing upgrade config, we must be in a post-upgrade
			// state, so no need to re-run the tests
			existingUc, _ := getUpgradeConfig(ctx, upgradeConfigResourceName, operatorNamespace, h)
			if existingUc != nil {
				ginkgo.Skip("skipping due to existing UpgradeConfig")
			}

			targetVersion := clusterVersion.Status.AvailableUpdates[0].Version
			targetChannel := clusterVersion.Spec.Channel

			// Add the upgradeconfig to the cluster
			uc := makeUpgradeConfig(upgradeConfigResourceName, operatorNamespace, "this is not a start time", targetVersion, targetChannel)
			err = addUpgradeConfig(ctx, uc, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())
			// Delete the upgradeconfig after the test
			defer func() {
				err := deleteUpgradeConfig(ctx, upgradeConfigResourceName, operatorNamespace, h)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Get a Prom API connection
			promClient, err := prometheus.CreateClusterClient(h)
			Expect(err).NotTo(HaveOccurred())
			promAPI := prometheusv1.NewAPI(promClient)

			// Wait a minute for the operator to flag this as a problem
			err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
				query := fmt.Sprintf("upgradeoperator_upgradeconfig_validation_failed{upgradeconfig_name=\"%s\"} == 1", upgradeConfigResourceName)
				context, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				value, _, err := promAPI.Query(context, query, time.Now())
				if err != nil {
					return false, fmt.Errorf("Unable to query prom API")
				}
				vector, _ := value.(model.Vector)
				if vector.Len() != 1 {
					return false, nil
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
	ginkgo.Context("upgradeconfig", func() {
		util.GinkgoIt("dedicated admin should not be able to manage the UpgradeConfig CR", func(ctx context.Context) {
			// Add the upgradeconfig to the cluster
			uc := makeMinimalUpgradeConfig(upgradeConfigForDedicatedAdminTestName, operatorNamespace)
			err := dedicatedAaddUpgradeConfig(ctx, uc, operatorNamespace, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

			err = addUpgradeConfig(ctx, uc, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())

			err = dedicatedADeleteUpgradeConfig(ctx, upgradeConfigForDedicatedAdminTestName, operatorNamespace, h)
			Expect(apierrors.IsForbidden(err)).To(BeTrue())

			err = deleteUpgradeConfig(ctx, upgradeConfigForDedicatedAdminTestName, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func getClusterVersion(ctx context.Context, h *helper.H) (*v1.ClusterVersion, error) {
	// get cluster version
	cfgClient := h.Cfg()
	getOpts := metav1.GetOptions{}
	cVersion, err := cfgClient.ConfigV1().ClusterVersions().Get(ctx, "version", getOpts)
	if err != nil {
		return nil, fmt.Errorf("couldn't get current ClusterVersion '%s': %v", "version", err)
	}
	return cVersion, nil
}

func makeUpgradeConfig(name string, ns string, startTime string, version string, channel string) upgradev1alpha1.UpgradeConfig {
	uc := upgradev1alpha1.UpgradeConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UpgradeConfig",
			APIVersion: "upgrade.managed.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: upgradev1alpha1.UpgradeConfigSpec{
			Desired: upgradev1alpha1.Update{
				Version: version,
				Channel: channel,
			},
			UpgradeAt:            startTime,
			PDBForceDrainTimeout: 60,
			Type:                 upgradev1alpha1.OSD,
		},
	}
	return uc
}

func makeMinimalUpgradeConfig(name string, ns string) upgradev1alpha1.UpgradeConfig {
	uc := upgradev1alpha1.UpgradeConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UpgradeConfig",
			APIVersion: "upgrade.managed.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: upgradev1alpha1.UpgradeConfigSpec{
			Type: upgradev1alpha1.OSD,
		},
	}
	return uc
}

func addUpgradeConfig(ctx context.Context, upgradeConfig upgradev1alpha1.UpgradeConfig, operatorNamespace string, h *helper.H) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(upgradeConfig.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
	}).Namespace(operatorNamespace).Create(ctx, &unstructuredObj, metav1.CreateOptions{})
	return err
}

func deleteUpgradeConfig(ctx context.Context, name string, operatorNamespace string, h *helper.H) error {
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
	}).Namespace(operatorNamespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func getUpgradeConfig(ctx context.Context, name string, ns string, h *helper.H) (*upgradev1alpha1.UpgradeConfig, error) {
	ucObj, err := h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
	}).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error retrieving upgradeconfig: %v", err)
	}
	var upgradeConfig upgradev1alpha1.UpgradeConfig
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ucObj.UnstructuredContent(), &upgradeConfig)
	if err != nil {
		// This, however, is probably error-worthy because it means our UpgradeConfig
		// has been messed with or something odd's occurred
		return nil, fmt.Errorf("error parsing upgradeconfig into object")
	}

	return &upgradeConfig, nil
}

// test for CR customresourcedefinition.apiextensions.k8s.io/upgradeconfigs.upgrade.managed.openshift.io
// dedicated admin should not be able to create/edit this CR

func dedicatedAaddUpgradeConfig(ctx context.Context, upgradeConfig upgradev1alpha1.UpgradeConfig, operatorNamespace string, h *helper.H) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(upgradeConfig.DeepCopy())
	if err != nil {
		return err
	}
	h.Impersonate(rest.ImpersonationConfig{
		UserName: "test-user@redhat.com",
		Groups: []string{
			"dedicated-admins",
		},
	})
	defer func() {
		h.Impersonate(rest.ImpersonationConfig{})
	}()
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
	}).Namespace(operatorNamespace).Create(ctx, &unstructuredObj, metav1.CreateOptions{})
	return (err)
}

func dedicatedADeleteUpgradeConfig(ctx context.Context, name string, operatorNamespace string, h *helper.H) error {
	h.Impersonate(rest.ImpersonationConfig{
		UserName: "test-user@redhat.com",
		Groups: []string{
			"dedicated-admins",
		},
	})
	defer func() {
		h.Impersonate(rest.ImpersonationConfig{})
	}()
	return h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
	}).Namespace(operatorNamespace).Delete(ctx, name, metav1.DeleteOptions{})
}
