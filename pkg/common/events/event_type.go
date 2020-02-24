package events

// EventType is the type of event to record
type EventType string

// This file defines different event types. Please add new types here so that we can easily track them.
const (
	// ------ Cluster provisioning events

	// InstallSuccessful when the cluster has successfully provisioned
	InstallSuccessful EventType = "InstallSuccessful"

	// FailedClusterProvision when the cluster has not been successfully provisioned
	InstallFailed EventType = "InstallFailed"

	// UpgradeSuccessful when the upgrade was successful
	UpgradeSuccessful EventType = "UpgradeSuccessful"

	// UpgradeFailed when the upgrade failed
	UpgradeFailed EventType = "UpgradeFailed"

	// NoHiveLogs when no logs from Hive were collected after a cluster provisioning event
	NoHiveLogs EventType = "NoHiveLogs"

	// ------ Addon installation events

	// InstallAddonsSuccessful when the addons installed successfully
	InstallAddonsSuccessful EventType = "InstallAddonsSuccessful"

	// InstallAddonsFailed when the addons failed to install
	InstallAddonsFailed EventType = "InstallAddonsFailed"
)
