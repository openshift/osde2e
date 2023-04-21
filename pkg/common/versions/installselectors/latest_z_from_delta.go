package installselectors

import (
	"fmt"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	registerSelector(latestZFromDelta{})
}

type latestZFromDelta struct{}

func (m latestZFromDelta) ShouldUse() bool {
	return viper.GetInt(config.Cluster.InstallLatestZFromDelta) != 0
}

func (m latestZFromDelta) Priority() int {
	return 70
}

func (m latestZFromDelta) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	latestZFromDelta := viper.GetInt64(config.Cluster.InstallLatestZFromDelta)
	versionType := fmt.Sprintf("latest Z from delta %d", latestZFromDelta)
	versions := versionList.AvailableVersions()

	if len(versions) == 0 {
		return nil, versionType, fmt.Errorf("no versions supplied, unable to select version")
	}

	defaultVersion, err := findDefaultVersion(versions)
	if err != nil {
		return nil, versionType, err
	}

	versionType = fmt.Sprintf("latest Z '%s' from delta %d", defaultVersion.Version().String(), latestZFromDelta)

	for _, version := range versions {
		if defaultVersion.Version().Minor() != version.Version().Minor() {
			continue
		}

		if defaultVersion.Version().Patch()+latestZFromDelta == version.Version().Patch() {
			return version.Version(), versionType, nil
		}
	}

	return nil, versionType, fmt.Errorf("no version found matching the selector")
}
