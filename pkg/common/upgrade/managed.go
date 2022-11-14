package upgrade

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openshift/osde2e/pkg/common/cluster/healthchecks"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"

	configv1 "github.com/openshift/api/config/v1"
	upgradev1alpha1 "github.com/openshift/managed-upgrade-operator/pkg/apis/upgrade/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/templates"
)

const (
	// Namespace in which the managed-upgrade-operator runs
	muoNamespace = "openshift-managed-upgrade-operator"
	// the name of the generated UpgradeConfig resource containing upgrade configuration
	upgradeConfigName = "managed-upgrade-config"

	// Name of the provider for syncing upgrade policy from OCM
	providerOCM = "OCM"
	// Name of the provider for syncing an UpgradeConfig locally
	providerLocal = "LOCAL"

	// name of the workload for pod disruption budget tests
	pdbWorkloadName = "pdb"
	// directory containing pod disruption budget workload assets
	pdbWorkloadDir = "workloads/e2e/pdb"
	// name of the workload for node drain tests
	drainWorkloadName = "node-drain-test"
	// directory containing node drain workload assets
	drainWorkloadDir = "workloads/e2e/drain"
	// Time to wait in seconds for workload to be created
	workloadCreationWaitTime = 3

	// config override template asset
	configOverrideTemplate = "upgrades/config.template"
	// config override values
	configProviderWatchInterval   = 15 // minutes
	configScaleTimeout            = 15 // minutes
	configUpgradeWindow           = 30 // minutes
	configNodeDrainTimeout        = 6  // minutes
	configExpectedDrainTime       = 7  // minutes
	configControlPlaneTime        = 90 // minutes
	configPdbDrainTimeoutOverride = 5  // minutes

)

// TriggerManagedUpgrade initiates an upgrade using the managed-upgrade-operator
func TriggerManagedUpgrade(h *helper.H) (*configv1.Update, error) {
	// Create any pre-upgrade workloads to test the managed-upgrade-operator with
	err := createUpgradeClusterWorkloads(h)
	if err != nil {
		return nil, err
	}

	// override the Hive-managed operator config with our testing one, so that
	// we don't give worker upgrades the same upgrade grace periods
	err = overrideOperatorConfig(h)
	if err != nil {
		return nil, fmt.Errorf("unable to override operator configuration: %v", err)
	}

	// Create the upgrade config and initiate the upgrade process
	err = scheduleUpgradeWithProvider(h)
	if err != nil {
		return nil, fmt.Errorf("can't initiate managed upgrade: %v", err)
	}

	// We need to force the operator to resync with its provider. Whilst this would happen naturally,
	// it's too long to wait for E2E (could be up to 1 hour). In lieu of a better way to trigger this,
	// let's bounce the deployment to hurry that process.
	err = restartOperator(h, muoNamespace)
	if err != nil {
		return nil, fmt.Errorf("error restarting managed-upgrade-operator: %v", err)
	}

	// wait for a few seconds to get the upgradeconfig synced from upgradepolicy
	for c := 0; c < 6; c++ {
		ucCreated, _ := isUpgradeConfigCreated(h)
		if !ucCreated {
			time.Sleep(30 * time.Second)
		}
	}

	// Reschedule the upgrade if flag specified
	if viper.GetBool(config.Upgrade.ManagedUpgradeRescheduled) {
		err = updateUpgradeWithProvider()
		if err != nil {
			return nil, fmt.Errorf("can't reschedule upgrade: %v", err)
		}
	}

	// The managed-upgrade-operator won't have updated the CVO version yet, and that's fine.
	// But let's return what it will look like, for the later 'is it upgraded yet' tests.
	cUpdate := &configv1.Update{}

	releaseName := viper.GetString(config.Upgrade.ReleaseName)
	if releaseName != "" {
		upgradeVersion, err := util.OpenshiftVersionToSemver(releaseName)
		if err != nil {
			cUpdate.Version = upgradeVersion.String()
		}
	}
	image := viper.GetString(config.Upgrade.Image)
	if image != "" {
		cUpdate.Image = image
	}

	return cUpdate, nil
}

