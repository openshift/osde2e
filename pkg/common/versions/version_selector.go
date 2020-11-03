package versions

import (
	"fmt"
	"log"
	"math"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/common/versions/installselectors"
	"github.com/openshift/osde2e/pkg/common/versions/upgradeselectors"
)

// GetVersionForInstall will get a version based upon available configuration options.
func GetVersionForInstall(versionList *spi.VersionList) (*semver.Version, string, error) {
	var selectedVersionSelector installselectors.Interface = nil

	curPriority := math.MinInt32

	versionSelectors := installselectors.GetVersionSelectors()

	for _, versionSelector := range versionSelectors {
		if versionSelector.ShouldUse() && versionSelector.Priority() > curPriority {
			selectedVersionSelector = versionSelector
			curPriority = versionSelector.Priority()
		}
	}

	if selectedVersionSelector == nil {
		return nil, "", fmt.Errorf("unable to find an install version selector")
	}

	return selectedVersionSelector.SelectVersion(versionList)
}

// GetVersionForUpgrade will get a version based upon available configuration options.
func GetVersionForUpgrade(installVersion *semver.Version, versionList *spi.VersionList, upgradeSource spi.UpgradeSource) (string, string, error) {
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

	release, image, err := selectedVersionSelector.SelectVersion(spi.NewVersionBuilder().Version(installVersion).Build(), versionList)

	if release == nil || release.Version().Original() == "" {
		if err != nil {
			log.Printf("Error selecting version: %s", err.Error())
		}
		return util.NoVersionFound, "", err
	}

	return release.Version().Original(), image, err
}
