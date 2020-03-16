package upgrade

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/osd"
)

const (
	// format string for release stream latest from release controller
	latestReleaseControllerURLFmt = "https://openshift-release.svc.ci.openshift.org/api/v1/releasestream/%s/latest"
	// format string for Cincinnati releases
	cincinnatiURLFmt = "%s/api/upgrades_info/v1/graph?channel=%s&arch=amd64"
)

type smallCincinnatiCache struct {
	Cache map[string][]*semver.Version

	mutex sync.Mutex
}

func (s *smallCincinnatiCache) Get(channel string) ([]*semver.Version, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Load Cincinnati data when we're trying to get information from a channel
	if _, ok := s.Cache[channel]; !ok {
		err := s.loadCincinnatiData(channel)

		if err != nil {
			return nil, err
		}
	}

	return s.Cache[channel], nil
}

func (s *smallCincinnatiCache) loadCincinnatiData(channel string) error {
	cincinnatiFormattedURL := fmt.Sprintf(cincinnatiURLFmt, osd.Environments.Choose(config.Instance.OCM.Env), channel)

	var req *http.Request

	// Cincinnati requires an Accept header, so we add it in here
	req, err := http.NewRequest("GET", cincinnatiFormattedURL, nil)
	req.Header.Set("Accept", "application/json")

	if err != nil {
		return fmt.Errorf("failed to create Cincinnati request for URL '%s': %v", cincinnatiFormattedURL, err)
	}

	resp, err := (&http.Client{}).Do(req)

	if err != nil {
		return fmt.Errorf("Request failed for URL '%s': %v", cincinnatiFormattedURL, err)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		err = fmt.Errorf("Failed reading body: %v", err)
		return err
	}

	var cincinnatiReleases cincinnatiReleaseNodes

	if err = json.Unmarshal(data, &cincinnatiReleases); err != nil {
		return fmt.Errorf("error decoding body of '%s': %v", data, err)
	}

	s.Cache[channel] = []*semver.Version{}
	for _, release := range cincinnatiReleases.Nodes {
		currentVersion, err := semver.NewVersion(release.Version)

		if err != nil {
			log.Printf("Unable to parse version for %s, skipping", release.Version)
			continue
		}

		s.Cache[channel] = append(s.Cache[channel], currentVersion)
	}

	return nil
}

var cache *smallCincinnatiCache = &smallCincinnatiCache{
	Cache: map[string][]*semver.Version{},
	mutex: sync.Mutex{},
}

// IsVersionInCincinnati returns true if the version can be found in Cincinnati
func IsVersionInCincinnati(version *semver.Version) (bool, error) {
	channel := VersionToChannel(version)
	cincinnatiVersions, err := cache.Get(channel)

	if err != nil {
		return false, fmt.Errorf("error loading Cincinnati data: %v", err)
	}

	for _, cincinnatiVersion := range cincinnatiVersions {
		if version.Equal(cincinnatiVersion) {
			return true, nil
		}
	}

	return false, nil
}

// LatestReleaseFromReleaseController retrieves latest release information for given releaseStream on the release controller.
func LatestReleaseFromReleaseController(releaseStream string, useReleaseControllerForInt bool) (name, pullSpec string, err error) {
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

	return ensureReleasePrefix(latest.Name), latest.PullSpec, nil
}

// VersionToChannel creates a Cincinnati channel version out of an OpenShift version.
func VersionToChannel(version *semver.Version) string {
	if strings.HasPrefix(version.Prerelease(), "rc") {
		return fmt.Sprintf("candidate-%d.%d", version.Major(), version.Minor())
	}

	return fmt.Sprintf("fast-%d.%d", version.Major(), version.Minor())
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
