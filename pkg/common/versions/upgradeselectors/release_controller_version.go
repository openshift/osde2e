package upgradeselectors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/common/versions/common"
)

const (
	// format string for release stream latest from release controller
	latestReleaseControllerURLFmt = "https://openshift-release.svc.ci.openshift.org/api/v1/releasestream/%s/latest"
)

func init() {
	registerSelector(releaseControllerUpgrade{})
}

// releaseControllerUpgrade will select an upgrade target based on the ReleaseController.
type releaseControllerUpgrade struct{}

func (r releaseControllerUpgrade) ShouldUse(upgradeSource spi.UpgradeSource) bool {
	return upgradeSource == spi.ReleaseControllerSource && config.Instance.Upgrade.NextReleaseAfterProdDefaultForUpgrade > -1
}

func (r releaseControllerUpgrade) Priority() int {
	return 40
}

func (r releaseControllerUpgrade) SelectVersion(installVersion *semver.Version, versionList *spi.VersionList) (string, string, error) {
	cfg := config.Instance

	// If we're using the release controller, we're trying to do relative version selection.
	// We'll confirm this in case things change in the future and just proceed with that assumption.
	nextVersion, err := common.NextReleaseAfterGivenVersionFromVersionList(versionList.Default(), versionList.AvailableVersions(), cfg.Upgrade.NextReleaseAfterProdDefaultForUpgrade)

	if err != nil {
		return "", "", fmt.Errorf("error determining next version to upgrade to: %v", err)
	}

	releaseStream := fmt.Sprintf("%d.%d.0-0.nightly", nextVersion.Major(), nextVersion.Minor())

	return latestReleaseFromReleaseController(releaseStream)
}

// latestAccepted information from release controller.
type latestAccepted struct {
	Name        string `json:"name"`
	PullSpec    string `json:"pullSpec"`
	DownloadURL string `json:"downloadURL"`
}

// latestReleaseFromReleaseController retrieves latest release information for given releaseStream on the release controller.
func latestReleaseFromReleaseController(releaseStream string) (name, pullSpec string, err error) {
	var resp *http.Response
	var data []byte

	latestURL := fmt.Sprintf(latestReleaseControllerURLFmt, releaseStream)
	resp, err = http.Get(latestURL)
	if err != nil {
		err = fmt.Errorf("failed to get latest for stream '%s': %v", releaseStream, err)
		return
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed reading body: %v", err)
		return
	}

	latest := latestAccepted{}
	if err = json.Unmarshal(data, &latest); err != nil {
		return "", "", fmt.Errorf("error decoding body of '%s': %v", data, err)
	}

	metadata.Instance.SetUpgradeVersionSource("release controller")

	if latest.Name == "" {
		return util.NoVersionFound, "", nil
	}

	releaseName := ensureReleasePrefix(latest.Name)

	if releaseStream == "" {
		return util.NoVersionFound, "", nil
	}

	return releaseName, latest.PullSpec, err
}

func ensureReleasePrefix(release string) string {
	if len(release) > 0 && !strings.Contains(release, "openshift-v") {
		log.Printf("Version %s didn't have prefix. Adding....", release)
		release = "openshift-v" + release
	}
	return release
}
