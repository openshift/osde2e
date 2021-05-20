package clusterproperties

const (
	// StatusProvisioning represents when the cluster has just been provisioned.
	StatusProvisioning = "provisioning"

	// StatusWaitingForReady represents the cluster currently waiting for ready.
	StatusWaitingForReady = "waiting-for-ready"

	// StatusHealthCheck represents the cluster checking for health.
	StatusHealthCheck = "health-check"

	// StatusHealthy represents the cluster passing the health check.
	StatusHealthy = "healthy"

	// StatusUnhealthy represents the cluster failing the health check.
	StatusUnhealthy = "unhealthy"

	// StatusUpgrading represents the cluster upgrading.
	StatusUpgrading = "upgrading"

	// StatusUpgradeHealthCheck represents the cluster checking for health during an upgrade.
	StatusUpgradeHealthCheck = "upgrade-health-check"

	// StatusUpgradeHealthy represents the upgraded cluster passing the health check.
	StatusUpgradeHealthy = "upgrade-healthy"

	// StatusUpgradeUnhealthy represents the upgraded cluster failing the health check.
	StatusUpgradeUnhealthy = "upgrade-unhealthy"

	// StatusUninstalling represents the cluster uninstalling.
	StatusUninstalling = "uninstalling"

	// StatusCompleted represents the cluster having finished its CI work and awaiting teardown.
	StatusCompleted = "completed"

	// StatusCompletedPassing represents the cluster having finished its CI and tests having passed
	StatusCompletedPassing = "completed-passing"

	// StatusCompletedFailing represents the cluster having finished its CI and tests having failed
	StatusCompletedFailing = "completed-failing"

	// StatusCompletedError represents the cluster that exhibits issues outside of the test results
	StatusCompletedError = "completed-error"

	// StatusResuming represents the cluster having just been woken up
	StatusResuming = "resuming"
)
