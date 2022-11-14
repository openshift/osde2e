package util

import (
	"context"

	"github.com/Masterminds/semver"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/api/config/v1"
)

var (
	// NoVersionFound when no version can be found.
	NoVersionFound = "NoVersionFound"

	// Version420 represents Openshift version 4.2.0 and above
	Version420 *semver.Constraints

	// Version440 represents Openshift version 4.4.0 and above
	Version440 *semver.Constraints

	// Version460 represents Openshift version 4.6.0 and above
	Version460 *semver.Constraints
)

// ClusterVersionProvider is a type that can return cluster version
// information. The *helper.H type implements this interface.
type ClusterVersionProvider interface {
	GetClusterVersion(context.Context) (*v1.ClusterVersion, error)
}

// OnSupportedVersionIt runs a ginkgo It() if and only if the cluster version meets the provided constraint.
// The cluster version is looked up using the provided helper.H.
func OnSupportedVersionIt(constraints *semver.Constraints, helper ClusterVersionProvider, description string, f func(context.Context), timeout float64) {
	getVersion := func(ctx context.Context) *semver.Version {
		ver, err := helper.GetClusterVersion(ctx)
		Expect(err).ToNot(HaveOccurred())
		return semver.MustParse(ver.Status.Desired.Version)
	}

	GinkgoIt(description, func(ctx context.Context) {
		if !constraints.Check(getVersion(ctx)) {
			ginkgo.Skip("unsupported version")
		}
		f(ctx)
	}, timeout)
}

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

	Version460, err = semver.NewConstraint(">= 4.6.0-0")
	if err != nil {
		panic(err)
	}
}
