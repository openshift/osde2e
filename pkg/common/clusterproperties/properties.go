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

	// AWSAccount is the aws account associated with the cluster, if applicable.
	AWSAccount = "AWSAccount"

	// AdHocTestImages are the additional test images running on cluster
	AdHocTestImages = "AdHocTestImages"

	// JobID is the name of the job ID that is associated with the cluster.
	JobID = "JobID"

	// ProvisionShardID is the shard ID that is set to provision a shard for the cluster.
	ProvisionShardID = "provision_shard_id"

	// Availability is the availability for reserved/claimed/used clusters
	Availability = "Availability"

	// Reserved represents availability of a cluster ready to be claimed up by test job
	Reserved = "reserved"

	// Claimed represents the availability of a cluster claimed up by test job
	Claimed = "claimed"

	// Used represents the availability when a test job is finished on a cluster
	Used = "used"
)
