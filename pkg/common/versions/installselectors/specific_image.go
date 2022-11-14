package installselectors

import (
	"fmt"
	"log"
	"strings"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/versions/common"
)

func init() {
	registerSelector(specificImage{})
}

// specificImage will grep out the version in a supplied image then attempt to install it
type specificImage struct{}

func (m specificImage) ShouldUse() bool {
	return viper.GetString(config.Cluster.ReleaseImageLatest) != ""
}

func (m specificImage) Priority() int {
	return 100
}

func (m specificImage) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	specificImage := viper.GetString(config.Cluster.ReleaseImageLatest)
	versionsWithoutDefault := removeDefaultVersion(versionList.AvailableVersions())
	versionType := "specific image"

	if specificImage == "" {
		return nil, versionType, fmt.Errorf("no image provided")
	}

	common.SortVersions(versionsWithoutDefault)

	versionFromImage := strings.Replace(specificImage, "registry.ci.openshift.org/ocp/release:", "", -1)

	if strings.Contains(versionFromImage, "nightly") {
		versionFromImage += "-nightly"
	}

	versionToMatch := semver.MustParse(versionFromImage)

	if versionToMatch == nil {
		return nil, versionType, fmt.Errorf("error parsing semver version for %s", specificImage)
	}

	log.Printf("Looking to match %s", versionToMatch)

	for _, version := range versionsWithoutDefault {
		if version.Version().Original() == versionToMatch.Original() {
			return version.Version(), versionType, nil
		}
	}

	return nil, versionType, fmt.Errorf("no valid nightly found for version %s", specificImage)
}
