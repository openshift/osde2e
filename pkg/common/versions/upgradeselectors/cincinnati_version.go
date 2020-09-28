package upgradeselectors

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
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/spf13/viper"
)

const (
	// format string for Cincinnati releases
	cincinnatiURLFmt = "https://api.openshift.com/api/upgrades_info/v1/graph?channel=%s&arch=amd64"
)

func init() {
	registerSelector(cincinnatiUpgrade{})
}

// cincinnatiUpgrade will select an upgrade target based on Cincinnaati.
type cincinnatiUpgrade struct{}

func (c cincinnatiUpgrade) ShouldUse(upgradeSource spi.UpgradeSource) bool {
	return upgradeSource == spi.CincinnatiSource && viper.GetBool(config.Upgrade.UpgradeToCISIfPossible)
}

func (c cincinnatiUpgrade) Priority() int {
	return 40
}

func (c cincinnatiUpgrade) SelectVersion(installVersion *semver.Version, versionList *spi.VersionList) (string, string, error) {
	var filteredVersionList = []*semver.Version{}

	for _, version := range versionList.AvailableVersions() {
		if filterOnCincinnati(installVersion, version.Version()) {
			filteredVersionList = append(filteredVersionList, version.Version())
		}
	}

	numResults := len(filteredVersionList)
	if numResults == 0 {
		viper.Set(config.Upgrade.ReleaseName, util.NoVersionFound)
		metadata.Instance.SetUpgradeVersionSource("none")
		return "", "", nil
	}

	cisUpgradeVersion := filteredVersionList[numResults-1]

	releaseName := ""
	// If the available cluster image set makes sense, then we'll just use that
	if !cisUpgradeVersion.LessThan(installVersion) {
		releaseName = util.SemverToOpenshiftVersion(cisUpgradeVersion)
		metadata.Instance.SetUpgradeVersionSource("cluster image set")
		viper.Set(config.Upgrade.UpgradeVersionEqualToInstallVersion, cisUpgradeVersion.Equal(installVersion))
	}

	return releaseName, "", nil
}

func filterOnCincinnati(installVersion *semver.Version, upgradeVersion *semver.Version) bool {
	versionInCincinnati, err := doesEdgeExistInCincinnati(installVersion, upgradeVersion)

	if err != nil {
		log.Printf("error while trying to filter on version in Cincinnati: %v", err)
		return false
	}

	return versionInCincinnati
}

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

// doesEdgeExistInCincinnati returns true if the version can be found in Cincinnati and the edge from the install version to the upgrade version exists.
func doesEdgeExistInCincinnati(installVersion, upgradeVersion *semver.Version) (bool, error) {
	channel, err := upgrade.VersionToChannel(upgradeVersion)
	if err != nil {
		return false, fmt.Errorf("error getting channel from provided version: %v", err)
	}

	cincinnatiVersions, err := cache.Get(channel)

	if err != nil {
		return false, fmt.Errorf("error loading Cincinnati data: %v", err)
	}

	if !strings.Contains(channel, "stable") {
		if i := strings.LastIndex(upgradeVersion.Original(), "-"); i != -1 {
			upgradeVersion, _ = semver.NewVersion(upgradeVersion.Original()[:i])
		}
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

	for _, edge := range cincinnatiVersions.Edges {
		if edge[0] == installIndex && edge[1] == upgradeIndex {
			return true, nil
		}
	}

	return false, nil
}

type cincinnatiReleaseNodes struct {
	Nodes []cincinnatiRelease `json:"nodes"`
	Edges [][]int             `json:"edges"`
}

type cincinnatiRelease struct {
	Version string `json:"version"`
	Payload string `json:"payload"`
}
