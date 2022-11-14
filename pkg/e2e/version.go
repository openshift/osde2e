package e2e

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/common/versions"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ChooseVersions sets versions in cfg if not set based on defaults and upgrade options.
// If a release stream is set for an upgrade the previous available version is used and it's image is used for upgrade.
func ChooseVersions() (err error) {
	var clusterVersion *semver.Version
	var versionSelector string
	var versionList *spi.VersionList
	if provider == nil {
		err = errors.New("osd must be setup when upgrading with release stream")
	} else {

		if viper.GetString(config.Cluster.ReleaseImageLatest) != "" || viper.GetString(config.Cluster.InstallSpecificNightly) != "" {
			viper.Set(config.Cluster.Channel, "nightly")
		}

		err = wait.PollImmediate(1*time.Minute, 30*time.Minute, func() (bool, error) {
			versionList, err = provider.Versions()
			if err != nil {
				return false, fmt.Errorf("error getting versions: %v", err)
			}

			clusterVersion, versionSelector, err = setupVersion(versionList)
			if err != nil {
				return false, err
			}
			if clusterVersion == nil && versionSelector == "specific image" {
				log.Printf("Waiting for %s CIS to sync with the Release Controller", viper.GetString(config.Cluster.ReleaseImageLatest))
				return false, nil
			}

			return true, nil
		})

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
func setupVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	var selectedVersion *semver.Version
	var clusterVersion string

	versionType := "user supplied version"

	clusterVersion = viper.GetString(config.Cluster.Version)
	if len(clusterVersion) == 0 {
		var err error

		selectedVersion, versionType, err = versions.GetVersionForInstall(versionList)
		if err == nil && selectedVersion != nil {
			if viper.GetBool(config.Cluster.EnoughVersionsForOldestOrMiddleTest) && viper.GetBool(config.Cluster.PreviousVersionFromDefaultFound) {
				viper.Set(config.Cluster.Version, util.SemverToOpenshiftVersion(selectedVersion))
			} else {
				log.Printf("Unable to get the %s.", versionType)
			}
		} else {
			return nil, versionType, nil
		}
	} else {
		var err error
		// Make sure the cluster version is valid
		selectedVersion, err = util.OpenshiftVersionToSemver(clusterVersion)

		if err != nil {
			return nil, versionType, fmt.Errorf("supplied version %s is invalid: %v", clusterVersion, err)
		}
	}

	if selectedVersion == nil {
		log.Printf("Unable to select a cluster version.")
	} else {
		log.Printf("Using the %s '%s'", versionType, selectedVersion.Original())
	}

	return selectedVersion, versionType, nil
}

// chooses version based on optimal upgrade path
func setupUpgradeVersion(clusterVersion *semver.Version, versionList *spi.VersionList) error {
	var err error

	if viper.GetString(config.Upgrade.ReleaseName) != "" || viper.GetString(config.Upgrade.Image) != "" {
		log.Printf("Using user supplied upgrade state.")
		return nil
	}

	if clusterVersion == nil {
		log.Printf("No install version found, skipping upgrade.")
		return nil
	}

	upgradeSource := provider.UpgradeSource()
	releaseName, image, err := versions.GetVersionForUpgrade(clusterVersion, versionList, upgradeSource)
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
