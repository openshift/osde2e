package upgradeselectors

import (
	"fmt"
	"strings"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
)

func init() {
	registerSelector(latestZVersion{})
}

// latestZVersion returns the latest patch version upgrade available
type latestZVersion struct{}

func (l latestZVersion) ShouldUse() bool {
	return viper.GetBool(config.Upgrade.UpgradeToLatestZ)
}

func (l latestZVersion) Priority() int {
	return 70
}

func (l latestZVersion) SelectVersion(installVersion *spi.Version, versionList *spi.VersionList) (*spi.Version, string, error) {
	var newestVersion *spi.Version

	newestVersion = installVersion

	for _, v := range versionList.FindVersion(installVersion.Version().Original()) {
		for upgradeVersion := range v.AvailableUpgrades() {
			if upgradeVersion.Minor() != installVersion.Version().Minor() {
				continue
			}
			if upgradeVersion.Prerelease() != "" {
				if strings.Contains(upgradeVersion.Prerelease(), "nightly") && upgradeVersion.GreaterThan(newestVersion.Version()) {
					newestVersion = spi.NewVersionBuilder().Version(upgradeVersion).Build()
				}
			} else {
				if newestVersion.Version().Prerelease() == "" {
					if upgradeVersion.GreaterThan(newestVersion.Version()) {
						newestVersion = spi.NewVersionBuilder().Version(upgradeVersion).Build()
					}
				}
			}
		}
	}
	if !newestVersion.Version().GreaterThan(installVersion.Version()) {
		return nil, "latest z version", fmt.Errorf("No available upgrade path for version %s", installVersion.Version().Original())
	}
	return newestVersion, "latest z version", nil
}
