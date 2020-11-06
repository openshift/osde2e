package installselectors

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/versions/common"
	"github.com/spf13/viper"
)

func init() {
	registerSelector(latestVersion{})
}

// LatestVersion will always select the latest version of openshift.
type latestVersion struct{}

func (l latestVersion) ShouldUse() bool {
	return viper.GetBool(config.Cluster.UseLatestVersionForInstall)
}

func (l latestVersion) Priority() int {
	return 70
}

func (l latestVersion) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	availableVersions := versionList.AvailableVersions()
	numVersions := len(availableVersions)
	versionType := "latest version"

	if numVersions == 0 {
		return nil, versionType, fmt.Errorf("not enough versions to select the latest version")
	}

	common.SortVersions(availableVersions)

	return availableVersions[len(availableVersions)-1].Version(), versionType, nil
}
