package versions

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
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

	v, selector, err := selectedVersionSelector.SelectVersion(versionList)
	if err != nil {
		log.Printf("No valid install version found for selector `%s`", selector)
		for _, vers := range versionList.AvailableVersions() {
			log.Printf("%s - Default? %v", vers.Version().Original(), vers.Default())
		}
		return nil, selector, nil

	}

	// Refactor: Second time I see channel being set.
	channel := viper.GetString(config.Cluster.Channel)
	if channel != "stable" && !strings.Contains(v.Original(), channel) {
		v = semver.MustParse(fmt.Sprintf("%s-%s", v.Original(), channel))
	}

	return v, selector, err
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

	release, selector, err := selectedVersionSelector.SelectVersion(spi.NewVersionBuilder().Version(installVersion).Build(), versionList)

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
