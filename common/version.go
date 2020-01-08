package common

import (
	"errors"
	"fmt"
	"log"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/osd"
	"github.com/openshift/osde2e/pkg/upgrade"
)

// ChooseVersions sets versions in cfg if not set based on defaults and upgrade options.
// If a release stream is set for an upgrade the previous available version is used and it's image is used for upgrade.
func ChooseVersions(cfg *config.Config, osd *osd.OSD) (err error) {
	// when defined, use set version
	if len(cfg.Cluster.Version) != 0 {
		err = nil
	} else if osd == nil {
		err = errors.New("osd must be setup when upgrading with release stream")
	} else if cfg.Upgrade.Image == "" && cfg.Upgrade.ReleaseStream != "" {
		err = setupUpgradeVersion(cfg, osd)
	} else {
		err = setupVersion(cfg, osd)
	}

	// Set the versions in metadata. If upgrade hasn't been chosen, it should still be omitted from the end result.
	metadata.Instance.ClusterVersion = cfg.Cluster.Version
	metadata.Instance.UpgradeVersion = cfg.Upgrade.ReleaseName

	return err
}

// chooses between default version and nightly based on target versions.
func setupVersion(cfg *config.Config, osd *osd.OSD) (err error) {
	if len(cfg.Cluster.Version) > 0 {
		return
	}
	if cfg.Upgrade.MajorTarget != 0 || cfg.Upgrade.MinorTarget != 0 {
		// don't require major to be set
		if cfg.Upgrade.MajorTarget == 0 {
			cfg.Upgrade.MajorTarget = -1
		}
		// look for the latest release and install it for this OSD cluster.
		if cfg.Cluster.Version, err = osd.LatestVersion(cfg.Upgrade.MajorTarget, cfg.Upgrade.MinorTarget); err == nil {
			log.Printf("CLUSTER_VERSION not set but a TARGET is, running '%s'", cfg.Cluster.Version)
		}
	}

	if len(cfg.Cluster.Version) == 0 {
		if cfg.Cluster.Version, err = osd.DefaultVersion(); err == nil {
			log.Printf("CLUSTER_VERSION not set, using the current default '%s'", cfg.Cluster.Version)
		} else {
			return fmt.Errorf("Error finding default cluster version: %v", err)
		}
	}

	return
}

// chooses version based on optimal upgrade path
func setupUpgradeVersion(cfg *config.Config, osd *osd.OSD) (err error) {
	// Decide the version to install
	err = setupVersion(cfg, osd)
	if err != nil {
		return err
	}

	clusterVersion, err := osd.OpenshiftVersionToSemver(cfg.Cluster.Version)
	if err != nil {
		log.Printf("error while parsing cluster version %s: %v", cfg.Cluster.Version, err)
		return err
	}

	if cfg.Upgrade.UpgradeToCISIfPossible {
		cisUpgradeVersionString, err := osd.LatestVersion(-1, -1)

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
		if cisUpgradeVersion.GreaterThan(clusterVersion) {
			log.Printf("Using cluster image set.")
			cfg.Upgrade.ReleaseName = cisUpgradeVersionString
			metadata.Instance.UpgradeVersionSource = "cluster image set"
			log.Printf("Selecting version '%s' to be able to upgrade to '%s'", cfg.Cluster.Version, cfg.Upgrade.ReleaseName)
			return nil
		}

		log.Printf("The most recent cluster image set is equal to the default. Falling back to upgrading with Cincinnati.")
	}

	cfg.Upgrade.ReleaseName, cfg.Upgrade.Image, err = upgrade.LatestRelease(cfg, cfg.Upgrade.ReleaseStream, true)
	if err != nil {
		return fmt.Errorf("couldn't get latest release from release-controller: %v", err)
	}

	upgradeVersion, err := osd.OpenshiftVersionToSemver(cfg.Upgrade.ReleaseName)
	if err != nil {
		log.Printf("error while parsing upgrade version %s: %v", cfg.Upgrade.ReleaseName, err)
		return err
	}

	if !clusterVersion.LessThan(upgradeVersion) {
		log.Printf("Cluster version is equal to or newer than the upgrade version. Looking up previous version...")
		if cfg.Cluster.Version, err = osd.PreviousVersion(cfg.Upgrade.ReleaseName); err != nil {
			return fmt.Errorf("failed retrieving previous version to '%s': %v", cfg.Upgrade.ReleaseName, err)
		}
	}

	// set upgrade image
	log.Printf("Selecting version '%s' to be able to upgrade to '%s' on release stream '%s'",
		cfg.Cluster.Version, cfg.Upgrade.ReleaseName, cfg.Upgrade.ReleaseStream)
	return
}
