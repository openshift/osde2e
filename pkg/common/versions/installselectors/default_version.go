package installselectors

import (
	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	registerSelector(defaultVersion{})
}

// DefaultVersion is the fallback selector.
type defaultVersion struct{}

func (d defaultVersion) ShouldUse() bool {
	return true
}

func (d defaultVersion) Priority() int {
	return 0
}

func (d defaultVersion) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	return versionList.Default(), "current default", nil
}
