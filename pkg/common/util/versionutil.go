package util

import (
	"github.com/Masterminds/semver"
)

var (
	// NoVersionFound when no version can be found.
	NoVersionFound = "NoVersionFound"

	// Version420 represents Openshift version 4.2.0 and above
	Version420 *semver.Constraints

	// Version440 represents Openshift version 4.4.0 and above
	Version440 *semver.Constraints
)

func init() {
	var err error

	Version420, err = semver.NewConstraint(">= 4.2.0-0")
	if err != nil {
		panic(err)
	}

	Version440, err = semver.NewConstraint(">= 4.4.0-0")

	if err != nil {
		panic(err)
	}
}
