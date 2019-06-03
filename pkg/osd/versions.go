package osd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/openshift-online/uhc-sdk-go/pkg/client/clustersmgmt/v1"
	"github.com/openshift-online/uhc-sdk-go/pkg/client/errors"
)

// LatestPrerelease gets latest prerelease containing str for major and minor versions. Negative versions match all.
func (u *OSD) LatestPrerelease(major, minor int64, str string) (string, error) {
	resp, err := u.getVersionList()
	if err != nil {
		return "", fmt.Errorf("failed getting list of OSD versions: %v", err)
	} else if resp != nil {
		err = errResp(resp.Error())
	}

	if err != nil {
		return "", fmt.Errorf("couldn't retrieve available versions: %v", err)
	}

	// parse versions, filter for major+minor nightlies, then sort
	var versions []*semver.Version
	resp.Items().Each(func(v *v1.Version) bool {
		name := strings.TrimPrefix(v.ID(), "openshift-")
		if version, err := semver.NewVersion(name); err != nil {
			log.Printf("could not parse version '%s': %v", v.ID(), err)
		} else if version.Major() != major && major >= 0 {
			return true
		} else if version.Minor() != minor && minor >= 0 {
			return true
		} else if strings.Contains(version.Prerelease(), str) {
			versions = append(versions, version)
		}
		return true
	})

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions available with prerelease '%s' for '%d.%d'", str, major, minor)
	}

	// return latest nightly
	sort.Sort(semver.Collection(versions))
	latest := versions[len(versions)-1]
	return "openshift-" + latest.Original(), nil
}

// TODO: remove when retrieving versions using uhc-sdk-go is supported
func (u *OSD) getVersionList() (*versionListResponse, error) {
	resp := new(versionListResponse)

	// retrieve version list
	versionEndpoint := "/api/clusters_mgmt/v1/versions"
	if rawResp, err := u.conn.Get().Path(versionEndpoint).Send(); err != nil {
		return nil, err
	} else if rawResp.Status() != http.StatusOK {
		if resp.err, err = errors.UnmarshalError(err); err != nil {
			return nil, err
		}
		return resp, nil
	} else if err = json.Unmarshal(rawResp.Bytes(), resp); err != nil {
		return nil, err
	}

	// convert list into uhc-sdk-go types
	if data, err := resp.Raw.MarshalJSON(); err != nil {
		return nil, err
	} else if resp.list, err = v1.UnmarshalVersionList(data); err != nil {
		return nil, err
	}
	return resp, nil
}

type versionListResponse struct {
	Kind  string          `json:"kind"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
	Total int             `json:"total"`
	Raw   json.RawMessage `json:"items"`

	list *v1.VersionList
	err  *errors.Error
}

func (r *versionListResponse) Items() *v1.VersionList {
	return r.list
}
func (r *versionListResponse) Error() *errors.Error {
	return r.err
}
