package osd

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/openshift-online/uhc-sdk-go/pkg/client/clustersmgmt/v1"
)

const (
	// query used to retrieve the current default version.
	defaultVersionSearch = "default = 't'"
)

// DefaultVersion returns the default version currently offered by OSD.
func (u *OSD) DefaultVersion() (string, error) {
	resp, err := u.versions().List().
		Search(defaultVersionSearch).
		Size(1).
		Send()
	if err == nil && resp != nil {
		err = errResp(resp.Error())
	}

	if err != nil {
		return "", fmt.Errorf("couldn't retrieve available versions: %v", err)
	}

	version := resp.Items().Get(0)
	if version == nil {
		return "", errors.New("version returned was nil")
	}

	return version.ID(), nil
}

// LatestPrerelease gets latest prerelease containing str for major and minor versions. Negative versions match all.
func (u *OSD) LatestPrerelease(major, minor int64, str string) (string, error) {
	resp, err := u.versions().List().Send()
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
