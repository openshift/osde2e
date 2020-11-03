package upgradeselectors

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
)

func init() {
	registerSelector(nextYVersion{})
}

// nextYVersion returns the next minor version upgrade available
type nextYVersion struct{}

func (l nextYVersion) ShouldUse() bool {
	return viper.GetBool(config.Upgrade.UpgradeToNextY)
}

func (l nextYVersion) Priority() int {
	return 70
}

func (l nextYVersion) SelectVersion(installVersion *spi.Version, versionList *spi.VersionList) (*spi.Version, string, error) {
	var newestVersion *spi.Version
	newestVersion = installVersion

	for _, v := range versionList.FindVersion(installVersion.Version().Original()) {
		for upgradeVersion := range v.AvailableUpgrades() {
			if upgradeVersion.Minor() == installVersion.Version().Minor()+1 && upgradeVersion.GreaterThan(newestVersion.Version()) {
				newestVersion = spi.NewVersionBuilder().Version(upgradeVersion).Build()
			}
		}
	}

	if !newestVersion.Version().GreaterThan(installVersion.Version()) {
		return nil, "next y version", fmt.Errorf("No available upgrade path for version %s", installVersion.Version().Original())
	}
	return newestVersion, "next y version", nil
}
