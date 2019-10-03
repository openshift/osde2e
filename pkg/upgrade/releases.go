package upgrade

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/osd"
)

const (
	// format string for release stream latest from release controller
	latestReleaseControllerURLFmt = "https://openshift-release.svc.ci.openshift.org/api/v1/releasestream/%s/latest"
	// format string for Cincinnati releases
	cincinnatiURLFmt = "%s/api/upgrades_info/v1/graph?channel=%s"
)

// LatestRelease retrieves latest release information for given releaseStream. Will use Cincinnati for stage/prod.
func LatestRelease(cfg *config.Config, releaseStream string, use_release_controller_for_int bool) (name, pullSpec string, err error) {
	var resp *http.Response
	var data []byte
	if cfg.OSDEnv == "int" && use_release_controller_for_int {
		log.Printf("Using the release controller.")
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

		return ensureReleasePrefix(latest.Name), latest.PullSpec, nil
	}

	log.Printf("Using Cincinnati.")

	// If stage or prod, use Cincinnati instead of the release controller
	cincinnatiFormattedURL := fmt.Sprintf(cincinnatiURLFmt, osd.Environments.Choose(cfg.OSDEnv), releaseStream)

	var req *http.Request

	// Cincinnati requires an Accept header, so we add it in here
	req, err = http.NewRequest("GET", cincinnatiFormattedURL, nil)
	req.Header.Set("Accept", "application/json")

	if err != nil {
		return "", "", fmt.Errorf("failed to create Cincinnati request for URL '%s': %v", cincinnatiFormattedURL, err)
	}

	resp, err = (&http.Client{}).Do(req)

	if err != nil {
		return "", "", fmt.Errorf("Request failed for URL '%s': %v", cincinnatiFormattedURL, err)
	}

	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		err = fmt.Errorf("Failed reading body: %v", err)
		return
	}

	var cincinnatiReleases cincinnatiReleaseNodes
	var latestVersion *semver.Version
	var latestCincinnatiRelease cincinnatiRelease

	if err = json.Unmarshal(data, &cincinnatiReleases); err != nil {
		return "", "", fmt.Errorf("error decoding body of '%s': %v", data, err)
	}

	for _, release := range cincinnatiReleases.Nodes {
		currentVersion, err := semver.NewVersion(release.Version)

		if err != nil {
			log.Printf("Unable to parse version for %s, skipping", release.Version)
			continue
		}

		if latestVersion == nil || currentVersion.GreaterThan(latestVersion) {
			latestVersion = currentVersion
			latestCincinnatiRelease = release
		}
	}

	return ensureReleasePrefix(latestCincinnatiRelease.Version), latestCincinnatiRelease.Payload, nil
}

func ensureReleasePrefix(release string) string {
	if len(release) > 0 && !strings.Contains(release, "openshift-v") {
		log.Printf("Version %s didn't have prefix. Adding....", release)
		release = "openshift-v" + release
	}
	return release
}

// latestAccepted information from release controller.
type latestAccepted struct {
	Name        string `json:"name"`
	PullSpec    string `json:"pullSpec"`
	DownloadURL string `json:"downloadURL"`
}

type cincinnatiReleaseNodes struct {
	Nodes []cincinnatiRelease `json:"nodes"`
}
type cincinnatiRelease struct {
	Version string `json:"version"`
	Payload string `json:"payload"`
}
