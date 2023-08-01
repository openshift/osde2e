package installselectors

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
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
	versionType := "current default"
	versionDefault := versionList.Default()
	if versionDefault == nil {
		return nil, versionType, fmt.Errorf("no default version set for channel group: %s", viper.GetString(config.Cluster.Channel))
	}
	return versionDefault, versionType, nil
}
