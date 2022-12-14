package util

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	// VersionPrefix is the string that every OSD version begins with.
	VersionPrefix = "openshift-v"
)

// RandomStr returns a random varchar string given a specified length
func RandomStr(length int) (str string) {
	rand.Seed(time.Now().UnixNano())
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

// GinkgoIt wraps the 2.0 Ginkgo It function to allow for additional functionality.
func GinkgoIt(text string, body func(ctx context.Context), timeout ...float64) bool {
	if len(timeout) == 0 {
		timeout = append(timeout, viper.GetFloat64(config.Tests.PollingTimeout))
	}
	defer ginkgo.GinkgoRecover()
	return ginkgo.It(text, ginkgo.Offset(1), func(ctx context.Context) {
		done := make(chan interface{})
		go func() {
			defer ginkgo.GinkgoRecover()
			body(ctx)
			close(done)
		}()
		duration := time.Duration(5) * time.Second
		if len(timeout) > 0 {
			duration = time.Duration(timeout[0]) * time.Second
		}
		Eventually(done, duration).Should(BeClosed())
	})
}
