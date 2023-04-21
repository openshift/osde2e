package installselectors

import (
	"fmt"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	registerSelector(latestYFromDelta{})
}

type latestYFromDelta struct{}

func (m latestYFromDelta) ShouldUse() bool {
	return viper.GetInt(config.Cluster.InstallLatestYFromDelta) != 0
}

func (m latestYFromDelta) Priority() int {
	return 70
}

func (m latestYFromDelta) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	latestYFromDelta := viper.GetInt64(config.Cluster.InstallLatestYFromDelta)
	versionType := fmt.Sprintf("latest Y from delta %d", latestYFromDelta)
	versions := versionList.AvailableVersions()

	if len(versions) == 0 {
		return nil, versionType, fmt.Errorf("no versions supplied, unable to select version")
	}

	defaultVersion, err := findDefaultVersion(versions)
	if err != nil {
		return nil, versionType, err
	}

	versionType = fmt.Sprintf("latest Y '%s' from delta %d", defaultVersion.Version().String(), latestYFromDelta)

	for _, version := range versions {
		if defaultVersion.Version().Minor()+latestYFromDelta == version.Version().Minor() {
			return version.Version(), versionType, nil
		}
	}

	return nil, versionType, fmt.Errorf("no version found matching the selector")
}
