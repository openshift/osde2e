package installselectors

import (
	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/versions/common"
)

func init() {
	registerSelector(oldestClusterImageSet{})
}

// oldestClusterImageSet will use the oldest image from the list of available versions.
type oldestClusterImageSet struct{}

func (o oldestClusterImageSet) ShouldUse() bool {
	return viper.GetBool(config.Cluster.UseOldestClusterImageSetForInstall)
}

func (o oldestClusterImageSet) Priority() int {
	return 60
}

func (o oldestClusterImageSet) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	versionsWithoutDefault := removeDefaultVersion(versionList.AvailableVersions())
	numVersions := len(versionsWithoutDefault)
	versionType := "oldest version"

	// We don't want to fail entirely if there aren't enough versions. It's valid and perhaps even expected
	// that we d on't have enough versions for a middle cluster image set.
	if numVersions < 2 {
		viper.Set(config.Cluster.EnoughVersionsForOldestOrMiddleTest, false)
		return nil, versionType, nil
	}

	common.SortVersions(versionsWithoutDefault)

	return versionsWithoutDefault[0].Version(), versionType, nil
}
