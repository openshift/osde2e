package mock

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver"
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
	if clusterID == "fail" {
		return fmt.Errorf("Fake error deleting cluster.")
	}

	delete(m.clusters, clusterID)
	return nil
}

// GetCluster mocks a get cluster operation.
func (m *MockProvider) GetCluster(clusterID string) (*spi.Cluster, error) {
	if clusterID == "fail" {
		return nil, fmt.Errorf("Failed to get versions: Some fake error.")
	}

	if cluster, ok := m.clusters[clusterID]; ok {
		return cluster, nil
	}
	return nil, fmt.Errorf("couldn't find cluster in mock provider")
}

// ClusterKubeconfig mocks a cluster kubeconfig operation.
func (m *MockProvider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	if clusterID == "fail" {
		return nil, fmt.Errorf("Failed to get versions: Some fake error.")
	}

	return nil, fmt.Errorf("cluster kubeconfig is currently unsupported by the mock provider")
}

// CheckQuota mocks a check quota operation.
func (m *MockProvider) CheckQuota() (bool, error) {
	if m.env == "fail" {
		return false, fmt.Errorf("Failed to get versions: Some fake error.")
	}

	return false, fmt.Errorf("check quota is currently unsupported by the mock provider")
}

// InstallAddons mocks an install addons operation.
func (m *MockProvider) InstallAddons(clusterID string, addonIDs []string) (int, error) {
	if clusterID == "fail" {
		return 0, fmt.Errorf("Failed to get versions: Some fake error.")
	}

	return 0, fmt.Errorf("install addons is currently unsupported by the mock provider")
}

// Versions mocks a versions operation.
func (m *MockProvider) Versions() (*spi.VersionList, error) {
	if m.env == "fail" {
		return nil, fmt.Errorf("Fake error returning version list")
	}
	versions := []*spi.Version{
		spi.NewVersionBuilder().
			Version(semver.MustParse("1.2.3")).
			Default(false).
			Build(),
		spi.NewVersionBuilder().
			Version(semver.MustParse("2.3.4")).
			Default(false).
			Build(),
		spi.NewVersionBuilder().
			Version(semver.MustParse("4.5.6")).
			Default(true).
			Build(),
	}
	return spi.NewVersionListBuilder().
		AvailableVersions(versions).
		DefaultVersionOverride(nil).
		Build(), nil
}

// Logs mocks a logs operation.
func (m *MockProvider) Logs(clusterID string) (map[string][]byte, error) {
	if clusterID == "fail" {
		return nil, fmt.Errorf("Failed to get versions: Some fake error.")
	}

	logs := make(map[string][]byte)
	logs["logs.txt"] = []byte("Here is some lovely log content.")
	logs["build.log"] = []byte("Additional logs with a different name.")

	return logs, nil
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
