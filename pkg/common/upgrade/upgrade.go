// Package upgrade provides utilities to per
package upgrade

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/util"
)

const (
	// ClusterVersionName is used to identify the default ClusterVersion.
	ClusterVersionName = "version"
)

var (
	// ActiveConditions have true statuses when an upgrade is ongoing.
	ActiveConditions = []configv1.ClusterStatusConditionType{
		configv1.OperatorProgressing,
		configv1.OperatorDegraded,
		configv1.ClusterStatusConditionType("Failing"),
	}

	// MaxDuration is how long an upgrade will run before failing.
	MaxDuration = 180 * time.Minute
)

// RunUpgrade uses the OpenShift extended suite to upgrade a cluster to the image provided in cfg.
func RunUpgrade() error {
	var done bool
	var msg string
	var err error
	var upgradeStarted time.Time

	// setup helper
	h := helper.NewOutsideGinkgo()
	if h == nil {
		return fmt.Errorf("Unable to generate helper outside ginkgo")
	}

	image := viper.GetString(config.Upgrade.Image)
	if image != "" {
		log.Printf("Upgrading cluster to UPGRADE_IMAGE '%s'", image)
	} else {
		log.Printf("Upgrading cluster to cluster image set with version %s", viper.GetString(config.Upgrade.ReleaseName))
	}

	upgradeStarted = time.Now()

	var desiredUpdate *configv1.Update

	// Check we are on a supported provider
	provider, err := providers.ClusterProvider()
	if err != nil {
		return fmt.Errorf("can't determine provider for managed upgrade: %s", err)
	}
	switch provider.Type() {
	case "rosa":
		fallthrough
	case "ocm":
		desiredUpdate, err = TriggerManagedUpgrade(h)
		if err != nil {
			return fmt.Errorf("failed triggering upgrade: %v", err)
		}
	default:
		return fmt.Errorf("unsupported provider for managed upgrades (%s)", provider.Type())
	}

	// When the upgrade being rescheduled, we should expect that the upgrade will not be triggered
	if viper.GetBool(config.Upgrade.ManagedUpgradeRescheduled) {
		time.Sleep(10 * time.Minute)
		triggered, err := isUpgradeTriggered(h, desiredUpdate)
		if triggered {
			return fmt.Errorf("the upgrade was triggered unexpectly: %v", err)
		} else {
			log.Println("Upgrade has been rescheduled/cancelled")
			return nil
		}
	}

	log.Println("Cluster acknowledged update request.")

	log.Println("Upgrading...")
	done = false
	if err = wait.PollImmediate(10*time.Second, MaxDuration, func() (bool, error) {
		// Keep the managed upgrade's configuration overrides in place, in case Hive has replaced them
		err = overrideOperatorConfig(h)
		// Log if it errored, but don't cancel the upgrade because of it
		if err != nil {
			log.Printf("problem overriding managed upgrade config: %v", err)
		}
		// If performing a managed upgrade, check if we want to wait for workers to fully upgrade too
		done, msg, err = isManagedUpgradeDone(h)

		if !done {
			log.Printf("Upgrade in progress: %s", msg)
		}
		return done, err
	}); err != nil {
		return fmt.Errorf("failed to upgrade cluster: %v", err)
	}

	if !done {
		return fmt.Errorf("failed to upgrade cluster: timed out after %d min waiting for upgrade", MaxDuration)
	}

	metadata.Instance.SetTimeToUpgradedCluster(time.Since(upgradeStarted).Seconds())

	if err = cluster.WaitForClusterReadyPostUpgrade(viper.GetString(config.Cluster.ID), nil); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	log.Println("Upgrade complete!")
	if viper.GetBool(config.Upgrade.ManagedUpgradeTestNodeDrain) {
		list, err := h.Kube().CoreV1().Pods(h.CurrentProject()).List(context.TODO(), metav1.ListOptions{LabelSelector: "app=node-drain-test"})
		if err != nil {
			return fmt.Errorf("Error listing pods: %s", err.Error())
		}
		if len(list.Items) != 0 {
			for _, item := range list.Items {
				log.Printf("Removing finalizers from %s", item.Name)
				item.Finalizers = []string{}
				h.Kube().CoreV1().Pods(h.CurrentProject()).Update(context.TODO(), &item, metav1.UpdateOptions{})
				log.Printf("Deleting pod %s", item.Name)
				h.Kube().CoreV1().Pods(h.CurrentProject()).Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
			}
		}
	}
	return nil
}

// VersionToChannel creates a Cincinnati channel version out of an OpenShift version.
// If the config.Instance.Upgrade.OnlyUpgradeToZReleases flag is set, this will use the install version
// in the global state object to determine the channel.
// The provider will be queried for the appropriate Cincinnati channel  to use unless a prelease version
// is being used, in which case the candidate channel will be used.
func VersionToChannel(version *semver.Version) (string, error) {
	useVersion := version
	if viper.GetBool(config.Upgrade.UpgradeToLatestZ) {
		var err error
		useVersion, err = util.OpenshiftVersionToSemver(viper.GetString(config.Cluster.Version))

		if err != nil {
			panic("cluster version stored in state object is invalid")
		}
	}

	if strings.HasPrefix(useVersion.Prerelease(), "rc") {
		return fmt.Sprintf("candidate-%d.%d", useVersion.Major(), useVersion.Minor()), nil
	}

	provider, err := providers.ClusterProvider()
	if err != nil {
		return "", fmt.Errorf("unable to get provider: %s", err)
	}

	return fmt.Sprintf("%s-%d.%d", provider.CincinnatiChannel(), useVersion.Major(), useVersion.Minor()), nil
}
