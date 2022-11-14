package mock

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/spi"
	"k8s.io/client-go/tools/clientcmd"
)

func TestClusterInteraction(t *testing.T) {
	mockProvider := makeMockProviderWithEnv("mockEnv")

	if hasQuota, err := mockProvider.CheckQuota(""); !hasQuota || err != nil {
		t.Errorf("expected quota or no error, got: %v, %v", hasQuota, err.Error())
	}

	clusterName1 := "cluster1"
	clusterName2 := "cluster2"

	if isValidClusterName1, err := mockProvider.IsValidClusterName(clusterName1); isValidClusterName1 != true || err != nil {
		t.Errorf("unexpected validation or error using cluster name: %s", clusterName1)
	}

	if isValidClusterName2, err := mockProvider.IsValidClusterName(clusterName2); isValidClusterName2 != true || err != nil {
		t.Errorf("unexpected validation or error using cluster name: %s", clusterName2)
	}

	clusterID1, _ := mockProvider.LaunchCluster(clusterName1)
	clusterID2, _ := mockProvider.LaunchCluster(clusterName2)

	cluster1, err := mockProvider.GetCluster(clusterID1)
	if err != nil {
		t.Errorf("error trying to get cluster 1: %v", err)
	}

	if cluster1.ID() != clusterID1 {
		t.Errorf("cluster IDs did not match for cluster 1. Expected %s, got %s", clusterID1, cluster1.ID())
	}

	cluster2, err := mockProvider.GetCluster(clusterID2)
	if err != nil {
		t.Errorf("error trying to get cluster 2: %v", err)
	}

	if cluster2.ID() != clusterID2 {
		t.Errorf("cluster IDs did not match for cluster 2. Expected %s, got %s", clusterID2, cluster2.ID())
	}

	mockProvider.DeleteCluster(clusterID1)

	_, err = mockProvider.GetCluster(clusterID1)

	if err == nil {
		t.Errorf("expected error when retrieving cluster 1 after deletion")
	}
}

func TestIntentionalFailures(t *testing.T) {
	// Setting the environment to fail will cause multiple common interactions to fail intentionally
	// Creation / retrieval / deletion of a cluster should all still work though. Some baseline
	// functionality should always work.
	mockProvider := makeMockProviderWithEnv("fail")

	// Quota Check
	quotaCheck, err := mockProvider.CheckQuota("")
	if quotaCheck {
		t.Error("expected quota to be false for fail environment")
	}
	if err == nil {
		t.Error("expected error to occur while checking quota")
	}

	failNames := []string{"error", "false"}
	for _, name := range failNames {
		if validClusterName, _ := mockProvider.IsValidClusterName(name); validClusterName != false {
			t.Errorf("expected an error to occur of validation to fail using cluster name: %s", name)
		}
	}

	clusterID1, _ := mockProvider.LaunchCluster("cluster1")

	// ClusterKubeconfig
	if kubeconfig, err := mockProvider.ClusterKubeconfig(clusterID1); kubeconfig != nil || err == nil {
		t.Errorf("expected error to occur retrieving clusterkubeconfig: %v, %v", kubeconfig, err)
	}

	// InstallAddons
	if addonsInstalled, err := mockProvider.InstallAddons(clusterID1, []string{"addon1", "addon2"}, map[string]map[string]string{}); addonsInstalled != 0 || err == nil {
		t.Errorf("expected error to occur installing addons: %v, %v", addonsInstalled, err)
	}

	// Versions
	if versions, err := mockProvider.Versions(); versions != nil || err == nil {
		t.Errorf("expected error to occur retrieving versions: %v, %v", versions, err)
	}

	// Logs
	if logs, err := mockProvider.Logs(clusterID1); logs != nil || err == nil {
		t.Errorf("expected error to occur retrieving logs: %v, %v", logs, err)
	}
}

func TestMockAddons(t *testing.T) {
	mockProvider := makeMockProviderWithEnv("mockEnv")

	clusterID1, _ := mockProvider.LaunchCluster("cluster1")

	toInstall := []string{"addon1", "addon2"}

	numInstalled, err := mockProvider.InstallAddons(clusterID1, toInstall, map[string]map[string]string{})
	if err != nil {
		t.Errorf("expected no error, got: %s", err.Error())
	}

	if numInstalled != len(toInstall) {
		t.Errorf("expected numInstalled to be 2, got %d", numInstalled)
	}

	cluster1, err := mockProvider.GetCluster(clusterID1)
	if err != nil {
		t.Errorf("error when retrieving cluster: %s", err.Error())
	}

	installedAddons := cluster1.Addons()
	if len(installedAddons) != len(toInstall) {
		t.Errorf("difference in addon list length: %d / %d", len(toInstall), len(installedAddons))
	}

	if !reflect.DeepEqual(toInstall, installedAddons) {
		t.Errorf("difference in addon array: %v, %v", toInstall, installedAddons)
	}
}

func TestClusterkubeconfig(t *testing.T) {
	mockProvider := makeMockProviderWithEnv("mockEnv")

	clusterID1, _ := mockProvider.LaunchCluster("cluster1")

	kubeconfig, err := mockProvider.ClusterKubeconfig(clusterID1)
	if err != nil {
		t.Errorf("expected no error, got %s", err.Error())
	}

	if _, err = clientcmd.NewClientConfigFromBytes(kubeconfig); err != nil {
		t.Errorf("invalid kubeconfig provided: %s", err.Error())
	}
}

func TestVersions(t *testing.T) {
	mockProvider := makeMockProviderWithEnv("mockEnv")

	versions, err := mockProvider.Versions()
	if err != nil {
		t.Errorf("error retrieving provider versions: %s", err.Error())
	}
	// We know our default test list should have 3
	if len(versions.AvailableVersions()) != 3 {
		t.Errorf("unexpected versionList length. Expected 3, got: %d", len(versions.AvailableVersions()))
	}

	// We know our default list version should be 4.5.6
	if versions.Default().String() != "4.5.6" {
		t.Errorf("unexpected default version. Expected 4.5.6, got: %s", versions.Default().String())
	}
}

func TestSetVersionList(t *testing.T) {
	mockProvider := makeMockProviderWithEnv("mockEnv")

	customVersions := []*spi.Version{
		spi.NewVersionBuilder().
			Version(semver.MustParse("4.3.10")).
			Default(false).
			Build(),
		spi.NewVersionBuilder().
			Version(semver.MustParse("4.3.13")).
			Default(true).
			Build(),
	}

	versionList := spi.NewVersionListBuilder().
		AvailableVersions(customVersions).
		DefaultVersionOverride(nil).
		Build()

	mockProvider.SetVersionList(versionList)

	versions, err := mockProvider.Versions()
	if err != nil {
		t.Errorf("error retrieving provider versions: %s", err.Error())
	}
	// Our custom list only has two entries
	if len(versions.AvailableVersions()) != 2 {
		t.Errorf("unexpected versionList length. Expected 2, got: %d", len(versions.AvailableVersions()))
	}

	// Our custom list default should be 4.3.13
	if versions.Default().String() != "4.3.13" {
		t.Errorf("unexpected default version. Expected 4.3.13, got: %s", versions.Default().String())
	}
}

func makeMockProviderWithEnv(env string) *MockProvider {
	viper.Reset()
	viper.Set(Env, env)
	// Setting the environment to fail will cause multiple common interactions to fail intentionally
	// Creation / retrieval / deletion of a cluster should all still work though. Some baseline
	// functionality should always work.
	mockProvider, _ := New()
	return mockProvider
}
