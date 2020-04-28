package spi

// UpgradeSource is the source that should be used when attempting upgrades.
type UpgradeSource string

const (
	// CincinnatiSource indicates that upgrades should use Cincinnati.
	CincinnatiSource UpgradeSource = "cincinnati"

	// ReleaseControllerSource indicates that upgrades should use the release controller.
	ReleaseControllerSource UpgradeSource = "release-controller"
)
