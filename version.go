package osde2e

import (
	"errors"
	"fmt"
	"log"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/osd"
	"github.com/openshift/osde2e/pkg/upgrade"
)

// ChooseVersions sets versions in cfg if not set based on defaults and upgrade options.
// If a release stream is set for an upgrade the previous available version is used and it's image is used for upgrade.
func ChooseVersions(cfg *config.Config, osd *osd.OSD) (err error) {
	// when defined, use set version
	if len(cfg.ClusterVersion) != 0 {
		return nil
	} else if osd == nil {
		return errors.New("osd must be setup when upgrading with release stream")
	} else if cfg.UpgradeImage == "" && cfg.UpgradeReleaseStream != "" {
		return setupUpgradeVersion(cfg, osd)
	} else {
		return setupVersion(cfg, osd, false)
	}
}

// chooses between default version and nightly based on target versions.
func setupVersion(cfg *config.Config, osd *osd.OSD, isUpgrade bool) (err error) {
	if len(cfg.ClusterVersion) > 0 {
		return
	}
	if cfg.MajorTarget == 0 && cfg.MinorTarget == 0 {
		// check to see if a target stream is set.
		// a target stream and a major/minor target should not be set at the same time.
		if len(cfg.TargetStream) > 0 {
			if cfg.ClusterVersion, _, err = upgrade.LatestRelease(cfg, cfg.TargetStream, !isUpgrade); err == nil {
				log.Printf("Target Release Stream set, using version  '%s'", cfg.ClusterVersion)
				// use defaults if no version targets
			} else {
				return fmt.Errorf("Couldn't get a valid release from stream %v", err)
			}
		}
	} else {
		// don't require major to be set
		if cfg.MajorTarget == 0 {
			cfg.MajorTarget = -1
		}
		// look for the default release and install it for this OSD cluster.
		if cfg.ClusterVersion, err = osd.LatestPrerelease(cfg.MajorTarget, cfg.MinorTarget, "nightly"); err == nil {
			log.Printf("CLUSTER_VERSION not set but a TARGET is, running nightly '%s'", cfg.ClusterVersion)
		}
	}

	if len(cfg.ClusterVersion) == 0 {
		if cfg.ClusterVersion, err = osd.DefaultVersion(); err == nil {
			log.Printf("CLUSTER_VERSION not set, using the current default '%s'", cfg.ClusterVersion)
		} else {
			return fmt.Errorf("Error finding default cluster version: %v", err)
		}
	}
	return
}

// chooses version based on optimal upgrade path
func setupUpgradeVersion(cfg *config.Config, osd *osd.OSD) (err error) {
	// Decide the version to install
	err = setupVersion(cfg, osd, true)
	if err != nil {
		return err
	}

	cfg.UpgradeReleaseName, cfg.UpgradeImage, err = upgrade.LatestRelease(cfg, cfg.UpgradeReleaseStream, true)
	if err != nil {
		return fmt.Errorf("couldn't get latest release from release-controller: %v", err)
	}

	log.Printf("Target stream: '%s', Upgrade stream: '%s'", cfg.TargetStream, cfg.UpgradeReleaseStream)

	if cfg.ClusterVersion == cfg.UpgradeReleaseName {
		log.Printf("Cluster version and target version are the same. Looking up previous version...")
		if cfg.ClusterVersion, err = osd.PreviousVersion(cfg.ClusterVersion); err != nil {
			return fmt.Errorf("failed retrieving previous version to '%s': %v", cfg.UpgradeReleaseName, err)
		}
	}

	// set upgrade image
	log.Printf("Selecting version '%s' to be able to upgrade to '%s' on release stream '%s'",
		cfg.ClusterVersion, cfg.UpgradeReleaseName, cfg.UpgradeReleaseStream)
	return
}

func buildVersion(cfg *config.Config) string {
	// use just version if not upgrading
	if cfg.UpgradeReleaseStream == "" && cfg.UpgradeImage == "" {
		return cfg.ClusterVersion
	}

	return fmt.Sprintf("%s-%s", cfg.ClusterVersion, cfg.UpgradeReleaseName)
}
