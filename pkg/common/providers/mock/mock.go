package mock

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/uuid"
	"github.com/markbates/pkger"
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
	versions *spi.VersionList
}

// New creates a new MockProvider.
func New(env string) (*MockProvider, error) {
	// Here we set a default
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

	versionList := spi.NewVersionListBuilder().
		AvailableVersions(versions).
		DefaultVersionOverride(nil).
		Build()

	return &MockProvider{
		env:      env,
		clusters: map[string]*spi.Cluster{},
		versions: versionList,
	}, nil
}

// LaunchCluster mocks a launch cluster operation.
func (m *MockProvider) LaunchCluster() (string, error) {
	clusterID := uuid.New().String()
	if m.env == "fail" {
		clusterID = m.env
	}

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
		return fmt.Errorf("fake error deleting cluster")
	}

	delete(m.clusters, clusterID)
	return nil
}

// GetCluster mocks a get cluster operation.
func (m *MockProvider) GetCluster(clusterID string) (*spi.Cluster, error) {
	if clusterID == "fail" {
		return nil, fmt.Errorf("failed to get versions: Some fake error")
	}

	if cluster, ok := m.clusters[clusterID]; ok {
		return cluster, nil
	}
	return nil, fmt.Errorf("couldn't find cluster in mock provider")
}

// ClusterKubeconfig mocks a cluster kubeconfig operation.
func (m *MockProvider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	var (
		fileReader http.File
		err        error
	)

	if clusterID == "fail" {
		return nil, fmt.Errorf("failed to get versions: Some fake error")
	}
	// This kubeconfig is valid and can be parsed, but attmping to use it will cause failures :)

	if fileReader, err = pkger.Open("/assets/providers/mock/kubeconfig"); err != nil {
		return nil, err
	}

	f, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return nil, err
	}
	return []byte(f), nil
}

// CheckQuota mocks a check quota operation.
func (m *MockProvider) CheckQuota() (bool, error) {
	if m.env == "fail" {
		return false, fmt.Errorf("failed to get versions: Some fake error")
	}

	// By default this will pass.
	// If you want a purposeful CheckQuota failure, you should set up a `fail` environment
	return true, nil
}

// InstallAddons mocks an install addons operation.
func (m *MockProvider) InstallAddons(clusterID string, addonIDs []string) (int, error) {
	if clusterID == "fail" {
		return 0, fmt.Errorf("failed to get versions: Some fake error")
	}

	cluster, err := m.GetCluster(clusterID)
	if err != nil {
		return 0, fmt.Errorf("Unable to retrieve cluster: %s", err.Error())
	}
	// We can't access the addons field directly so we have to rebuild the cluster object from scratch
	// This is fine as any real provider would call an external API to update or retrieve addons and
	// we lose no state doing this.
	m.clusters[clusterID] = spi.NewClusterBuilder().
		ID(clusterID).
		Name(cluster.Name()).
		Version(cluster.Version()).
		State(cluster.State()).
		CloudProvider(cluster.CloudProvider()).
		Region(cluster.Region()).
		ExpirationTimestamp(cluster.ExpirationTimestamp()).
		Flavour(cluster.Flavour()).
		Addons(addonIDs).
		Build()

	return len(addonIDs), nil
}

// Versions mocks a versions operation.
func (m *MockProvider) Versions() (*spi.VersionList, error) {
	if m.env == "fail" {
		return nil, fmt.Errorf("Fake error returning version list")
	}

	return m.versions, nil

}

// Logs mocks a logs operation.
func (m *MockProvider) Logs(clusterID string) (map[string][]byte, error) {
	if clusterID == "fail" {
		return nil, fmt.Errorf("failed to get versions: Some fake error")
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

// SetVersionList lets us provide novel versions allowing us to properly flex
// version selection using the Mock provider
func (m *MockProvider) SetVersionList(list *spi.VersionList) {
	m.versions = list
}

// Type returns the provisioner type: mock
func (m *MockProvider) Type() string {
	return "mock"
}
