package upgradeselectors

import (
	"fmt"
	"strings"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	registerSelector(latestVersion{})
}

// LatestVersion will always select the latest version of openshift.
type latestVersion struct{}

func (l latestVersion) ShouldUse() bool {
	return viper.GetBool(config.Upgrade.UpgradeToLatest)
}

func (l latestVersion) Priority() int {
	return 70
}

func (l latestVersion) SelectVersion(installVersion *spi.Version, versionList *spi.VersionList) (*spi.Version, string, error) {
	var newestVersion *spi.Version
	newestVersion = installVersion

	for _, v := range versionList.FindVersion(installVersion.Version().Original()) {
		for upgradeVersion := range v.AvailableUpgrades() {
			upgradeIsNightly := strings.Contains(upgradeVersion.Original(), "nightly")
			newestIsNightly := strings.Contains(newestVersion.Version().Original(), "nightly")
			if (upgradeIsNightly && !newestIsNightly) || upgradeVersion.GreaterThan(newestVersion.Version()) {
				newestVersion = spi.NewVersionBuilder().Version(upgradeVersion).Build()
			}
		}
	}

	if !newestVersion.Version().GreaterThan(installVersion.Version()) {
		return nil, "latest version", fmt.Errorf("No available upgrade path for version %s", installVersion.Version().Original())
	}
	return newestVersion, "latest version", nil
}
