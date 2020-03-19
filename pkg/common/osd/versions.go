package osd

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const (
	// query used to retrieve the current default version.
	defaultVersionSearch = "default = 't'"

	// VersionPrefix is the string that every OSD version begins with.
	VersionPrefix = "openshift-"

	// PageSize is the number of results to get per page from the cluster versions endpoint
	PageSize = 100
)

var (
	// Version440 represents Openshift version 4.4.0 and above
	Version440 *semver.Constraints
)

func init() {
	var err error
	Version440, err = semver.NewConstraint(">= 4.4.0-0")

	if err != nil {
		panic(err)
	}
}

// Use for the semver list filter to include all results.
func noFilter(_ *semver.Version) bool {
	return true
}

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

	versions, err := u.getSemverList(-1, -1, "", noFilter)
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

// LatestVersion gets latest release on OSD.
func (u *OSD) LatestVersion() (string, error) {
	return u.LatestVersionWithTarget(-1, -1, "")
}

// LatestVersionWithFilter gets latest release on OSD and applies the given filter to it.
func (u *OSD) LatestVersionWithFilter(filter func(*semver.Version) bool) (string, error) {
	versions, err := u.getSemverList(-1, -1, "", filter)
	if err != nil {
		return "", fmt.Errorf("couldn't created sorted version list: %v", err)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions available after applying filter")
	}

	latest := versions[len(versions)-1]
	return VersionPrefix + latest.Original(), nil
}

// LatestVersionWithTarget gets latest release for major and minor versions. Negative versions match all.
func (u *OSD) LatestVersionWithTarget(major, minor int64, suffix string) (string, error) {
	versions, err := u.getSemverList(major, minor, suffix, noFilter)
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

// MiddleVersion gets the middle version in the ordered list of cluster image sets known to OCM. It will also return true if there were enough versions to select from.
func (u *OSD) MiddleVersion() (string, bool, error) {
	versionList, err := u.EnabledNoDefaultVersionList()
	if err != nil {
		return "", false, err
	}

	if len(versionList) <= 1 {
		return "", false, nil
	}

	return versionList[len(versionList)/2], true, nil
}

// OldestVersion gets the middle version in the ordered list of cluster image sets known to OCM. It will also return true if there were enough versions to select from.
func (u *OSD) OldestVersion() (string, bool, error) {
	versionList, err := u.EnabledNoDefaultVersionList()
	if err != nil {
		return "", false, err
	}

	if len(versionList) <= 1 {
		return "", false, nil
	}

	return versionList[0], true, nil
}

// EnabledNoDefaultVersionList returns a sorted list of the enabled but not default versions currently offered by OSD.
func (u *OSD) EnabledNoDefaultVersionList() ([]string, error) {
	semverVersions, err := u.getSemverList(-1, -1, "", noFilter)
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
func (u *OSD) getSemverList(major, minor int64, str string, filter func(*semver.Version) bool) (versions []*semver.Version, err error) {
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
			if version, err := OpenshiftVersionToSemver(v.ID()); err != nil {
				log.Printf("could not parse version '%s': %v", v.ID(), err)
			} else if version.Major() != major && major >= 0 {
				return true
			} else if version.Minor() != minor && minor >= 0 {
				return true
			} else if strings.Contains(version.Prerelease(), str) && filter(version) && v.Enabled() {
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
func OpenshiftVersionToSemver(openshiftVersion string) (*semver.Version, error) {
	name := strings.TrimPrefix(openshiftVersion, VersionPrefix)
	return semver.NewVersion(name)
}
