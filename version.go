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
	}

	// if release stream target is set, try to use previous version
	if cfg.UpgradeImage == "" && cfg.UpgradeReleaseStream != "" {
		if osd == nil {
			return errors.New("osd must be setup when upgrading with release stream")
		}

		name, pullSpec, err := upgrade.LatestRelease(cfg.UpgradeReleaseStream)
		if err != nil {
			return fmt.Errorf("couldn't get latest release from release-controller: %v", err)
		}

		// get earlier available version from OSD
		if cfg.ClusterVersion, err = osd.PreviousVersion(name); err != nil {
			return fmt.Errorf("failed retrieving previous version to '%s': %v", name, err)
		}

		// set upgrade image
		log.Printf("Selecting version '%s' to be able to upgrade to '%s' on release stream '%s'",
			cfg.ClusterVersion, name, cfg.UpgradeReleaseStream)
		cfg.UpgradeImage = pullSpec
	} else if cfg.ClusterVersion, err = OSD.DefaultVersion(); err != nil {
		return fmt.Errorf("failed to get default version: %v", err)
	} else {
		log.Printf("CLUSTER_VERSION not set, using the current default '%s'", cfg.ClusterVersion)
	}
	return nil
}
