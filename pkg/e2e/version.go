package e2e

import (
	"errors"
	"fmt"
	"log"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/state"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/common/versions"
)

// ChooseVersions sets versions in cfg if not set based on defaults and upgrade options.
// If a release stream is set for an upgrade the previous available version is used and it's image is used for upgrade.
func ChooseVersions() (err error) {
	state := state.Instance

	// when defined, use set version
	if provider == nil {
		err = errors.New("osd must be setup when upgrading with release stream")
	} else {
		versionList, err := provider.Versions()

		if err != nil {
			return fmt.Errorf("error getting versions: %v", err)
		}

		clusterVersion, err := setupVersion(versionList)

		if err != nil {
			return fmt.Errorf("error while selecting install version: %v", err)
		}

		err = setupUpgradeVersion(clusterVersion, versionList)

		if err != nil {
			return fmt.Errorf("error while selecting upgrade version: %v", err)
		}
	}

	// Set the versions in metadata. If upgrade hasn't been chosen, it should still be omitted from the end result.
	metadata.Instance.SetClusterVersion(state.Cluster.Version)
	metadata.Instance.SetUpgradeVersion(state.Upgrade.ReleaseName)

	return err
}

// chooses between default version and nightly based on target versions.
func setupVersion(versionList *spi.VersionList) (*semver.Version, error) {
	var selectedVersion *semver.Version

	state := state.Instance

	versionType := "user supplied version"

	if len(state.Cluster.Version) == 0 {
		var err error

		selectedVersion, versionType, err = versions.GetVersionForInstall(versionList)
		if err == nil {
			if state.Cluster.EnoughVersionsForOldestOrMiddleTest && state.Cluster.PreviousVersionFromDefaultFound {
				state.Cluster.Version = util.SemverToOpenshiftVersion(selectedVersion)
			} else {
				log.Printf("Unable to get the %s.", versionType)
			}
		} else {
			return nil, fmt.Errorf("error finding default cluster version: %v", err)
		}
	} else {
		var err error
		// Make sure the cluster version is valid
		selectedVersion, err = util.OpenshiftVersionToSemver(state.Cluster.Version)

		if err != nil {
			return nil, fmt.Errorf("supplied version %s is invalid: %v", state.Cluster.Version, err)
		}
	}

	if selectedVersion == nil {
		log.Printf("Unable to select a cluster version.")
	} else {
		log.Printf("Using the %s '%s'", versionType, state.Cluster.Version)
	}

	return selectedVersion, nil
}

// chooses version based on optimal upgrade path
func setupUpgradeVersion(clusterVersion *semver.Version, versionList *spi.VersionList) error {
	var err error
	state := state.Instance

	if state.Upgrade.ReleaseName != "" || state.Upgrade.Image != "" {
		log.Printf("Using user supplied upgrade state.")
		return nil
	}

	if clusterVersion == nil {
		log.Printf("No install version found, skipping upgrade.")
		return nil
	}

	upgradeSource := provider.UpgradeSource()
	state.Upgrade.ReleaseName, state.Upgrade.Image, err = versions.GetVersionForUpgrade(clusterVersion, versionList, upgradeSource)

	if err != nil {
		return fmt.Errorf("error selecting an upgrade version: %v", err)
	}

	if state.Upgrade.ReleaseName == "" && state.Upgrade.Image == "" && err == nil {
		log.Printf("No upgrade selector found. Not selecting an upgrade version.")
		return nil
	}

	// set upgrade image
	log.Printf("Selecting version '%s' to be able to upgrade to '%s' using upgrade source '%s'",
		state.Cluster.Version, state.Upgrade.ReleaseName, upgradeSource)
	return nil
}
