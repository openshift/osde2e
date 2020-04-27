// Package spi defines the service provider interface for cluster providers.
package spi

// Provisioner is the interface that must be implemented in order to provision clusters in osde2e.
type Provisioner interface {
	// LaunchCluster creates a new cluster and returns the cluster ID.
	LaunchCluster() (string, error)

	// DeleteCluster deletes a cluster.
	DeleteCluster(clusterID string) error

	// GetCluster gets a cluster.
	GetCluster(clusterID string) (*Cluster, error)

	// ClusterKubeconfig should return the raw kubeconfig for the cluster.
	ClusterKubeconfig(clusterID string) ([]byte, error)

	// CheckQuota will return true if there is enough quota to provision a cluster.
	CheckQuota() (bool, error)

	// InstallAddons will install addons onto the cluster.
	InstallAddons(clusterID string, addonIDs []string) (int, error)

	// AvailableVersions returns a sorted list of versions.
	AvailableVersions() ([]Version, error)

	// Logs will get logs relevant to the cluster from the provisioner.
	Logs(clusterID string) (logs map[string][]byte, err error)
}
