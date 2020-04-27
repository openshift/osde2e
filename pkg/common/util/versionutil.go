package util

import (
	"github.com/Masterminds/semver"
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
