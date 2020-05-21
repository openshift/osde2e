package installselectors

import (
	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// Interface is the interface for version selection implementations for installs.
type Interface interface {
	// ShouldUse will return true if the version selector should be used.
	ShouldUse() bool

	// Priority is the integer priority for the selector. The higher the integer,
	// the higher the priority. 0 is the minimum.
	Priority() int

	// SelectVersion will select a version to install.
	SelectVersion(versionList *spi.VersionList) (*semver.Version, string, error)
}
