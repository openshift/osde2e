// Package upgrade provides utilities to per
package upgrade

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Masterminds/semver"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/osd"
	"github.com/openshift/osde2e/pkg/common/state"
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
func RunUpgrade(OSD *osd.OSD) error {
	var done bool
	var msg string
	var err error
	var upgradeStarted time.Time

	// setup helper
	h := &helper.H{
		State: state.Instance,
	}
	h.SetupNoProj()
	defer h.Cleanup()

	if h.Upgrade.Image != "" {
		log.Printf("Upgrading cluster to UPGRADE_IMAGE '%s'", h.Upgrade.Image)
	} else {
		log.Printf("Upgrading cluster to cluster image set with version %s", h.Upgrade.ReleaseName)
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

	if err = OSD.WaitForClusterReady(); err != nil {
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
	cVersion, err = cfgClient.ConfigV1().ClusterVersions().Get(ClusterVersionName, getOpts)
	if err != nil {
		return cVersion, fmt.Errorf("couldn't get current ClusterVersion '%s': %v", ClusterVersionName, err)
	}

	// set requested upgrade targets
	if h.Upgrade.Image != "" {
		cVersion.Spec.DesiredUpdate = &configv1.Update{
			Version: strings.Replace(h.Upgrade.ReleaseName, "openshift-v", "", -1),
			Image:   h.Upgrade.Image,
			Force:   true, // Force if we have an image specified
		}
	} else {
		upgradeVersion := strings.Replace(h.Upgrade.ReleaseName, "openshift-v", "", -1)
		installVersion := strings.Replace(state.Instance.Cluster.Version, "openshift-v", "", -1)

		upgradeVersionParsed := semver.MustParse(upgradeVersion)
		installVersionParsed := semver.MustParse(installVersion)

		if upgradeVersionParsed.GreaterThan(installVersionParsed) {
			cVersion.Spec.Channel = VersionToChannel(upgradeVersionParsed)
			// Upgrade the channel
			if strings.HasPrefix(upgradeVersionParsed.Prerelease(), "rc") {
				cVersion.Spec.Channel = fmt.Sprintf("candidate-%d.%d", upgradeVersionParsed.Major(), upgradeVersionParsed.Minor())
			} else {
				cVersion.Spec.Channel = fmt.Sprintf("fast-%d.%d", upgradeVersionParsed.Major(), upgradeVersionParsed.Minor())
			}
			cVersion, err = cfgClient.ConfigV1().ClusterVersions().Update(cVersion)
			if err != nil {
				return cVersion, fmt.Errorf("couldn't update desired release channel: %v", err)
			}

			// https://github.com/openshift/managed-cluster-config/blob/master/scripts/cluster-upgrade.sh#L258
			time.Sleep(15 * time.Second)

			cVersion, err = cfgClient.ConfigV1().ClusterVersions().Get(ClusterVersionName, getOpts)
			if err != nil {
				return cVersion, fmt.Errorf("couldn't get current ClusterVersion '%s' after updating release channel: %v", ClusterVersionName, err)
			}
		}

		// Assume CIS has all the information required. Just pass version info.
		cVersion.Spec.DesiredUpdate = &configv1.Update{
			Version: strings.Replace(h.Upgrade.ReleaseName, "openshift-v", "", -1),
		}
	}

	updatedCV, err := cfgClient.ConfigV1().ClusterVersions().Update(cVersion)
	if err != nil {
		return updatedCV, fmt.Errorf("couldn't update desired ClusterVersion: %v", err)
	}

	// wait for update acknowledgement
	updateGeneration := updatedCV.Generation
	if err = wait.PollImmediate(15*time.Second, 5*time.Minute, func() (bool, error) {
		if cVersion, err = cfgClient.ConfigV1().ClusterVersions().Get(ClusterVersionName, getOpts); err != nil {
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
	cVersion, err := cfgClient.ConfigV1().ClusterVersions().Get(ClusterVersionName, getOpts)
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
