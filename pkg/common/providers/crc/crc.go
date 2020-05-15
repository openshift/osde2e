package crc

import (
	"github.com/openshift/osde2e/pkg/common/spi"
)

const (
	// CloudProvider indicates that the cloud provider is just a CRC.
	CloudProvider = "crc"

	// Region indicates that the region used is just a CRC.
	Region = "local"
)

// Provider for unit testing.
type Provider struct {
	env      string
	clusters map[string]*spi.Cluster
}

// New creates a new Provider.
func New(env string) (*Provider, error) {
	return &Provider{
		env:      env,
		clusters: map[string]*spi.Cluster{},
	}, nil
}

// LaunchCluster CRCs a launch cluster operation.
func (m *Provider) LaunchCluster() (string, error) {
	return "", nil
}

// DeleteCluster CRCs a delete cluster operation.
func (m *Provider) DeleteCluster(clusterID string) error {
	return nil
}

// GetCluster CRCs a get cluster operation.
func (m *Provider) GetCluster(clusterID string) (*spi.Cluster, error) {
	return nil, nil
}

// ClusterKubeconfig CRCs a cluster kubeconfig operation.
func (m *Provider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	return nil, nil
}

// CheckQuota CRCs a check quota operation.
func (m *Provider) CheckQuota() (bool, error) {
	return true, nil
}

// InstallAddons CRCs an install addons operation.
func (m *Provider) InstallAddons(clusterID string, addonIDs []string) (int, error) {
	return 0, nil
}

// Versions CRCs a versions operation.
func (m *Provider) Versions() (*spi.VersionList, error) {
	return nil, nil
}

// Logs CRCs a logs operation.
func (m *Provider) Logs(clusterID string) (map[string][]byte, error) {
	return nil, nil
}

// Environment CRCs an environment operation.
func (m *Provider) Environment() string {
	return m.env
}

// UpgradeSource CRCs an environment source operation.
func (m *Provider) UpgradeSource() spi.UpgradeSource {
	return spi.CincinnatiSource
}

// CincinnatiChannel CRCs a cincinnati channel operation.
func (m *Provider) CincinnatiChannel() spi.CincinnatiChannel {
	return spi.CincinnatiStableChannel
}
