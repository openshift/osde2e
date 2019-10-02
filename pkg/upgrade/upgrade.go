// Package upgrade provides utilities to per
package upgrade

import (
	"fmt"
	"log"
	"strings"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/helper"
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
func RunUpgrade(cfg *config.Config) error {
	// setup helper
	h := &helper.H{
		Config: cfg,
	}
	h.Setup()
	defer h.Cleanup()

	log.Printf("Upgrading cluster to UPGRADE_IMAGE '%s'", cfg.UpgradeImage)
	desired, err := TriggerUpgrade(h, cfg)
	if err != nil {
		return fmt.Errorf("failed triggering upgrade: %v", err)
	}
	log.Println("Cluster acknowledged update request.")

	log.Println("Upgrading...")
	if err = wait.PollImmediate(10*time.Second, MaxDuration, func() (bool, error) {
		done, msg, err := IsUpgradeDone(h, desired.Spec.DesiredUpdate)
		if !done {
			log.Printf("Upgrade in progress: %s", msg)
		}
		return done, err
	}); err != nil {
		return fmt.Errorf("failed to upgrade cluster: %v", err)
	}
	log.Println("Upgrade complete!")
	return nil
}

// TriggerUpgrade uses a helper to perform an upgrade.
func TriggerUpgrade(h *helper.H, cfg *config.Config) (*configv1.ClusterVersion, error) {
	// setup Config client
	cfgClient := h.Cfg()

	// get current Version
	getOpts := metav1.GetOptions{}
	cVersion, err := cfgClient.ConfigV1().ClusterVersions().Get(ClusterVersionName, getOpts)
	if err != nil {
		return cVersion, fmt.Errorf("couldn't get current ClusterVersion '%s': %v", ClusterVersionName, err)
	}

	// set requested upgrade targets
	cVersion.Spec.DesiredUpdate = &configv1.Update{
		Version: strings.Replace(cfg.UpgradeReleaseName, "openshift-v", "", -1),
		Image:   cfg.UpgradeImage,
		Force:   true,
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
	if curDesired.Image != desired.Image || curDesired.Version != desired.Version {
		return false, fmt.Sprintf("desired not yet updated; desired: %v, cur: %v", desired, curDesired), nil
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
		} else if latest.Image != desired.Image || latest.Version != desired.Version {
			return false, fmt.Sprintf("latest in history doesn't match desired; desired: %v, cur: %v", desired, latest), nil
		}
	}

	done = true
	return
}
