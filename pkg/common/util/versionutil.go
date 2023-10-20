package util

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/api/config/v1"
)

// NoVersionFound when no version can be found.
var NoVersionFound = "NoVersionFound"

// ClusterVersionProvider is a type that can return cluster version
// information. The *helper.H type implements this interface.
type ClusterVersionProvider interface {
	GetClusterVersion(context.Context) (*v1.ClusterVersion, error)
}

// OnSupportedVersionIt runs a ginkgo It() if and only if the cluster version meets the provided constraint.
// The cluster version is looked up using the provided helper.H.
func OnSupportedVersionIt(constraints *semver.Constraints, helper ClusterVersionProvider, description string, f func(context.Context)) {
	getVersion := func(ctx context.Context) *semver.Version {
		ver, err := helper.GetClusterVersion(ctx)
		Expect(err).ToNot(HaveOccurred())
		return semver.MustParse(ver.Status.Desired.Version)
	}

	ginkgo.It(description, func(ctx context.Context) {
		if !constraints.Check(getVersion(ctx)) {
			ginkgo.Skip("unsupported version")
		}
		f(ctx)
	})
}

func init() {
	// register new Versions here
}
