package util

import (
	"math/rand"
	"strings"

	"github.com/Masterminds/semver/v3"
)

const (
	// VersionPrefix is the string that every OSD version begins with.
	VersionPrefix = "openshift-v"
)

// RandomStr returns a random varchar string given a specified length
func RandomStr(length int) (str string) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < length; i++ {
		c := string(chars[rand.Intn(len(chars))])
		str += c
	}
	return
}

// OpenshiftVersionToSemver converts an OpenShift version to a semver string which can then be used for comparisons.
func OpenshiftVersionToSemver(openshiftVersion string) (*semver.Version, error) {
	name := strings.TrimPrefix(openshiftVersion, VersionPrefix)
	return semver.NewVersion(name)
}

// SemverToOpenshiftVersion converts an OpenShift version to a semver string which can then be used for comparisons.
func SemverToOpenshiftVersion(version *semver.Version) string {
	return VersionPrefix + version.String()
}
