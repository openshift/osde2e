package osd

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/config"
)

const (
	// query used to retrieve the current default version.
	defaultVersionSearch = "default = 't'"

	// VersionPrefix is the string that every OSD version begins with.
	VersionPrefix = "openshift-"

	// PageSize is the number of results to get per page from the cluster versions endpoint
	PageSize = 100
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

// PreviousVersion returns the first available previous version for the given version.
func (u *OSD) PreviousVersion(verStr string) (string, error) {
	verStr = strings.TrimPrefix(verStr, VersionPrefix)
	vers, err := semver.NewVersion(verStr)
	if err != nil {
		return "", fmt.Errorf("couldn't parse given verStr '%s': %v", verStr, err)
	}

	versions, err := u.getSemverList(-1, -1, "")
	if err != nil {
		return "", fmt.Errorf("couldn't created sorted version list: %v", err)
	}

	for i := len(versions) - 1; i >= 0; i-- {
		v := versions[i]
		if v.LessThan(vers) {
			return VersionPrefix + v.Original(), nil
		}
	}
	return "", fmt.Errorf("no versions available before '%s'", verStr)
}

// LatestVersion gets latest release for major and minor versions. Negative versions match all.
func (u *OSD) LatestVersion(major, minor int64) (string, error) {
	suffix := ""

	if config.Cfg.OCM.Env == "int" {
		suffix = "nightly"
	}

	versions, err := u.getSemverList(major, minor, suffix)
	if err != nil {
		return "", fmt.Errorf("couldn't created sorted version list: %v", err)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions available for '%d.%d'", major, minor)
	}

	// return latest nightly
	latest := versions[len(versions)-1]
	return VersionPrefix + latest.Original(), nil
}

// EnabledNoDefaultVersionList returns a sorted list of the enabled but not default versions currently offered by OSD.
func (u *OSD) EnabledNoDefaultVersionList() ([]string, error) {
	semverVersions, err := u.getSemverList(-1, -1, "")
	if err != nil {
		return nil, fmt.Errorf("couldn't created sorted version list: %v", err)
	}

	defaultVersion, err := u.DefaultVersion()
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve the default version: %v", err)
	}

	var versionList []string
	for _, sv := range semverVersions {
		version := VersionPrefix + sv.Original()
		if strings.Compare(version, defaultVersion) == 0 {
			continue
		}
		versionList = append(versionList, version)
	}

	return versionList, nil
}

// getSemverList as sorted semvers containing str for major and minor versions. Negative versions match all.
func (u *OSD) getSemverList(major, minor int64, str string) (versions []*semver.Version, err error) {
	page := 1

	log.Printf("Querying cluster versions endpoint.")
	for {
		log.Printf("Getting page %d from the versions endpoint.", page)
		resp, err := u.versions().List().Page(page).Size(PageSize).Send()

		if err != nil {
			err = fmt.Errorf("failed getting list of OSD versions: %v", err)
		} else if resp != nil {
			err = errResp(resp.Error())
		}

		if err != nil {
			return versions, fmt.Errorf("couldn't retrieve available versions: %v", err)
		}

		// parse versions, filter for major+minor nightlies, then sort
		resp.Items().Each(func(v *v1.Version) bool {
			if version, err := u.OpenshiftVersionToSemver(v.ID()); err != nil {
				log.Printf("could not parse version '%s': %v", v.ID(), err)
			} else if version.Major() != major && major >= 0 {
				return true
			} else if version.Minor() != minor && minor >= 0 {
				return true
			} else if strings.Contains(version.Prerelease(), str) && v.Enabled() {
				versions = append(versions, version)
			}
			return true
		})

		// If we've looked at all the results, stop collecting them.
		if page*PageSize >= resp.Total() {
			break
		}
		page++
	}

	sort.Sort(semver.Collection(versions))
	return versions, nil
}

// OpenshiftVersionToSemver converts an OpenShift version to a semver string which can then be used for comparisons.
func (u *OSD) OpenshiftVersionToSemver(openshiftVersion string) (*semver.Version, error) {
	name := strings.TrimPrefix(openshiftVersion, VersionPrefix)
	return semver.NewVersion(name)
}
