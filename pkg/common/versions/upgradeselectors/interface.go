package upgradeselectors

import (
	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// Interface is the interface for version selection implementations for upgrades.
type Interface interface {
	// ShouldUse will return true if the version selector should be used.
	ShouldUse(upgradeSource spi.UpgradeSource) bool

	// Priority is the integer priority for the selector. The higher the integer,
	// the higher the priority. 0 is the minimum.
	Priority() int

	// SelectVersion will select a version to upgrade. This will be populated as a release name and an image.
	// If the image is blank, OpenShift will use Cincinnati to attempt to upgrade.
	SelectVersion(installVersion *semver.Version, versionList *spi.VersionList) (string, string, error)
}
