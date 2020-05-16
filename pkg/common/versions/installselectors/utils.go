package installselectors

import "github.com/openshift/osde2e/pkg/common/spi"

func removeDefaultVersion(versions []*spi.Version) []*spi.Version {
	versionsWithoutDefault := []*spi.Version{}

	for _, version := range versions {
		if !version.Default() {
			versionsWithoutDefault = append(versionsWithoutDefault, version)
		}
	}

	return versionsWithoutDefault
}
