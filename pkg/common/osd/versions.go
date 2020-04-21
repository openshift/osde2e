package osd

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	// query used to retrieve the current default version.
	defaultVersionSearch = "default = 't'"

	// VersionPrefix is the string that every OSD version begins with.
	VersionPrefix = "openshift-"

	// PageSize is the number of results to get per page from the cluster versions endpoint
	PageSize = 100

	// NoVersionFound is the value placed into a version string when no valid Cincinnati version can be selected.
	NoVersionFound = "NoVersionFound"
)

var (
	// Version440 represents Openshift version 4.4.0 and above
	Version440 *semver.Constraints

	prodDefaultVersion string

	once sync.Once = sync.Once{}
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

// DefaultVersionInProd will return the current default version in the production environment.
func DefaultVersionInProd() (string, error) {
	once.Do(func() {
		var OSD *OSD
		var err error

		// Create a new OCM client
		if OSD, err = New(config.Instance.OCM.Token, Environments.Choose("prod"), config.Instance.OCM.Debug); err != nil {
			log.Printf("error setting up OCM client for prod: %v", err)
			return
		}

		prodDefaultVersion, err = OSD.DefaultVersion()

		if err != nil {
			log.Printf("error getting default version from prod prod: %v", err)
			return
		}
	})

	if prodDefaultVersion == "" {
		return "", fmt.Errorf("unable to get default version in prod")
	}

	return prodDefaultVersion, nil
}

// DefaultVersion returns the default version currently offered by OSD.
func (u *OSD) DefaultVersion() (string, error) {
	var resp = &v1.VersionsListResponse{}

	err := retryer().Do(func() error {
		var err error
		resp, err = u.versions().List().
			Search(defaultVersionSearch).
			Size(1).
			Send()

		if err != nil {
			return err
		}

		if resp != nil && resp.Error() != nil {
			return errResp(resp.Error())
		}

		return nil
	})

	if err != nil {
		log.Print("error getting cluster versions from DefaultVersion.Response")
		log.Printf("Response Headers: %v", resp.Header())
		log.Printf("Response Error(s): %v", resp.Error())
		log.Printf("HTTP Code: %d", resp.Status())
		log.Printf("Size of response: %d", resp.Size())

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
		return NoVersionFound, nil
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

// NextReleaseAfterProdDefault will return the latest version of a release after the default in production.
// The integer "releasesFromProdDefault" is the "distance" from the current prod default to get the latest version for.
// With a prod default of 4.3.2, a releaseFromProdDefault value of 1 would produce 4.4.0-0.nightly... on integration.
// With a prod default of 4.3.2, a releaseFromProdDefault value of 2 would produce 4.5.0-0.nightly... on integration.
func (u *OSD) NextReleaseAfterProdDefault(releasesFromProdDefault int) (string, error) {
	currentProdDefault, err := DefaultVersionInProd()
	if err != nil {
		return "", err
	}

	currentProdDefaultSemver, err := OpenshiftVersionToSemver(currentProdDefault)
	if err != nil {
		return "", err
	}

	versionList, err := u.EnabledNoDefaultVersionList()
	if err != nil {
		return "", err
	}

	return nextReleaseAfterGivenVersionFromVersionList(currentProdDefaultSemver, versionList, releasesFromProdDefault)
}

// nextReleaseAfterGivenVersionFromVersionList will attempt to look for the next valid X.Y stream release, given a delta (releaseFromGivenVersion)
// Example In/Out
// In: 4.3.12, [4.3.13, 4.4.0, 4.5.0], 2
// Out: 4.5.0, nil
func nextReleaseAfterGivenVersionFromVersionList(givenVersion *semver.Version, versionList []string, releasesFromGivenVersion int) (string, error) {
	versionBuckets := map[string]string{}

	// Assemble a map that lists a release (x.y.0) to its latest version, with nightlies taking precedence over all else
	for _, version := range versionList {
		versionSemver, err := OpenshiftVersionToSemver(version)
		if err != nil {
			log.Printf("Unable to parse %s, skipping", version)
			continue
		}

		majorMinor := createMajorMinorStringFromSemver(versionSemver)
		if _, ok := versionBuckets[majorMinor]; !ok {
			versionBuckets[majorMinor] = version
		} else {
			currentGreatestVersion, err := OpenshiftVersionToSemver(versionBuckets[majorMinor])
			if err != nil {
				return "", err
			}

			versionIsNightly := strings.Contains(versionSemver.Prerelease(), "nightly")
			currentIsNightly := strings.Contains(currentGreatestVersion.Prerelease(), "nightly")

			// Make sure nightlies take precedence over other versions
			if versionIsNightly && !currentIsNightly {
				versionBuckets[majorMinor] = version
			} else if currentIsNightly && !versionIsNightly {
				continue
			} else if currentGreatestVersion.LessThan(versionSemver) {
				versionBuckets[majorMinor] = version
			}
		}
	}

	// Parse all major minor versions (x.y.0) into semver versions and place them in an array.
	// This is done explicitly so that we can utilize the semver library's sorting capability.
	majorMinorList := []*semver.Version{}
	for k := range versionBuckets {
		parsedMajorMinor, err := semver.NewVersion(k)
		if err != nil {
			return "", err
		}

		majorMinorList = append(majorMinorList, parsedMajorMinor)
	}

	sort.Sort(semver.Collection(majorMinorList))

	// Now that the list is sorted, we want to locate the major minor of the given version in the list.
	givenMajorMinor, err := semver.NewVersion(createMajorMinorStringFromSemver(givenVersion))

	if err != nil {
		return "", err
	}

	indexOfGivenMajorMinor := -1
	for i, majorMinor := range majorMinorList {
		if majorMinor.Equal(givenMajorMinor) {
			indexOfGivenMajorMinor = i
			break
		}
	}

	if indexOfGivenMajorMinor == -1 {
		return "", fmt.Errorf("unable to find current prod default in %s environment", config.Instance.OCM.Env)
	}

	// Next, we'll go the given version distance ahead of the given version. We want to do it this way instead of guessing
	// the next minor release so that we can handle major releases in the future, In other words, if the Openshift
	// 4.y line stops at 4.13, we'll still be able to pick 5.0 if it's the next release after 4.13.
	nextMajorMinorIndex := indexOfGivenMajorMinor + releasesFromGivenVersion

	if len(majorMinorList) <= nextMajorMinorIndex {
		return "", fmt.Errorf("there is no eligible next release on the %s environment", config.Instance.OCM.Env)
	}
	nextMajorMinor := createMajorMinorStringFromSemver(majorMinorList[nextMajorMinorIndex])

	if _, ok := versionBuckets[nextMajorMinor]; !ok {
		return "", fmt.Errorf("no major/minor version found for %s", nextMajorMinor)
	}

	return versionBuckets[nextMajorMinor], nil
}

func createMajorMinorStringFromSemver(version *semver.Version) string {
	return fmt.Sprintf("%d.%d", version.Major(), version.Minor())
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
		var resp *v1.VersionsListResponse
		err = retryer().Do(func() error {
			var err error

			resp, err = u.versions().List().Page(page).Size(PageSize).Send()

			if err != nil {
				return err
			}

			if resp != nil && resp.Error() != nil {
				return errResp(resp.Error())
			}

			return nil
		})

		if err != nil {
			err = fmt.Errorf("failed getting list of OSD versions: %v", err)
		} else if resp != nil {
			err = errResp(resp.Error())
		}

		if err != nil {
			log.Print("error getting cluster versions from getSemverList.Response")
			log.Printf("Response Headers: %v", resp.Header())
			log.Printf("Response Error(s): %v", resp.Error())
			log.Printf("HTTP Code: %d", resp.Status())
			log.Printf("Size of response: %d", resp.Size())

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
