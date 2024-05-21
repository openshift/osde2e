package installselectors

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	registerSelector(triggeredNightlies{})
}

type triggeredNightlies struct{}

func (t triggeredNightlies) ShouldUse() bool {
	return strings.Contains(os.Getenv("RELEASE_IMAGE_LATEST"), "nightly")
}

func (t triggeredNightlies) Priority() int {
	return 100
}

func (t triggeredNightlies) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	// RELEASE_IMAGE_LATEST is a tag populated in release controller jobs.
	// It has the following form.
	// registry.ci.openshift.org/ocp/release:4.15.0-0.nightly-2024-05-15-103159
	// Extract the version tag from it.
	releaseImageLatestRegex := regexp.MustCompile(`\d.\d+.\d-\d.nightly-\d{4}-\d{2}-\d{2}-\d+`)

	releaseImageLatest := os.Getenv("RELEASE_IMAGE_LATEST")
	matches := releaseImageLatestRegex.FindStringSubmatch(releaseImageLatest)
	if len(matches) == 0 {
		return nil, t.String(), fmt.Errorf("failed to match regular expression with RELEASE_IMAGE_LATEST: %q", releaseImageLatest)
	}
	payloadName := matches[0] + "-nightly"

	versionsWithoutDefault := removeDefaultVersion(versionList.AvailableVersions())

	versionToMatch, err := semver.NewVersion(payloadName)
	if err != nil {
		return nil, t.String(), fmt.Errorf("error parsing semver version for %s", payloadName)
	}

	log.Printf("Looking to match %s", versionToMatch)

	for _, version := range versionsWithoutDefault {
		if version.Version().Original() == versionToMatch.Original() {
			return version.Version(), t.String(), nil
		}
	}

	return nil, t.String(), fmt.Errorf("failed to find version %q", payloadName)
}

func (t triggeredNightlies) String() string {
	return "triggered nightly"
}