// Override the managed-upgrade-operator's existing configmap with an e2e-focused one, if
// the existing configmap contains different values
func overrideOperatorConfig(h *helper.H) error {
	// Retrieve the existing operator config data
	cm, err := h.Kube().CoreV1().ConfigMaps(muoNamespace).Get(context.TODO(), "managed-upgrade-operator-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("managed-upgrade-operator configmap not found: %v", err)
	}
	cfgData, cfgFound := cm.Data["config.yaml"]
	if !cfgFound {
		return fmt.Errorf("managed-upgrade-operator configmap missing mandatory key config.yaml")
	}

	// select correct environment
	providerEnv := viper.GetString(ocmprovider.Env)
	url := ocmprovider.Environments.Choose(providerEnv)
	replaceValues := struct {
		ProviderSource         string
		ProviderEnvironmentUrl string
		ProviderWatchInterval  int
		ControlPlaneTime       int
		ScaleTimeout           int
		UpgradeWindow          int
		NodeDrainTimeout       int
		ExpectedDrainTime      int
	}{
		ProviderSource:         getProviderSource(),
		ProviderEnvironmentUrl: url,
		ProviderWatchInterval:  configProviderWatchInterval,
		ControlPlaneTime:       configControlPlaneTime,
		ScaleTimeout:           configScaleTimeout,
		UpgradeWindow:          configUpgradeWindow,
		NodeDrainTimeout:       configNodeDrainTimeout,
		ExpectedDrainTime:      configExpectedDrainTime,
	}

	configOverrideTemplate, err := templates.LoadTemplate(configOverrideTemplate)
	if err != nil {
		return fmt.Errorf("can't read upgrade config override template: %v", err)
	}
	configOverride, err := h.ConvertTemplateToString(configOverrideTemplate, replaceValues)
	if err != nil {
		return fmt.Errorf("can't parse upgrade config override template: %v", err)
	}

	// only update if we observe a difference between current config and testing config
	if cfgData != configOverride {
		cm.Data["config.yaml"] = configOverride
		_, err = h.Kube().CoreV1().ConfigMaps(muoNamespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("managed-upgrade-operator configmap could not be updated: %v", err)
		}
	}

	return nil
}

// Override the cluster's UpgradeConfig with custom values where needed.
// This is primarily being used to override the Pod Disruption Budget timeout
// in order to minimize the worker upgrade length.
func overrideUpgradeConfig(uc upgradev1alpha1.UpgradeConfig, h *helper.H) error {
	// only update if we need to
	if uc.Spec.PDBForceDrainTimeout == configPdbDrainTimeoutOverride {
		return nil
	}

	uc.Spec.PDBForceDrainTimeout = configPdbDrainTimeoutOverride
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(uc.DeepCopy())
	if err != nil {
		return fmt.Errorf("can't convert UpgradeConfig to unstructured resource: %v", err)
	}
	uobj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "upgrade.managed.openshift.io",
		Version:  "v1alpha1",
		Resource: "upgradeconfigs",
	}).Namespace(muoNamespace).Update(context.TODO(), &uobj, metav1.UpdateOptions{})

	if err != nil {
		return err
	}
	return nil
}

func createManagedUpgradeWorkload(workLoadName string, workLoadDir string, podPrefixes []string, h *helper.H) error {
	if _, ok := h.GetWorkload(workLoadName); ok {
		return nil
	}

	// Create all K8s objects that are within the workLoadDir
	log.Printf("Applying %s workload from %s\n", workLoadName, workLoadDir)
	obj, err := helper.ApplyYamlInFolder(workLoadDir, h.CurrentProject(), h.Kube())
	if err != nil {
		return fmt.Errorf("can't create %s workload: %v", workLoadName, err)
	}

	// Log how many objects have been created
	log.Printf("%v object(s) created for %s workload from %s path\n", len(obj), workLoadName, workLoadDir)

	// Give the cluster a second to churn before checking
	time.Sleep(workloadCreationWaitTime * time.Second)

	// Wait for all pods to come up healthy
	err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		if check, err := healthchecks.CheckPodHealth(h.Kube().CoreV1(), nil, h.CurrentProject(), podPrefixes...); !check || err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("%s workload not running correctly: %v", workLoadName, err)
	}

	// If success, add the workload to the list of installed workloads
	h.AddWorkload(workLoadName, h.CurrentProject())

	return nil
}

