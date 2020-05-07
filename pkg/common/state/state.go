// Package state provides common state across osde2e
package state

import "github.com/openshift/osde2e/pkg/common/spi"

// Instance is the global state for osde2e runs
var Instance = new(State)

// State dictates the behavior of cluster tests.
type State struct {
	Cluster ClusterState `yaml:"cluster"`

	CloudProvider CloudProviderState `yaml:"cloudProvider"`

	Kubeconfig KubeconfigState `yaml:"kubeconfig"`

	Upgrade UpgradeState `yaml:"upgrade"`

	// InstalledWorkloads is an internal variable used to track currently installed workloads in this test run.
	InstalledWorkloads map[string]string

	// Phase is an internal variable used to track the current set of tests being run (install, upgrade).
	Phase string

	// Project is both the project and SA automatically created to house all objects created during an osde2e-run
	Project string
}

// CloudProviderState contains state information pertaining to which cloud provider to use for cluster provisioning.
type CloudProviderState struct {
	// CloudProviderID is the cloud provider ID to use to provision the cluster.
	CloudProviderID string `env:"CLOUD_PROVIDER_ID" sect:"cloudProvider" default:"aws" yaml:"providerId"`

	// Region is the cloud provider region to use to provision the cluster.
	Region string `env:"CLOUD_PROVIDER_REGION" sect:"cloudProvider" default:"us-east-1" yaml:"region"`
}

// ClusterState contains state information about the active cluster.
type ClusterState struct {
	// ID identifies the cluster. If set at start, an existing cluster is tested.
	ID string `json:"cluster_id,omitempty" env:"CLUSTER_ID" sect:"cluster" yaml:"id"`

	// Name is the name of the cluster being created.
	Name string `json:"cluster_name,omitempty" env:"CLUSTER_NAME" sect:"cluster" yaml:"name"`

	// Version is the version of the cluster being deployed.
	Version string `json:"cluster_version,omitempty" env:"CLUSTER_VERSION" sect:"version"  yaml:"version"`

	// EnoughVersionsForOldestOrMiddleTest is true if there were enough versions for an older/middle test.
	EnoughVersionsForOldestOrMiddleTest bool `default:"true"`

	// PreviousVersionFromDefaultFound is true if a previous version from default was found.
	PreviousVersionFromDefaultFound bool `default:"true"`

	// State is the cluster state observed by OCM.
	State spi.ClusterState `json:"cluster_state,omitempty" yaml:"state"`
}

// KubeconfigState stores information required to talk to the Kube API
type KubeconfigState struct {
	// Contents is the actual contents of a valid Kubeconfig
	Contents []byte `yaml:"contents"`
}

// UpgradeState stores information required to perform OSDe2e upgrade testing
type UpgradeState struct {
	// ReleaseName is the name of the release in a release stream.
	ReleaseName string `env:"UPGRADE_RELEASE_NAME" sect:"upgrade" yaml:"releaseName"`

	// Image is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.
	Image string `env:"UPGRADE_IMAGE" sect:"upgrade" yaml:"image"`

	// UpgradeVersionEqualToInstallVersion is true if the install version and upgrade versions are the same.
	UpgradeVersionEqualToInstallVersion bool
}
