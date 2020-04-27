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
	"github.com/openshift/osde2e/pkg/common/state"
	"github.com/openshift/osde2e/pkg/common/util"
)

const (
	// format string for release stream latest from release controller
	latestReleaseControllerURLFmt = "https://openshift-release.svc.ci.openshift.org/api/v1/releasestream/%s/latest"
	// format string for Cincinnati releases
	cincinnatiURLFmt = "https://api.openshift.com/api/upgrades_info/v1/graph?channel=%s&arch=amd64"
)

type smallCincinnatiCache struct {
	Cache map[string]smallCincinnatiCacheObject

	mutex sync.Mutex
}

type smallCincinnatiCacheObject struct {
	Versions []*semver.Version
	Edges    [][]int
}

// Get returns a specific channel from Cincinnati via our cache
func (s *smallCincinnatiCache) Get(channel string) (smallCincinnatiCacheObject, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Load Cincinnati data when we're trying to get information from a channel
	if _, ok := s.Cache[channel]; !ok {
		err := s.loadCincinnatiData(channel)

		if err != nil {
			return smallCincinnatiCacheObject{}, err
		}
	}

	return s.Cache[channel], nil
}

// loadCincinnatiData populates our cache with a given channel's data
func (s *smallCincinnatiCache) loadCincinnatiData(channel string) error {
	cincinnatiFormattedURL := fmt.Sprintf(cincinnatiURLFmt, channel)

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

	versions := []*semver.Version{}

	for _, release := range cincinnatiReleases.Nodes {
		currentVersion, err := semver.NewVersion(release.Version)

		if err != nil {
			log.Printf("Unable to parse version for %s, skipping", release.Version)
			continue
		}

		versions = append(versions, currentVersion)
	}

	s.Cache[channel] = smallCincinnatiCacheObject{
		Versions: versions,
		Edges:    cincinnatiReleases.Edges,
	}

	return nil
}

var cache *smallCincinnatiCache = &smallCincinnatiCache{
	Cache: map[string]smallCincinnatiCacheObject{},
	mutex: sync.Mutex{},
}

// DoesEdgeExistInCincinnati returns true if the version can be found in Cincinnati and the edge from the install version to the upgrade version exists.
func DoesEdgeExistInCincinnati(installVersion, upgradeVersion *semver.Version) (bool, error) {
	channel := VersionToChannel(upgradeVersion)
	cincinnatiVersions, err := cache.Get(channel)

	if err != nil {
		return false, fmt.Errorf("error loading Cincinnati data: %v", err)
	}

	installIndex := -1
	upgradeIndex := -1
	for i, cincinnatiVersion := range cincinnatiVersions.Versions {
		if installVersion.Equal(cincinnatiVersion) {
			installIndex = i
		}
		if upgradeVersion.Equal(cincinnatiVersion) {
			upgradeIndex = i
		}
	}

	targetEdge := []int{installIndex, upgradeIndex}

	for _, edge := range cincinnatiVersions.Edges {
		if len(edge) != len(targetEdge) {
			continue
		}

		match := true
		for i := range edge {
			if edge[i] != targetEdge[i] {
				match = false
				break
			}
		}

		if match {
			return true, nil
		}
	}

	return false, nil
}

// LatestReleaseFromReleaseController retrieves latest release information for given releaseStream on the release controller.
func LatestReleaseFromReleaseController(releaseStream string) (name, pullSpec string, err error) {
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
// If the config.Instance.Upgrade.OnlyUpgradeToZReleases flag is set, this will use the install version
// in the global state object to determine the channel.
// If production is targeted for this cluster provision, the stable channel will be used.
// If stage is targeted for this cluster provision, the fast channel will be used.
// Regardless of environment, if Prerelease is populated in the version object, then the candidate channel will be used.
func VersionToChannel(version *semver.Version) string {
	useVersion := version
	if config.Instance.Upgrade.OnlyUpgradeToZReleases {
		var err error
		useVersion, err = util.OpenshiftVersionToSemver(state.Instance.Cluster.Version)

		if err != nil {
			panic("cluster version stored in state object is invalid")
		}
	}

	if strings.HasPrefix(useVersion.Prerelease(), "rc") {
		return fmt.Sprintf("candidate-%d.%d", useVersion.Major(), useVersion.Minor())
	}

	environment := config.Instance.OCM.Env

	if environment == "stage" {
		return fmt.Sprintf("fast-%d.%d", useVersion.Major(), useVersion.Minor())
	}

	return fmt.Sprintf("stable-%d.%d", useVersion.Major(), useVersion.Minor())
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
	Edges [][]int             `json:"edges"`
}

type cincinnatiRelease struct {
	Version string `json:"version"`
	Payload string `json:"payload"`
}
