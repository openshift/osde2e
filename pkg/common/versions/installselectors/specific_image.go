package installselectors

import (
	"fmt"
	"log"
	"strings"

	"github.com/Masterminds/semver/v3"
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
	// When running a /pj-rehearse, the `RELEASE_IMAGE_LATEST` provided is not
	// in a format that can be used by rosa so don't use this selector unless
	// the image contains nightly in the name:
	// `4.13.0-0.nightly-2023-08-24-052924`
	releaseImageLatest := viper.GetString(config.Cluster.ReleaseImageLatest)
	log.Printf("specific image value: %q", releaseImageLatest)
	return releaseImageLatest != "" && strings.Contains(releaseImageLatest, "nightly")
}

func (m specificImage) Priority() int {
	return 90
}

func (m specificImage) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	specificImage := viper.GetString(config.Cluster.ReleaseImageLatest)
	versionsWithoutDefault := removeDefaultVersion(versionList.AvailableVersions())
	versionType := "specific image"

	if specificImage == "" {
		return nil, versionType, fmt.Errorf("no image provided")
	}

	common.SortVersions(versionsWithoutDefault)

	versionFromImage := strings.ReplaceAll(specificImage, "registry.ci.openshift.org/ocp/release:", "")

	if strings.Contains(versionFromImage, "nightly") {
		versionFromImage += "-nightly"
	}

	versionToMatch, err := semver.NewVersion(versionFromImage)
	if err != nil {
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

func (m specificImage) String() string {
	return "specific image"
}
