package upgrade

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	// format string for release stream latest
	latestReleaseURLFmt = "https://openshift-release.svc.ci.openshift.org/api/v1/releasestream/%s/latest"
)

// LatestRelease retrieves latest release information for given releaseStream.
func LatestRelease(releaseStream string) (name, pullSpec string, err error) {
	latestURL := fmt.Sprintf(latestReleaseURLFmt, releaseStream)
	resp, err := http.Get(latestURL)
	if err != nil {
		err = fmt.Errorf("failed to get latest for stream '%s': %v", releaseStream, err)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed reading body: %v", err)
		return
	}

	latest := latestAccepted{}
	if err = json.Unmarshal(data, &latest); err != nil {
		err = fmt.Errorf("error decoding body of '%s': %v", data, err)
	}

	return latest.Name, latest.PullSpec, nil
}

// latestAccepted information from release controller.
type latestAccepted struct {
	Name        string `json:"name"`
	PullSpec    string `json:"pullSpec"`
	DownloadURL string `json:"downloadURL"`
}
