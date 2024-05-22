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
	log.Printf("PROW_JOB_ID: %q", os.Getenv("PROW_JOB_ID"))
	log.Printf("RELEASE_IMAGE_LATEST: %q", os.Getenv("RELEASE_IMAGE_LATEST"))
	return strings.Contains(os.Getenv("PROW_JOB_ID"), "nightly") || strings.Contains(os.Getenv("RELEASE_IMAGE_LATEST"), "nightly")
}

func (t triggeredNightlies) Priority() int {
	return 100
}

func (t triggeredNightlies) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	// PROW_JOB_ID is an env var populated in release controller prow jobs in the following form.
	// 4.15.0-0.nightly-2024-05-22-165653-<jobname>
	// RELEASE_IMAGE_LATEST is an env var populated in release controller jobs in the following form.
	// registry.ci.openshift.org/ocp/release:4.15.0-0.nightly-2024-05-15-103159
	// we will use whichever of these two vars is available to get version tag
	matchTag := os.Getenv("RELEASE_IMAGE_LATEST")
	if matchTag == "" {
		matchTag = os.Getenv("PROW_JOB_ID")
	}

	versionRegex := regexp.MustCompile(`\d.\d+.\d-\d.nightly-\d{4}-\d{2}-\d{2}-\d+`)
	matches := versionRegex.FindStringSubmatch(matchTag)

	if len(matches) == 0 {
		return nil, t.String(), fmt.Errorf("failed to match nightly version tag: %q", matchTag)
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
