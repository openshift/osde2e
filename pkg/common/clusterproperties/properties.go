package clusterproperties

// Common cluster properties
const (
	// MadeByOSDe2e property to attach to clusters.
	MadeByOSDe2e = "MadeByOSDe2e"

	// OwnedBy property which will tell who made the cluster.
	OwnedBy = "OwnedBy"

	// InstalledVersion property will tell which OSD version was installed in the cluster initially.
	InstalledVersion = "InstalledVersion"

	// UpgradeVersion property will tell which OSD version was installed in the cluster recently as an upgrade.
	UpgradeVersion = "UpgradeVersion"

	// Status the status for the cluster
	Status = "Status"

	// JobName is the name of job that is associated with the cluster.
	JobName = "JobName"

	// JobID is the name of the job ID that is associated with the cluster.
	JobID = "JobID"

	// ProvisionShardID is the shard ID that is set to provision a shard for the cluster.
	ProvisionShardID = "provision_shard_id"
)
