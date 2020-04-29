package mock

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/state"
	"github.com/openshift/osde2e/pkg/common/util"
)

const (
	// MockCloudProvider indicates that the cloud provider is just a mock.
	MockCloudProvider = "mock"

	// MockRegion indicates that the region used is just a mock.
	MockRegion = "mock-region"
)

// MockProvider for unit testing.
type MockProvider struct {
	env      string
	clusters map[string]*spi.Cluster
}

// New creates a new MockProvider.
func New(env string) (*MockProvider, error) {
	return &MockProvider{
		env:      env,
		clusters: map[string]*spi.Cluster{},
	}, nil
}

// LaunchCluster mocks a launch cluster operation.
func (m *MockProvider) LaunchCluster() (string, error) {
	clusterID := uuid.New().String()

	m.clusters[clusterID] = spi.NewClusterBuilder().
		ID(clusterID).
		Name(util.RandomStr(5)).
		Version(state.Instance.Cluster.Version).
		State(spi.ClusterStateReady).
		CloudProvider(MockCloudProvider).
		Region(MockRegion).
		ExpirationTimestamp(time.Now()).
		Flavour("osd-4").
		Build()

	return clusterID, nil
}

// DeleteCluster mocks a delete cluster operation.
func (m *MockProvider) DeleteCluster(clusterID string) error {
	delete(m.clusters, clusterID)
	return nil
}

// GetCluster mocks a get cluster operation.
func (m *MockProvider) GetCluster(clusterID string) (*spi.Cluster, error) {
	if cluster, ok := m.clusters[clusterID]; ok {
		return cluster, nil
	}
	return nil, fmt.Errorf("couldn't find cluster in mock provider")
}

// ClusterKubeconfig mocks a cluster kubeconfig operation.
func (m *MockProvider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	return nil, fmt.Errorf("cluster kubeconfig is currently unsupported by the mock provider")
}

// CheckQuota mocks a check quota operation.
func (m *MockProvider) CheckQuota() (bool, error) {
	return false, fmt.Errorf("check quota is currently unsupported by the mock provider")
}

// InstallAddons mocks an install addons operation.
func (m *MockProvider) InstallAddons(clusterID string, addonIDs []string) (int, error) {
	return 0, fmt.Errorf("install addons is currently unsupported by the mock provider")
}

// Versions mocks a versions operation.
func (m *MockProvider) Versions() (*spi.VersionList, error) {
	return nil, fmt.Errorf("versions is currently unsupported by the mock provider")
}

// Logs mocks a logs operation.
func (m *MockProvider) Logs(clusterID string) (map[string][]byte, error) {
	return map[string][]byte{}, fmt.Errorf("versions is currently unsupported by the mock provider")
}

// Environment mocks an environment operation.
func (m *MockProvider) Environment() string {
	return m.env
}

// UpgradeSource mocks an environment source operation.
func (m *MockProvider) UpgradeSource() spi.UpgradeSource {
	return spi.CincinnatiSource
}

// CincinnatiChannel mocks a cincinnati channel operation.
func (m *MockProvider) CincinnatiChannel() spi.CincinnatiChannel {
	return spi.CincinnatiStableChannel
}
