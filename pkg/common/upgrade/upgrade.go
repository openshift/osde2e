// Package upgrade provides utilities to per
package upgrade

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/spf13/viper"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
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
	MaxDuration = 90 * time.Minute
)

// RunUpgrade uses the OpenShift extended suite to upgrade a cluster to the image provided in cfg.
func RunUpgrade(provider spi.Provider) error {
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

	desired, err := TriggerUpgrade(h)
	if err != nil {
		return fmt.Errorf("failed triggering upgrade: %v", err)
	}
	log.Println("Cluster acknowledged update request.")

	log.Println("Upgrading...")
	done = false
	if err = wait.PollImmediate(10*time.Second, MaxDuration, func() (bool, error) {
		done, msg, err = IsUpgradeDone(h, desired.Spec.DesiredUpdate)
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

	if err = cluster.WaitForClusterReady(provider, viper.GetString(config.Cluster.ID)); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	log.Println("Upgrade complete!")
	return nil
}

// TriggerUpgrade uses a helper to perform an upgrade.
func TriggerUpgrade(h *helper.H) (*configv1.ClusterVersion, error) {
	var cVersion *configv1.ClusterVersion
	var err error
	// setup Config client
	cfgClient := h.Cfg()

	// get current Version
	getOpts := metav1.GetOptions{}
	cVersion, err = cfgClient.ConfigV1().ClusterVersions().Get(context.TODO(), ClusterVersionName, getOpts)
	if err != nil {
		return cVersion, fmt.Errorf("couldn't get current ClusterVersion '%s': %v", ClusterVersionName, err)
	}

	image := viper.GetString(config.Upgrade.Image)
	releaseName := viper.GetString(config.Upgrade.ReleaseName)

	// set requested upgrade targets
	if image != "" {
		cVersion.Spec.DesiredUpdate = &configv1.Update{
			Version: strings.Replace(releaseName, "openshift-v", "", -1),
			Image:   image,
			Force:   true, // Force if we have an image specified
		}
	} else {
		upgradeVersion := strings.Replace(releaseName, "openshift-v", "", -1)
		installVersion := strings.Replace(viper.GetString(config.Cluster.Version), "openshift-v", "", -1)

		upgradeVersionParsed := semver.MustParse(upgradeVersion)
		installVersionParsed := semver.MustParse(installVersion)

		if upgradeVersionParsed.GreaterThan(installVersionParsed) {
			cVersion.Spec.Channel, err = VersionToChannel(upgradeVersionParsed)
			if err != nil {
				return cVersion, fmt.Errorf("unable to channel from version: %v", err)
			}

			cVersion, err = cfgClient.ConfigV1().ClusterVersions().Update(context.TODO(), cVersion, metav1.UpdateOptions{})
			if err != nil {
				return cVersion, fmt.Errorf("couldn't update desired release channel: %v", err)
			}

			// https://github.com/openshift/managed-cluster-config/blob/master/scripts/cluster-upgrade.sh#L258
			time.Sleep(15 * time.Second)

			cVersion, err = cfgClient.ConfigV1().ClusterVersions().Get(context.TODO(), ClusterVersionName, getOpts)
			if err != nil {
				return cVersion, fmt.Errorf("couldn't get current ClusterVersion '%s' after updating release channel: %v", ClusterVersionName, err)
			}
		}

		// Assume CIS has all the information required. Just pass version info.
		cVersion.Spec.DesiredUpdate = &configv1.Update{
			Version: strings.Replace(releaseName, "openshift-v", "", -1),
		}
	}

	updatedCV, err := cfgClient.ConfigV1().ClusterVersions().Update(context.TODO(), cVersion, metav1.UpdateOptions{})
	if err != nil {
		return updatedCV, fmt.Errorf("couldn't update desired ClusterVersion: %v", err)
	}

	// wait for update acknowledgement
	updateGeneration := updatedCV.Generation
	if err = wait.PollImmediate(15*time.Second, 5*time.Minute, func() (bool, error) {
		if cVersion, err = cfgClient.ConfigV1().ClusterVersions().Get(context.TODO(), ClusterVersionName, getOpts); err != nil {
			return false, err
		}
		return cVersion.Status.ObservedGeneration >= updateGeneration, nil
	}); err != nil {
		return updatedCV, fmt.Errorf("cluster did not acknowledge update in a timely manner: %v", err)
	}

	return updatedCV, nil
}

// IsUpgradeDone returns with done true when an upgrade is complete at desired and any available msg.
func IsUpgradeDone(h *helper.H, desired *configv1.Update) (done bool, msg string, err error) {
	// retrieve current ClusterVersion
	cfgClient, getOpts := h.Cfg(), metav1.GetOptions{}
	cVersion, err := cfgClient.ConfigV1().ClusterVersions().Get(context.TODO(), ClusterVersionName, getOpts)
	if err != nil {
		log.Printf("error getting ClusterVersion '%s': %v", ClusterVersionName, err)
	}

	// ensure working towards correct desired
	curDesired := cVersion.Status.Desired
	if curDesired.Version != desired.Version {
		return false, fmt.Sprintf("desired not yet updated; desired: %v, cur: %v", desired.Version, curDesired.Version), nil
	}

	// check if any ActiveConditions indicate an upgrade is ongoing
	for _, aCondition := range ActiveConditions {
		for _, c := range cVersion.Status.Conditions {
			if c.Type == aCondition && c.Status == configv1.ConditionTrue {
				return false, c.Message, nil
			}
		}
	}

	// check that latest history entry is desired and completed
	if len(cVersion.Status.History) > 0 {
		latest := &cVersion.Status.History[0]
		if latest == nil || latest.State != configv1.CompletedUpdate {
			return false, "history doesn't have a completed update", nil
		} else if latest.Version != desired.Version {
			return false, fmt.Sprintf("latest in history doesn't match desired; desired: %v, cur: %v", desired, latest), nil
		}
	}

	done = true
	return
}

// VersionToChannel creates a Cincinnati channel version out of an OpenShift version.
// If the config.Instance.Upgrade.OnlyUpgradeToZReleases flag is set, this will use the install version
// in the global state object to determine the channel.
// The provider will be queried for the appropriate Cincinnati channel  to use unless a prelease version
// is being used, in which case the candidate channel will be used.
func VersionToChannel(version *semver.Version) (string, error) {
	useVersion := version
	if viper.GetBool(config.Upgrade.OnlyUpgradeToZReleases) {
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
