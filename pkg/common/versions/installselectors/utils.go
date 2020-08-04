package installselectors

import (
	"sort"

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

func sortVersions(availableVersions []*spi.Version) {
	sort.SliceStable(availableVersions, func(i, j int) bool {
		this := availableVersions[i]
		that := availableVersions[j]

		if this == nil || that == nil {
			return false
		}

		return this.Version().LessThan(that.Version())
	})
}
