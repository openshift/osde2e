package installselectors

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/versions/common"
)

func init() {
	registerSelector(nextVersionAfterProdDefault{})
}

// nextVersionAfterProdDefault will select a version that is N releases from the current production default.
type nextVersionAfterProdDefault struct{}

func (n nextVersionAfterProdDefault) ShouldUse() bool {
	return config.Instance.Cluster.NextReleaseAfterProdDefault > -1
}

func (n nextVersionAfterProdDefault) Priority() int {
	return 40
}

func (n nextVersionAfterProdDefault) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	numReleasesAfterProdDefault := config.Instance.Cluster.NextReleaseAfterProdDefault
	defaultVersion := versionList.Default()
	selectedVersion, err := common.NextReleaseAfterGivenVersionFromVersionList(defaultVersion, versionList.AvailableVersions(), numReleasesAfterProdDefault)
	return selectedVersion, fmt.Sprintf("%d release(s) from the default version in prod", numReleasesAfterProdDefault), err
}
