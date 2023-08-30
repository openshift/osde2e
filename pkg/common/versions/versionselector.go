package versions

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/common/versions/installselectors"
	"github.com/openshift/osde2e/pkg/common/versions/upgradeselectors"
	"k8s.io/apimachinery/pkg/util/wait"
)

type VersionSelector struct {
	Provider       spi.Provider
	versionList    *spi.VersionList
	clusterVersion *semver.Version
}

// SelectClusterVersions sets versions in cfg if not set based on defaults and upgrade options.
// If a release stream is set for an upgrade the previous available version is used and it's image is used for upgrade.
func (v *VersionSelector) SelectClusterVersions() error {
	var err error
	var versionSelector string

	if v.Provider == nil {
		err = errors.New("no cluster provider was setup")
	} else {
		if viper.GetString(config.Cluster.ReleaseImageLatest) != "" || viper.GetString(config.Cluster.InstallSpecificNightly) != "" {
			viper.Set(config.Cluster.Channel, "nightly")
		}

		err = wait.PollImmediate(1*time.Minute, 30*time.Minute, func() (bool, error) {
			v.versionList, err = v.Provider.Versions()
			if err != nil {
				return false, fmt.Errorf("error getting versions: %v", err)
			}

			v.clusterVersion, versionSelector, err = v.setInstallVersion()
			if err != nil {
				return false, err
			}
			if v.clusterVersion == nil && (versionSelector == "specific image" || versionSelector == "specific nightly") {
				log.Println("Waiting for CIS to sync with the Release Controller")
				return false, nil
			}

			return true, nil
		})

		if err != nil {
			return fmt.Errorf("error while selecting install version: %v", err)
		}

		if err = v.setUpgradeVersion(); err != nil {
			return fmt.Errorf("error while selecting upgrade version: %v", err)
		}
	}

	// Set the versions in metadata. If upgrade hasn't been chosen, it should still be omitted from the end result.
	metadata.Instance.SetClusterVersion(viper.GetString(config.Cluster.Version))
	metadata.Instance.SetUpgradeVersion(viper.GetString(config.Upgrade.ReleaseName))

	return nil
}

// setInstallVersion chooses the cluster install version to use
func (v *VersionSelector) setInstallVersion() (*semver.Version, string, error) {
	var err error
	var selectedVersion *semver.Version
	versionType := "user supplied version"

	clusterVersion := viper.GetString(config.Cluster.Version)
	if len(clusterVersion) == 0 {
		selectedVersion, versionType, err = v.getInstallVersion()
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

// setUpgradeVersion chooses the cluster upgrade version
func (v *VersionSelector) setUpgradeVersion() error {
	if viper.GetString(config.Upgrade.ReleaseName) != "" || viper.GetString(config.Upgrade.Image) != "" {
		log.Printf("Using user supplied upgrade state.")
		return nil
	}

	if v.clusterVersion == nil {
		log.Printf("No install version found, skipping upgrade.")
		return nil
	}

	upgradeSource := v.Provider.UpgradeSource()
	releaseName, image, err := v.getUpgradeVersion()
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

// getInstallVersion will get a version based upon available configuration options.
func (v *VersionSelector) getInstallVersion() (*semver.Version, string, error) {
	var selectedVersionSelector installselectors.Interface = nil

	curPriority := math.MinInt32

	versionSelectors := installselectors.GetVersionSelectors()

	// Review: This is a hack to get around the fact that the version selector for the latest version
	for _, versionSelector := range versionSelectors {
		if versionSelector.ShouldUse() && versionSelector.Priority() > curPriority {
			selectedVersionSelector = versionSelector
			curPriority = versionSelector.Priority()
		}
	}

	if selectedVersionSelector == nil {
		return nil, "", fmt.Errorf("unable to find an install version selector")
	}

	log.Printf("Using version selector %q", selectedVersionSelector)

	version, selector, err := selectedVersionSelector.SelectVersion(v.versionList)
	if err != nil {
		log.Printf("Unable to find image using selector `%s`. Error: %v", selector, err)
		return nil, selector, nil

	}

	// Refactor: Second time I see channel being set.
	channel := viper.GetString(config.Cluster.Channel)
	if channel != "stable" && !strings.Contains(version.Original(), channel) {
		version = semver.MustParse(fmt.Sprintf("%s-%s", version.Original(), channel))
	}

	return version, selector, err
}

// getUpgradeVersion will get a version based upon available configuration options.
func (v *VersionSelector) getUpgradeVersion() (string, string, error) {
	var selectedVersionSelector upgradeselectors.Interface = nil

	curPriority := math.MinInt32

	versionSelectors := upgradeselectors.GetVersionSelectors()

	for _, versionSelector := range versionSelectors {
		if versionSelector.ShouldUse() && versionSelector.Priority() > curPriority {
			selectedVersionSelector = versionSelector
			curPriority = versionSelector.Priority()
		}
	}

	// If no version selector has been found for an upgrade, assume that an upgrade is not being asked for.
	if selectedVersionSelector == nil {
		return "", "", nil
	}

	release, selector, err := selectedVersionSelector.SelectVersion(spi.NewVersionBuilder().Version(v.clusterVersion).Build(), v.versionList)

	if release == nil || release.Version().Original() == "" {
		if err != nil {
			log.Printf("Error selecting version: %s", err.Error())
		}
		return util.NoVersionFound, "", err
	}

	openshiftRelease := fmt.Sprintf("openshift-v%s", release.Version().Original())

	log.Printf("Selected %s using selector `%s`", openshiftRelease, selector)

	return openshiftRelease, "", err
}
