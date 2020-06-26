package versions

import (
	"errors"
	"fmt"
	"log"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/spf13/viper"
)

// ChooseVersions sets versions in cfg if not set based on defaults and upgrade options.
// If a release stream is set for an upgrade the previous available version is used and it's image is used for upgrade.
func ChooseVersions() (err error) {
	log.Println("Entered choose version function")
	provider, err := providers.ClusterProvider()
	if err != nil {
		return fmt.Errorf("error getting cluster provider: %v", err)
	}
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
	metadata.Instance.SetClusterVersion(viper.GetString(config.Cluster.Version))
	metadata.Instance.SetUpgradeVersion(viper.GetString(config.Upgrade.ReleaseName))

	return err
}

// chooses between default version and nightly based on target versions.
func setupVersion(versionList *spi.VersionList) (*semver.Version, error) {
	var selectedVersion *semver.Version

	versionType := "user supplied version"

	clusterVersion := viper.GetString(config.Cluster.Version)
	if len(clusterVersion) == 0 {
		var err error

		selectedVersion, versionType, err = GetVersionForInstall(versionList)
		if err == nil {
			if viper.GetBool(config.Cluster.EnoughVersionsForOldestOrMiddleTest) && viper.GetBool(config.Cluster.PreviousVersionFromDefaultFound) {
				viper.Set(config.Cluster.Version, util.SemverToOpenshiftVersion(selectedVersion))
				clusterVersion = util.SemverToOpenshiftVersion(selectedVersion)
			} else {
				log.Printf("Unable to get the %s.", versionType)
			}
		} else {
			return nil, fmt.Errorf("error finding default cluster version: %v", err)
		}
	} else {
		var err error
		// Make sure the cluster version is valid
		selectedVersion, err = util.OpenshiftVersionToSemver(clusterVersion)

		if err != nil {
			return nil, fmt.Errorf("supplied version %s is invalid: %v", clusterVersion, err)
		}
	}

	if selectedVersion == nil {
		log.Printf("Unable to select a cluster version.")
	} else {
		log.Printf("Using the %s '%s'", versionType, clusterVersion)
	}

	return selectedVersion, nil
}

// chooses version based on optimal upgrade path
func setupUpgradeVersion(clusterVersion *semver.Version, versionList *spi.VersionList) error {
	var err error
	provider, err := providers.ClusterProvider()
	if viper.GetString(config.Upgrade.ReleaseName) != "" || viper.GetString(config.Upgrade.Image) != "" {
		log.Printf("Using user supplied upgrade state.")
		return nil
	}

	if clusterVersion == nil {
		log.Printf("No install version found, skipping upgrade.")
		return nil
	}

	upgradeSource := provider.UpgradeSource()
	releaseName, image, err := GetVersionForUpgrade(clusterVersion, versionList, upgradeSource)

	if err != nil {
		return fmt.Errorf("error selecting an upgrade version: %v", err)
	}

	if releaseName == "" && image == "" && err == nil {
		log.Printf("No upgrade selector found. Not selecting an upgrade version.")
		return nil
	}

	viper.Set(config.Upgrade.ReleaseName, releaseName)
	viper.Set(config.Upgrade.Image, image)

	// set upgrade image
	log.Printf("Selecting version '%s' to be able to upgrade to '%s' using upgrade source '%s'",
		viper.GetString(config.Cluster.Version), releaseName, upgradeSource)
	return nil
}
