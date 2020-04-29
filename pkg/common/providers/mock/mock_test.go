package mock

import "testing"

func TestClusterInteraction(t *testing.T) {
	mockProvider, _ := New("mockEnv")

	clusterID1, _ := mockProvider.LaunchCluster()
	clusterID2, _ := mockProvider.LaunchCluster()

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
