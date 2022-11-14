// Package spi defines the service provider interface for cluster providers.
package spi

import (
	"time"
)

// AddOnID is a string used as the identifier for an addon
type AddOnID = string

// AddOnParams is a key-value store of parameters for an addon's installation
type AddOnParams = map[string]string

// Provider is the interface that must be implemented in order to provision clusters in osde2e.
type Provider interface {
	// IsValidClusterName validates that the proposed name used for creating the cluster.
	//
	// Currently this validates if the proposed clusterName already exists before attempting to
	// create and cycling on OCM errors.
	IsValidClusterName(clusterName string) (bool, error)

	// LaunchCluster creates a new cluster and returns the cluster ID.
	//
	// This is expected to kick off the cluster provisioning process and
	// report back an identifier without waiting. Subsequent calls within OSDe2e will
	// use the status reported by GetCluster to determine the provision state of
	// the cluster and wait for it to start before running tests.
	LaunchCluster(clusterName string) (string, error)

	// DeleteCluster deletes a cluster.
	//
	// Calling this will start a cluster deletion and return as soon as the process
	// has begun. OSDe2e will not wait for the cluster to delete.
	DeleteCluster(clusterID string) error

	// ScaleCluster scales a cluster.
	//
	// This will start a cluster scaling operation. This should grow or shrink a cluster to
	// the desired number of compute nodes. This is expected to kick off the cluster scaling
	// process and return without waiting. Subsequent calls within OSDe2e will use the status
	// reported by GetCluster and status calls from the cluster itself to determine if the
	// scaling has finished.
	ScaleCluster(clusterID string, numComputeNodes int) error

	// ListCluster lists clusters from a provider based on a SQL-like query.
	ListClusters(query string) ([]*Cluster, error)

	// GetCluster gets a cluster.
	//
	// This is what OSDe2e will use to gather cluster information, including whether
	// the cluser has finished provisioning.
	GetCluster(clusterID string) (*Cluster, error)

	// ClusterKubeconfig should return the raw kubeconfig for the cluster.
	//
	// OSDe2e needs administrative cluster level access for a cluster, so this should
	// return a raw kubeconfig that will allow OSDe2e to connect with administrative
	// access.
	ClusterKubeconfig(clusterID string) ([]byte, error)

	// CheckQuota will return true if there is enough quota to provision a cluster.
	//
	// To prevent a provisioning attempt, OSDe2e will first check the quota first. This quota
	// is currently expected to be configured by the global config object.
	CheckQuota(sku string) (bool, error)

	// InstallAddons will install addons onto the cluster.
	//
	// OpenShift dedicated has the notion of addon installation, which users can request from
	// the OCM API. If you wish to emulate this support, the provider will need to support a similar
	// mechanism.
	InstallAddons(clusterID string, addonIDs []AddOnID, params map[AddOnID]AddOnParams) (int, error)

	// Versions returns a sorted list of supported OpenShift versions.
	//
	// A version list of the available OpenShift versions supported by this provider. One of the
	// versions is expected to be labeled as "default." The provider can also set a default version
	// override, which is useful if you want to select relative versions to test against, e.g.
	// 4.3.12 + nightly of next release == 4.4.0-0.nightly.
	Versions() (*VersionList, error)

	// Logs will get logs relevant to the cluster from the provider.
	//
	// Any provider level logs that are relevant to the cluster.
	Logs(clusterID string) (map[string][]byte, error)

	// Metrics will get metrics relevant to the cluster from the provider.
	Metrics(clusterID string) (bool, error)

	// Environment retrives the environment from the provider.
	//
	// This is for providers that have situations like "integration," "stage," and "production"
	// environments. There's no restriction on what values are expected to be returned here.
	Environment() string

	// UpgradeSource is what upgrade source to use when attempting an upgrade.
	//
	// This returns what OSDe2e should use to try to select an upgrade version. Right now only
	// Cincinnati and the release controller are supported.
	UpgradeSource() UpgradeSource

	// CincinnatiChannel is the Cincinnati channel to use for upgrades (where applicable).
	//
	// If the upgrade channel uses a Cincinnati source, this will dictate what channel should be
	// used. This is only a prefix, so "fast," "stable," etc.
	CincinnatiChannel() CincinnatiChannel

	// Type is the Provider type, specific to each Plugin
	//
	// This simply returns the name of the Provider
	Type() string

	// ExtendExpiry extends the expiration time of an existing cluster.
	ExtendExpiry(clusterID string, hours uint64, minutes uint64, seconds uint64) error

	// Expire sets the expiration of an existing cluster to the current time.
	Expire(clusterID string) error

	// AddProperty adds a new property to the properties field of an existing cluster.
	AddProperty(cluster *Cluster, tag string, value string) error

	// Upgrade requests the provider initiate a cluster upgrade to the given version
	Upgrade(clusterID string, version string, t time.Time) error

	// GetUpgradePolicyID gets the first upgrade policy from the top
	GetUpgradePolicyID(clusterID string) (string, error)

	// UpdateSchedule updates the existing upgrade policy for re-scheduling
	UpdateSchedule(clusterID string, version string, t time.Time, policyID string) error

	// DetermineMachineType selects a random machine type for a given cluster.
	DetermineMachineType(cloudProvider string) (string, error)

	// Hibernate triggers a hibernation of the cluster
	// If hibernation is unsupported by the provider, it will log that it's unsupported
	// but still return True.
	Hibernate(clusterID string) bool

	// Resume triggers a hibernated cluster to wake up
	// If hibernation is unsupported by the provider, it will log that it's unsupported
	// but still return True.
	Resume(clusterID string) bool

	// AddClusterProxy adds a cluster-wide proxy to the cluster.
	AddClusterProxy(clusterId string, httpsProxy string, httpProxy string, userCABundle string) error

	// RemoveClusterProxy removes the cluster proxy configuration for the supplied cluster
	RemoveClusterProxy(clusterId string) error

	// RemoveUserCABundle removes only the Additional Trusted CA Bundle from the cluster
	RemoveUserCABundle(clusterId string) error

	// LoadUserCaBundleData loads CA contents from CA cert file
	LoadUserCaBundleData(file string) (string, error)
}
