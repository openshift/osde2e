package installselectors

import (
	"fmt"
	"log"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	registerSelector(deltaReleaseFromDefault{})
}

// deltaReleaseFromDefault is the selector which will select a release that is a delta from the default. The
// delta can be negative or positive.
type deltaReleaseFromDefault struct{}

func (d deltaReleaseFromDefault) ShouldUse() bool {
	return viper.GetInt(config.Cluster.DeltaReleaseFromDefault) != 0
}

func (d deltaReleaseFromDefault) Priority() int {
	return 50
}

func (d deltaReleaseFromDefault) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	availableVersions := versionList.AvailableVersions()
	defaultIndex := findDefaultVersionIndex(availableVersions)
	deltaReleasesFromDefault := viper.GetInt(config.Cluster.DeltaReleaseFromDefault)
	versionType := fmt.Sprintf("version %d releases from the default", deltaReleasesFromDefault)

	if defaultIndex < 0 {
		log.Printf("unable to find default version in avaialable version list")
		viper.Set(config.Cluster.PreviousVersionFromDefaultFound, false)
	}

	targetIndex := defaultIndex + deltaReleasesFromDefault

	if targetIndex < 0 || targetIndex >= len(availableVersions) {
		log.Printf("not enough enabled versions to go back %d releases", deltaReleasesFromDefault)
		viper.Set(config.Cluster.PreviousVersionFromDefaultFound, false)
		return nil, versionType, nil
	}

	return availableVersions[targetIndex].Version(), versionType, nil
}

func findDefaultVersionIndex(versions []*spi.Version) int {
	for index, version := range versions {
		if version.Default() {
			return index
		}
	}

	return -1
}
