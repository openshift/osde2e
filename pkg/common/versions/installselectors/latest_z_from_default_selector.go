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
	registerSelector(latestZVersion{})
}

// LatestVersion will always select the latest version of openshift.
type latestZVersion struct{}

func (l latestZVersion) ShouldUse() bool {
	return viper.GetBool(config.Cluster.LatestZReleaseAfterProdDefault)
}

func (l latestZVersion) Priority() int {
	return 70
}

func (l latestZVersion) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	availableVersions := versionList.AvailableVersions()
	numVersions := len(availableVersions)
	versionType := "latest z version from default"

	if numVersions == 0 {
		return nil, versionType, fmt.Errorf("not enough versions to select the latest version")
	}

	common.SortVersions(availableVersions)

	var defaultVersion *spi.Version
	var latestZVersion *spi.Version

	for _, version := range availableVersions {
		if version.Default() {
			defaultVersion = version
			latestZVersion = version
		}

		if defaultVersion == nil {
			continue
		}

		if version.Version().Minor() == defaultVersion.Version().Minor() {
			if version.Version().GreaterThan(latestZVersion.Version()) {
				latestZVersion = version
			}
		}
	}

	if latestZVersion == nil || latestZVersion.Version().Equal(defaultVersion.Version()) {
		return nil, versionType, fmt.Errorf("no version matching selector found")
	}

	return latestZVersion.Version(), versionType, nil
}