// IsManagedUpgradeDone returns with done true when a managed upgrade is complete.
func isManagedUpgradeDone(h *helper.H) (done bool, msg string, err error) {
	// retrieve UpgradeConfig
	ucList, err := h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
	}).Namespace(muoNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// The API may sometimes be unavailable, this is fine. We don't want
		// to return an error because there's every chance the upgrade is still going
		// and the API will come back.
		return false, fmt.Sprintf("error getting UpgradeConfig: %v", err), nil
	}
	if len(ucList.Items) < 1 {
		// The API returned successfully, but thinks there's no UpgradeConfigs.
		// This is an unexpected state to be in, so let's verify with the provider
		// that it thinks there's an upgrade going on.
		clusterID := viper.GetString(config.Cluster.ID)
		clusterProvider, err := providers.ClusterProvider()
		if err != nil {
			// We shouldn't be upgrading if there's no cluster provider
			return false, "", fmt.Errorf("error getting clusterprovider for upgrade: %v", err)
		}
		policyID, err := clusterProvider.GetUpgradePolicyID(clusterID)
		if err != nil {
			// The provider errored when attempting to query its upgrade policies; try again next time
			return false, fmt.Sprintf("couldn't check upgrade policy state from provider: %v", err), nil
		}
		if policyID != "" {
			// The provider has an upgrade policy. MUO might be waiting to sync it into an upgradeconfig
			return false, fmt.Sprintf("cluster provider has an upgrade policy, but cluster has no upgradeconfig yet"), nil
		}
		if policyID == "" {
			// The provider successfully returned that it contains no upgrade policies.
			// If we're in this state with no UC and no policy, the upgrade must have either
			// failed or been cancelled. Either way, there's no point monitoring for the upgrade any more.
			return true, "", fmt.Errorf("the provider no longer has an upgrade policy, the upgrade has been cancelled or failed")
		}
	}

	// We have an UpgradeConfig on-cluster, so we can use that to reliably tell what the upgrade state is
	ucObj := ucList.Items[0]
	var upgradeConfig upgradev1alpha1.UpgradeConfig
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ucObj.UnstructuredContent(), &upgradeConfig)
	if err != nil {
		// This, however, is probably error-worthy because it means our UpgradeConfig
		// has been messed with or something odd's occurred
		return false, "", fmt.Errorf("error parsing upgradeconfig into object")
	}

	// First attempt any necessary UpgradeConfig overrides we need to do during our regular sync/check
	err = overrideUpgradeConfig(upgradeConfig, h)
	if err != nil {
		// log it, but this isn't a problem worth failing out over
		log.Printf("could not apply UpgradeConfig overrides: %v", err)
	}

	upgradeHistory := upgradeConfig.Status.History.GetHistory(upgradeConfig.Spec.Desired.Version)

	// If the Upgrade History is empty, the upgrade hasn't started, try again later
	if upgradeHistory == nil {
		return false, fmt.Sprintf("upgrade yet to commence"), nil
	}

	// If the Upgrade History phase indicates the upgrade has failed, we can bail out here
	if upgradeHistory.Phase == upgradev1alpha1.UpgradePhaseFailed {
		return true, "", fmt.Errorf("managed-upgrade-operator has indicated upgrade failed to commence")
	}

	// If the UpgradeHistory phase indicates the upgrade has completed, we can also bail out here (successfully)
	if upgradeHistory.Phase == upgradev1alpha1.UpgradePhaseUpgraded {
		return true, "", nil
	}

	// Otherwise, report back what current phase and upgrade condition the upgrade is at
	upgradeConditions := upgradeHistory.Conditions
	if len(upgradeConditions) == 0 {
		return false, fmt.Sprintf("current upgrade status is pending"), nil
	}
	statusMsg := upgradeConditions[0].Message
	statusTimeStr := "unknown"
	statusTime := upgradeHistory.StartTime
	if statusTime != nil {
		statusTimeStr = upgradeHistory.StartTime.String()
	}

	return false, fmt.Sprintf(`current upgrade status is "%s" since "%s"`, statusMsg, statusTimeStr), nil
}

