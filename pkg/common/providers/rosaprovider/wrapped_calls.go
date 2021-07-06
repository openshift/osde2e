package rosaprovider

import (
	"time"

	"github.com/openshift/osde2e/pkg/common/spi"
)

// The rest of the SPI functions will be wrapped by the OCM provider until the ROSA provider can be adequately refactored.

// DeleteCluster will call DeleteCluster from the OCM provider.
func (m *ROSAProvider) DeleteCluster(clusterID string) error {
	return m.ocmProvider.DeleteCluster(clusterID)
}

// ScaleCluster will call ScaleCluster from the OCM provider.
func (m *ROSAProvider) ScaleCluster(clusterID string, numComputeNodes int) error {
	return m.ocmProvider.ScaleCluster(clusterID, numComputeNodes)
}

// ListClusters will call ListClusters from the OCM provider.
func (m *ROSAProvider) ListClusters(query string) ([]*spi.Cluster, error) {
	return m.ocmProvider.ListClusters(query)
}

// GetCluster will call GetCluster from the OCM provider.
func (m *ROSAProvider) GetCluster(clusterID string) (*spi.Cluster, error) {
	return m.ocmProvider.GetCluster(clusterID)
}

// ClusterKubeconfig will call ClusterKubeconfig from the OCM provider.
func (m *ROSAProvider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	return m.ocmProvider.ClusterKubeconfig(clusterID)
}

// CheckQuota will call CheckQuota from the OCM provider.
func (m *ROSAProvider) CheckQuota(sku string) (bool, error) {
	return m.ocmProvider.CheckQuota(sku)
}

// InstallAddons will call InstallAddons from the OCM provider.
func (m *ROSAProvider) InstallAddons(clusterID string, addonIDs []spi.AddOnID, addonParams map[spi.AddOnID]spi.AddOnParams) (int, error) {
	return m.ocmProvider.InstallAddons(clusterID, addonIDs, addonParams)
}

// Logs will call Logs from the OCM provider.
func (m *ROSAProvider) Logs(clusterID string) (map[string][]byte, error) {
	return m.ocmProvider.Logs(clusterID)
}

// Environment will call Environment from the OCM provider.
func (m *ROSAProvider) Environment() string {
	return m.ocmProvider.Environment()
}

// Metrics will call Metrics from the OCM provider.
func (m *ROSAProvider) Metrics(clusterID string) (bool, error) {
	return m.ocmProvider.Metrics(clusterID)
}

// UpgradeSource will call UpgradeSource from the OCM provider.
func (m *ROSAProvider) UpgradeSource() spi.UpgradeSource {
	return m.ocmProvider.UpgradeSource()
}

// CincinnatiChannel will call CincinnatiChannel from the OCM provider.
func (m *ROSAProvider) CincinnatiChannel() spi.CincinnatiChannel {
	return m.ocmProvider.CincinnatiChannel()
}

// ExtendExpiry will call ExtendExpiry from the OCM provider.
func (m *ROSAProvider) ExtendExpiry(clusterID string, hours uint64, minutes uint64, seconds uint64) error {
	return m.ocmProvider.ExtendExpiry(clusterID, hours, minutes, seconds)
}

// Expire will call Expire from the OCM provider.
func (m *ROSAProvider) Expire(clusterID string) error {
	return m.ocmProvider.Expire(clusterID)
}

// AddProperty will call AddProperty from the OCM provider.
func (m *ROSAProvider) AddProperty(cluster *spi.Cluster, tag string, value string) error {
	return m.ocmProvider.AddProperty(cluster, tag, value)
}

// Upgrade initiates a cluster upgrade from the OCM provider.
func (m *ROSAProvider) Upgrade(clusterID string, version string, t time.Time) error {
	return m.ocmProvider.Upgrade(clusterID, version, t)
}

// GetUpgradePolicyID fetchs the upgrade policy from the OCM provider
func (m *ROSAProvider) GetUpgradePolicyID(clusterID string) (string, error) {
	return m.ocmProvider.GetUpgradePolicyID(clusterID)
}

// UpdateSchedule mocks reschedule the upgrade via the OCM provider
func (m *ROSAProvider) UpdateSchedule(clusterID string, version string, t time.Time, policyID string) error {
	return m.ocmProvider.UpdateSchedule(clusterID, version, t, policyID)
}

// DetermineMachineType calls DetermineMachineType from the OCM provider
func (m *ROSAProvider) DetermineMachineType(cloudProvider string) (string, error) {
	return m.ocmProvider.DetermineMachineType(cloudProvider)
}

// Resume calls DetermineMachineType from the OCM provider
func (m *ROSAProvider) Resume(id string) bool {
	return m.ocmProvider.Resume(id)
}

// Hibernate calls DetermineMachineType from the OCM provider
func (m *ROSAProvider) Hibernate(id string) bool {
	return m.ocmProvider.Hibernate(id)
}
