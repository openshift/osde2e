package installselectors

import (
	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/versions/common"
)

func init() {
	registerSelector(middleClusterImageSet{})
}

// MiddleClusterImageSet will use the image in the middle of the available versions.
type middleClusterImageSet struct{}

func (m middleClusterImageSet) ShouldUse() bool {
	return viper.GetBool(config.Cluster.UseMiddleClusterImageSetForInstall)
}

func (m middleClusterImageSet) Priority() int {
	return 60
}

func (m middleClusterImageSet) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	versionsWithoutDefault := removeDefaultVersion(versionList.AvailableVersions())
	numVersions := len(versionsWithoutDefault)
	versionType := "middle version"

	// We don't want to fail entirely if there aren't enough versions. It's valid and perhaps even expected
	// that we d on't have enough versions for a middle cluster image set.
	if numVersions < 2 {
		viper.Set(config.Cluster.EnoughVersionsForOldestOrMiddleTest, false)
		return nil, versionType, nil
	}

	common.SortVersions(versionsWithoutDefault)

	return versionsWithoutDefault[numVersions/2].Version(), versionType, nil
}
