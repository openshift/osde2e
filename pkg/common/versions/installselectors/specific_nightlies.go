package installselectors

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/versions/common"
)

func init() {
	registerSelector(specificNightlies{})
}

// SpecificNightlies attempts to parse a config option as semver and use the major.minor to look for nightlies
type specificNightlies struct{}

func (m specificNightlies) ShouldUse() bool {
	return viper.GetString(config.Cluster.InstallSpecificNightly) != ""
}

func (m specificNightlies) Priority() int {
	return 60
}

func (m specificNightlies) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	specificNightly := viper.GetString(config.Cluster.InstallSpecificNightly)
	versionsWithoutDefault := removeDefaultVersion(versionList.AvailableVersions())
	versionType := "specific nightly"

	if specificNightly == "" {
		return nil, versionType, fmt.Errorf("no version to match nightly found")
	}

	common.SortVersions(versionsWithoutDefault)

	versionToMatch := semver.MustParse(specificNightly)

	if versionToMatch == nil {
		return nil, versionType, fmt.Errorf("error parsing semver version for %s", specificNightly)
	}

	for i := len(versionsWithoutDefault) - 1; i > -1; i-- {
		if strings.Contains(versionsWithoutDefault[i].Version().Original(), "nightly") && versionsWithoutDefault[i].Version().Major() == versionToMatch.Major() && versionsWithoutDefault[i].Version().Minor() == versionToMatch.Minor() {
			// Since we're going through a list in reverse-order, the first X.Y that matches should be the latest!
			return versionsWithoutDefault[i].Version(), versionType, nil
		}
	}

	return nil, versionType, fmt.Errorf("no valid nightly found for version %s", specificNightly)
}
