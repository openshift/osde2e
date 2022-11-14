package installselectors

import (
	"fmt"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/versions/common"
)

func init() {
	registerSelector(latestYVersion{})
}

// LatestVersion will always select the latest version of openshift.
type latestYVersion struct{}

func (l latestYVersion) ShouldUse() bool {
	return viper.GetBool(config.Cluster.LatestYReleaseAfterProdDefault)
}

func (l latestYVersion) Priority() int {
	return 70
}

func (l latestYVersion) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	availableVersions := versionList.AvailableVersions()
	numVersions := len(availableVersions)
	versionType := "latest y version from default"

	if numVersions == 0 {
		return nil, versionType, fmt.Errorf("not enough versions to select the latest version")
	}

	common.SortVersions(availableVersions)

	var defaultVersion *spi.Version
	var latestYVersion *spi.Version
	var nextMinor int64

	for _, version := range availableVersions {
		if version.Default() {
			defaultVersion = version
			nextMinor = defaultVersion.Version().Minor() + 1
			continue
		}

		if defaultVersion == nil {
			continue
		}

		if version.Version().Minor() == nextMinor {
			if latestYVersion == nil {
				latestYVersion = version
			}
			if version.Version().GreaterThan(latestYVersion.Version()) {
				latestYVersion = version
			}
		}
	}

	if latestYVersion == nil {
		return nil, versionType, fmt.Errorf("no version matching selector found")
	}

	return latestYVersion.Version(), versionType, nil
}
