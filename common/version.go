package common

import (
	"errors"
	"fmt"
	"log"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/osd"
	"github.com/openshift/osde2e/pkg/state"
	"github.com/openshift/osde2e/pkg/upgrade"
)

// ChooseVersions sets versions in cfg if not set based on defaults and upgrade options.
// If a release stream is set for an upgrade the previous available version is used and it's image is used for upgrade.
func ChooseVersions(osd *osd.OSD) (err error) {
	cfg := config.Instance
	state := state.Instance

	// when defined, use set version
	if len(state.Cluster.Version) != 0 {
		err = nil
	} else if osd == nil {
		err = errors.New("osd must be setup when upgrading with release stream")
	} else if state.Upgrade.Image == "" && (cfg.Upgrade.ReleaseStream != "" || cfg.Upgrade.UpgradeToCISIfPossible) {
		err = setupUpgradeVersion(osd)
	} else {
		err = setupVersion(osd)
	}

	// Set the versions in metadata. If upgrade hasn't been chosen, it should still be omitted from the end result.
	metadata.Instance.SetClusterVersion(state.Cluster.Version)
	metadata.Instance.SetUpgradeVersion(state.Upgrade.ReleaseName)

	return err
}

// chooses between default version and nightly based on target versions.
func setupVersion(osd *osd.OSD) (err error) {
	cfg := config.Instance
	state := state.Instance
	suffix := ""

	if len(state.Cluster.Version) == 0 && (cfg.Upgrade.MajorTarget != 0 || cfg.Upgrade.MinorTarget != 0) {
		majorTarget := cfg.Upgrade.MajorTarget
		// don't require major to be set
		if majorTarget == 0 {
			majorTarget = -1
		}

		if cfg.OCM.Env == "int" && cfg.Upgrade.ReleaseStream == "" {
			suffix = "nightly"
		}

		// look for the latest release and install it for this OSD cluster.
		if state.Cluster.Version, err = osd.LatestVersion(majorTarget, cfg.Upgrade.MinorTarget, suffix); err == nil {
			log.Printf("CLUSTER_VERSION not set but a TARGET is, running '%s'", state.Cluster.Version)
		}
	}

	if len(state.Cluster.Version) == 0 {
		var err error
		var versionType string
		if cfg.Upgrade.UseLatestVersionForInstall {
			state.Cluster.Version, err = osd.LatestVersion(-1, -1, "")
			versionType = "latest version"
		} else {
			state.Cluster.Version, err = osd.DefaultVersion()
			versionType = "current default"
		}

		if err == nil {
			log.Printf("CLUSTER_VERSION not set, using the %s '%s'", versionType, state.Cluster.Version)
		} else {
			return fmt.Errorf("Error finding default cluster version: %v", err)
		}
	}

	return
}

// chooses version based on optimal upgrade path
func setupUpgradeVersion(osd *osd.OSD) (err error) {
	cfg := config.Instance
	state := state.Instance

	// Decide the version to install
	err = setupVersion(osd)
	if err != nil {
		return err
	}

	clusterVersion, err := osd.OpenshiftVersionToSemver(state.Cluster.Version)
	if err != nil {
		log.Printf("error while parsing cluster version %s: %v", state.Cluster.Version, err)
		return err
	}

	if cfg.Upgrade.UpgradeToCISIfPossible {
		suffix := ""
		if cfg.OCM.Env == "int" {
			suffix = "nightly"
		}
		cisUpgradeVersionString, err := osd.LatestVersion(-1, -1, suffix)

		if err != nil {
			log.Printf("unable to get the most recent version of openshift from OSD: %v", err)
			return err
		}

		cisUpgradeVersion, err := osd.OpenshiftVersionToSemver(cisUpgradeVersionString)

		if err != nil {
			log.Printf("unable to parse most recent version of openshift from OSD: %v", err)
			return err
		}

		// If the available cluster image set makes sense, then we'll just use that
		if !cisUpgradeVersion.LessThan(clusterVersion) {
			log.Printf("Using cluster image set.")
			state.Upgrade.ReleaseName = cisUpgradeVersionString
			metadata.Instance.SetUpgradeVersionSource("cluster image set")
			state.Upgrade.UpgradeVersionEqualToInstallVersion = cisUpgradeVersion.Equal(clusterVersion)
			log.Printf("Selecting version '%s' to be able to upgrade to '%s'", state.Cluster.Version, state.Upgrade.ReleaseName)
			return nil
		}

		if state.Upgrade.ReleaseName != "" {
			log.Printf("The most recent cluster image set is equal to the default. Falling back to upgrading with Cincinnati.")
		} else {
			return fmt.Errorf("couldn't get latest cluster image set release and no Cincinnati fallback")
		}
	}

	state.Upgrade.ReleaseName, state.Upgrade.Image, err = upgrade.LatestRelease(cfg.Upgrade.ReleaseStream, true)
	if err != nil {
		return fmt.Errorf("couldn't get latest release from release-controller: %v", err)
	}

	upgradeVersion, err := osd.OpenshiftVersionToSemver(state.Upgrade.ReleaseName)
	if err != nil {
		log.Printf("error while parsing upgrade version %s: %v", state.Upgrade.ReleaseName, err)
		return err
	}

	if !clusterVersion.LessThan(upgradeVersion) {
		log.Printf("Cluster version is equal to or newer than the upgrade version. Looking up previous version...")
		if state.Cluster.Version, err = osd.PreviousVersion(state.Upgrade.ReleaseName); err != nil {
			return fmt.Errorf("failed retrieving previous version to '%s': %v", state.Upgrade.ReleaseName, err)
		}
	}

	// set upgrade image
	log.Printf("Selecting version '%s' to be able to upgrade to '%s' on release stream '%s'",
		state.Cluster.Version, state.Upgrade.ReleaseName, cfg.Upgrade.ReleaseStream)
	return
}

// GetEnabledNoDefaultVersions returns a sorted list of the enabled but not default versions currently offered by OSD
func GetEnabledNoDefaultVersions() ([]string, error) {
	cfg := config.Instance

	OSD, err := osd.New(cfg.OCM.Token, cfg.OCM.Env, cfg.OCM.Debug)
	if err != nil {
		return nil, fmt.Errorf("could not setup OSD: %v", err)
	}

	return OSD.EnabledNoDefaultVersionList()
}