// Requests a cluster upgrade from the cluster provider
func scheduleUpgradeWithProvider(h *helper.H) error {
	switch getProviderSource() {
	case providerOCM:
		return scheduleOCMUpgrade()
	case providerLocal:
		return scheduleLocalUpgrade(h)
	default:
		return fmt.Errorf("unable to determine upgrade policy provider to schedule upgrade")
	}
}

// Reschedule the upgrade via the provider
func updateUpgradeWithProvider() error {
	// Only supported by OCM
	if getProviderSource() == providerLocal {
		return nil
	}

	releaseName := viper.GetString(config.Upgrade.ReleaseName)
	upgradeVersion, err := util.OpenshiftVersionToSemver(releaseName)
	if err != nil {
		return fmt.Errorf("error parsing releasename into semver: %v", err)
	}

	clusterID := viper.GetString(config.Cluster.ID)
	clusterProvider, err := providers.ClusterProvider()
	if err != nil {
		return fmt.Errorf("error getting clusterprovider for upgrade: %v", err)
	}
	policyID, err := clusterProvider.GetUpgradePolicyID(clusterID)
	if err != nil {
		return fmt.Errorf("unable to retrieve upgrade policy ID to update: %v", err)
	}
	if policyID == "" {
		return fmt.Errorf("no upgrade policy exists on the cluster to update")
	}

	newT := time.Now().UTC().Add(300 * time.Minute)
	err = clusterProvider.UpdateSchedule(clusterID, upgradeVersion.String(), newT, policyID)
	if err != nil {
		return fmt.Errorf("error updating the upgrade schedule via provider: %v", err)
	}

	return nil
}

// Scales down and scales up the operator deployment to initiate a pod restart
func restartOperator(h *helper.H, ns string) error {
	log.Printf("restarting managed-upgrade-operator to force upgrade resync..")

	err := wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		// scale down
		s, err := h.Kube().AppsV1().Deployments(ns).GetScale(context.TODO(), "managed-upgrade-operator", metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		sc := *s
		sc.Spec.Replicas = 0
		_, err = h.Kube().AppsV1().Deployments(ns).UpdateScale(context.TODO(), "managed-upgrade-operator", &sc, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}

		// scale up
		s, err = h.Kube().AppsV1().Deployments(ns).GetScale(context.TODO(), "managed-upgrade-operator", metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		sc = *s
		sc.Spec.Replicas = 1
		_, err = h.Kube().AppsV1().Deployments(ns).UpdateScale(context.TODO(), "managed-upgrade-operator", &sc, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}
		log.Printf("managed-upgrade-operator restart complete..")
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("couldn't restart managed-upgrade-operator for config re-sync: %v", err)
	}
	return nil
}

// this makes sure that the upgradeconfig has been synced from provider to the cluster
func isUpgradeConfigCreated(h *helper.H) (bool, error) {
	ucList, err := h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
	}).Namespace(muoNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	if len(ucList.Items) < 1 {
		return false, nil
	}
	return true, nil
}

// check the upgradeconfig status to determine if the upgrade is started
func isUpgradeTriggered(h *helper.H, desired *configv1.Update) (bool, error) {
	// retrieve UpgradeConfig
	ucObj, err := h.Dynamic().Resource(schema.GroupVersionResource{
		Group: "upgrade.managed.openshift.io", Version: "v1alpha1", Resource: "upgradeconfigs",
	}).Namespace(muoNamespace).Get(context.TODO(), upgradeConfigName, metav1.GetOptions{})
	if err != nil {
		// The API may sometimes be unavailable, this is fine. We don't want
		// to return an error because there's every chance the upgrade is still going
		// and the API will come back.
		return false, fmt.Errorf("failed to get the upgrade config: %v", err)
	}

	var upgradeConfig upgradev1alpha1.UpgradeConfig
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ucObj.UnstructuredContent(), &upgradeConfig)
	if err != nil {
		// This, however, is probably error-worthy because it means our UpgradeConfig
		// has been messed with or something odd's occurred
		return false, err
	}

	// Check if the Upgrade is trigger
	upgradeHistory := upgradeConfig.Status.History.GetHistory(desired.Version)
	if upgradeHistory == nil {
		return false, fmt.Errorf("upgrade has not been scheduled")
	}
	if upgradeHistory.Phase != "Pending" {
		return true, fmt.Errorf("cluster upgrade has been triggered")
	}

	return false, nil
}

