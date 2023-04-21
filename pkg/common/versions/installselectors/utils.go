package installselectors

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/spi"
)

func removeDefaultVersion(versions []*spi.Version) []*spi.Version {
	versionsWithoutDefault := []*spi.Version{}

	for _, version := range versions {
		if !version.Default() {
			versionsWithoutDefault = append(versionsWithoutDefault, version)
		}
	}

	return versionsWithoutDefault
}

// findDefaultVersion returns the default version from the supplied versions
func findDefaultVersion(versions []*spi.Version) (*spi.Version, error) {
	for _, version := range versions {
		if version.Default() {
			return version, nil
		}
	}
	return nil, fmt.Errorf("no default version found")
}
