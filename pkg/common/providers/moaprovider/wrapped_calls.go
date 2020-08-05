package moaprovider

import (
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// The rest of the SPI functions will be wrapped by the OCM provider until the MOA provider can be adequately refactored.

// DeleteCluster will call DeleteCluster from the OCM provider.
func (m *MOAProvider) DeleteCluster(clusterID string) error {
	return m.ocmProvider.DeleteCluster(clusterID)
}

// ScaleCluster will call ScaleCluster from the OCM provider.
func (m *MOAProvider) ScaleCluster(clusterID string, numComputeNodes int) error {
	return m.ocmProvider.ScaleCluster(clusterID, numComputeNodes)
}

// ListClusters will call ListClusters from the OCM provider.
func (m *MOAProvider) ListClusters(query string) ([]*spi.Cluster, error) {
	return m.ocmProvider.ListClusters(query)
}

// GetCluster will call GetCluster from the OCM provider.
func (m *MOAProvider) GetCluster(clusterID string) (*spi.Cluster, error) {
	return m.ocmProvider.GetCluster(clusterID)
}

// ClusterKubeconfig will call ClusterKubeconfig from the OCM provider.
func (m *MOAProvider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	return m.ocmProvider.ClusterKubeconfig(clusterID)
}

// CheckQuota will call CheckQuota from the OCM provider.
func (m *MOAProvider) CheckQuota() (bool, error) {
	return m.ocmProvider.CheckQuota()
}

// InstallAddons will call InstallAddons from the OCM provider.
func (m *MOAProvider) InstallAddons(clusterID string, addonIDs []string) (int, error) {
	return m.ocmProvider.InstallAddons(clusterID, addonIDs)
}

// Versions will call Versions from the OCM provider.
func (m *MOAProvider) Versions() (*spi.VersionList, error) {
	return m.ocmProvider.Versions()
}

// Logs will call Logs from the OCM provider.
func (m *MOAProvider) Logs(clusterID string) (map[string][]byte, error) {
	return m.ocmProvider.Logs(clusterID)
}

// Environment will call Environment from the OCM provider.
func (m *MOAProvider) Environment() string {
	return m.ocmProvider.Environment()
}

// Metrics will call Metrics from the OCM provider.
func (m *MOAProvider) Metrics(clusterID string) (*v1.ClusterMetrics, error) {
	return m.ocmProvider.Metrics(clusterID)
}

// UpgradeSource will call UpgradeSource from the OCM provider.
func (m *MOAProvider) UpgradeSource() spi.UpgradeSource {
	return m.ocmProvider.UpgradeSource()
}

// CincinnatiChannel will call CincinnatiChannel from the OCM provider.
func (m *MOAProvider) CincinnatiChannel() spi.CincinnatiChannel {
	return m.ocmProvider.CincinnatiChannel()
}

// ExtendExpiry will call ExtendExpiry from the OCM provider.
func (m *MOAProvider) ExtendExpiry(clusterID string, hours uint64, minutes uint64, seconds uint64) error {
	return m.ocmProvider.ExtendExpiry(clusterID, hours, minutes, seconds)
}

// AddProperty will call AddProperty from the OCM provider.
func (m *MOAProvider) AddProperty(clusterID string, tag string, value string) error {
	return m.ocmProvider.AddProperty(clusterID, tag, value)
}