// create cluster workloads prior to upgrade commencement which will test
// manage-upgrade-operator's drain strategy functions
func createUpgradeClusterWorkloads(h *helper.H) error {
	// Create Pod Disruption Budget test workloads if desired
	if viper.GetBool(config.Upgrade.ManagedUpgradeTestPodDisruptionBudgets) {
		pdbPodPrefixes := []string{"pdb"}
		err := createManagedUpgradeWorkload(pdbWorkloadName, pdbWorkloadDir, pdbPodPrefixes, h)
		if err != nil {
			return fmt.Errorf("unable to setup PDB workload for upgrade: %v", err)
		}
	}

	// Create Node Drain test workloads if desired
	if viper.GetBool(config.Upgrade.ManagedUpgradeTestNodeDrain) {
		drainPodPrefixes := []string{"node-drain-test"}
		err := createManagedUpgradeWorkload(drainWorkloadName, drainWorkloadDir, drainPodPrefixes, h)
		if err != nil {
			return fmt.Errorf("unable to setup node drain test workload for upgrade: %v", err)
		}
	}

	return nil
}

// determine which policy provider to use for an upgrade policy source
func getProviderSource() string {
	// If upgrading using an image, then the provider must be LOCAL
	if viper.GetString(config.Upgrade.Image) != "" {
		return providerLocal
	}
	return providerOCM
}

func scheduleOCMUpgrade() error {
	releaseName := viper.GetString(config.Upgrade.ReleaseName)
	if releaseName == "" {
		return fmt.Errorf("missing release name used for upgrade")
	}
	upgradeVersion, err := util.OpenshiftVersionToSemver(releaseName)
	if err != nil {
		return fmt.Errorf("unable to semantic-version parse release name: %v", releaseName)
	}
	clusterID := viper.GetString(config.Cluster.ID)
	clusterProvider, err := providers.ClusterProvider()
	if err != nil {
		return fmt.Errorf("error getting clusterprovider for upgrade: %v", err)
	}

	// Our time will be as closely allowed as possible by the provider (now + 7 min)
	t := time.Now().UTC().Add(7 * time.Minute)

	err = clusterProvider.Upgrade(clusterID, upgradeVersion.String(), t)
	if err != nil {
		return fmt.Errorf("error initiating upgrade from provider: %v", err)
	}

	return nil
}

func scheduleLocalUpgrade(h *helper.H) error {
	// Validate the upgrade type is supported
	var upgradeType upgradev1alpha1.UpgradeType
	switch viper.GetString(config.Upgrade.Type) {
	case "OSD":
		upgradeType = upgradev1alpha1.OSD
	case "ARO":
		upgradeType = upgradev1alpha1.ARO
	default:
		return fmt.Errorf("unsupported upgrade type: %v", viper.GetString(config.Upgrade.Type))
	}

	localUpgradeConfig := upgradev1alpha1.UpgradeConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UpgradeConfig",
			APIVersion: "upgrade.managed.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "managed-upgrade-config",
			Namespace: muoNamespace,
		},
		Spec: upgradev1alpha1.UpgradeConfigSpec{
			Desired: upgradev1alpha1.Update{
				Image: viper.GetString(config.Upgrade.Image),
			},
			UpgradeAt:            time.Now().UTC().Format(time.RFC3339),
			PDBForceDrainTimeout: configPdbDrainTimeoutOverride,
			Type:                 upgradeType,
			CapacityReservation:  true,
		},
		Status: upgradev1alpha1.UpgradeConfigStatus{},
	}

	// Create the UpgradeConfig on-cluster
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(localUpgradeConfig.DeepCopy())
	if err != nil {
		return fmt.Errorf("can't convert UpgradeConfig to unstructured resource: %v", err)
	}
	uobj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "upgrade.managed.openshift.io",
		Version:  "v1alpha1",
		Resource: "upgradeconfigs",
	}).Namespace(muoNamespace).Create(context.TODO(), &uobj, metav1.CreateOptions{})

	return err
}
