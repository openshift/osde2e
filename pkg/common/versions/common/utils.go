package common

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// NextReleaseAfterGivenVersionFromVersionList will attempt to look for the next valid X.Y stream release, given a delta (releaseFromGivenVersion)
// Example In/Out
// In: 4.3.12, [4.3.13, 4.4.0, 4.5.0], 2
// Out: 4.5.0, nil
func NextReleaseAfterGivenVersionFromVersionList(givenVersion *semver.Version, versionList []*spi.Version, releasesFromGivenVersion int) (*semver.Version, error) {
	versionBuckets := map[string]*semver.Version{}

	// Assemble a map that lists a release (x.y.0) to its latest version, with nightlies taking precedence over all else
	for _, version := range versionList {
		versionSemver := version.Version()
		majorMinor := createMajorMinorStringFromSemver(versionSemver)
		if _, ok := versionBuckets[majorMinor]; !ok {
			versionBuckets[majorMinor] = versionSemver
		} else {
			currentGreatestVersion := versionBuckets[majorMinor]
			versionIsNightly := strings.Contains(versionSemver.Prerelease(), "nightly")
			currentIsNightly := strings.Contains(currentGreatestVersion.Prerelease(), "nightly")

			// Make sure nightlies take precedence over other versions
			if versionIsNightly && !currentIsNightly {
				versionBuckets[majorMinor] = versionSemver
			} else if currentIsNightly && !versionIsNightly {
				continue
			} else if currentGreatestVersion.LessThan(versionSemver) {
				versionBuckets[majorMinor] = versionSemver
			}
		}
	}

	// Parse all major minor versions (x.y.0) into semver versions and place them in an array.
	// This is done explicitly so that we can utilize the semver library's sorting capability.
	majorMinorList := []*semver.Version{}
	for k := range versionBuckets {
		parsedMajorMinor, err := semver.NewVersion(k)
		if err != nil {
			return nil, err
		}

		majorMinorList = append(majorMinorList, parsedMajorMinor)
	}

	sort.Sort(semver.Collection(majorMinorList))

	// Now that the list is sorted, we want to locate the major minor of the given version in the list.
	givenMajorMinor, err := semver.NewVersion(createMajorMinorStringFromSemver(givenVersion))
	if err != nil {
		return nil, err
	}

	indexOfGivenMajorMinor := -1
	for i, majorMinor := range majorMinorList {
		if majorMinor.Equal(givenMajorMinor) {
			indexOfGivenMajorMinor = i
			break
		}
	}

	if indexOfGivenMajorMinor == -1 {
		return nil, fmt.Errorf("unable to find given version from list of available versions")
	}

	// Next, we'll go the given version distance ahead of the given version. We want to do it this way instead of guessing
	// the next minor release so that we can handle major releases in the future, In other words, if the Openshift
	// 4.y line stops at 4.13, we'll still be able to pick 5.0 if it's the next release after 4.13.
	nextMajorMinorIndex := indexOfGivenMajorMinor + releasesFromGivenVersion

	if len(majorMinorList) <= nextMajorMinorIndex {
		return nil, fmt.Errorf("there is no eligible next release from the list of available versions")
	}
	nextMajorMinor := createMajorMinorStringFromSemver(majorMinorList[nextMajorMinorIndex])

	if _, ok := versionBuckets[nextMajorMinor]; !ok {
		return nil, fmt.Errorf("no major/minor version found for %s", nextMajorMinor)
	}

	return versionBuckets[nextMajorMinor], nil
}

// SortVersions accepts a pointer to a list of spi.Versions and sorts it
func SortVersions(availableVersions []*spi.Version) {
	sort.SliceStable(availableVersions, func(i, j int) bool {
		this := availableVersions[i]
		that := availableVersions[j]

		if this == nil || that == nil {
			return false
		}

		return this.Version().LessThan(that.Version())
	})
}

func createMajorMinorStringFromSemver(version *semver.Version) string {
	if version == nil {
		return ""
	}
	return fmt.Sprintf("%d.%d", version.Major(), version.Minor())
}
