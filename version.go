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
		return setupVersion(cfg, osd)
	}
}

// chooses between default version and nightly based on target versions.
func setupVersion(cfg *config.Config, osd *osd.OSD) (err error) {
	if cfg.MajorTarget == 0 && cfg.MinorTarget == 0 {
		// use defaults if no version targets
		if cfg.ClusterVersion, err = OSD.DefaultVersion(); err == nil {
			log.Printf("CLUSTER_VERSION not set, using the current default '%s'", cfg.ClusterVersion)
		}
	} else {
		// don't require major to be set
		if cfg.MajorTarget == 0 {
			cfg.MajorTarget = -1
		}

		if cfg.ClusterVersion, err = osd.LatestPrerelease(cfg.MajorTarget, cfg.MinorTarget, "nightly"); err == nil {
			log.Printf("CLUSTER_VERSION not set but a TARGET is, running nightly '%s'", cfg.ClusterVersion)
		}
	}
	return
}

// chooses version based on optimal upgrade path
func setupUpgradeVersion(cfg *config.Config, osd *osd.OSD) (err error) {
	cfg.UpgradeReleaseName, cfg.UpgradeImage, err = upgrade.LatestRelease(cfg, cfg.UpgradeReleaseStream)
	if err != nil {
		return fmt.Errorf("couldn't get latest release from release-controller: %v", err)
	}

	if cfg.MajorTarget == 0 && cfg.MinorTarget == 0 {
		// use defaults if no version targets
		if cfg.ClusterVersion, err = OSD.DefaultVersion(); err == nil {
			log.Printf("CLUSTER_VERSION not set, using the current default '%s'", cfg.ClusterVersion)
		}
	} else if len(cfg.TargetStream) != 0 {
		if cfg.ClusterVersion, _, err = upgrade.LatestRelease(cfg, cfg.TargetStream) err == nil {
			log.Printf("Target Release Stream set, using version  '%s'", cfg.ClusterVersion)
		}
	} else {
		// get earlier available version from OSD
		if cfg.ClusterVersion, err = osd.PreviousVersion(cfg.UpgradeReleaseName); err != nil {
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
