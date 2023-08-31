package installselectors

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	registerSelector(triggeredNightlies{})
}

type triggeredNightlies struct{}

func (t triggeredNightlies) ShouldUse() bool {
	return strings.Contains(os.Getenv("PROW_JOB_ID"), "nightly")
}

func (t triggeredNightlies) Priority() int {
	return 100
}

func (t triggeredNightlies) SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error) {
	prowJobID := os.Getenv("PROW_JOB_ID")
	jobNameSafe := os.Getenv("JOB_NAME_SAFE")
	payloadName := strings.ReplaceAll(prowJobID, "-"+jobNameSafe, "")
	payloadName += "-nightly"

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
