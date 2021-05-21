package spi

import (
	"reflect"
	"testing"
	"time"
)

func TestClusterBuilder(t *testing.T) {
	expirationTimestamp := time.Now()
	creationTimestamp := time.Now().Add(-6 * time.Hour)
	builtCluster := NewClusterBuilder().
		ID("test-id").
		Name("test-name").
		Version("test-version").
		CloudProvider("test-cloud-provider").
		Region("test-region").
		State(ClusterStateReady).
		ExpirationTimestamp(expirationTimestamp).
		CreationTimestamp(creationTimestamp).
		Flavour("test-flavour").
		Addons([]string{"test-addon1", "test-addon2"}).
		Build()

	definedCluster := Cluster{
		id:                  "test-id",
		name:                "test-name",
		version:             "test-version",
		cloudProvider:       "test-cloud-provider",
		region:              "test-region",
		state:               ClusterStateReady,
		expirationTimestamp: expirationTimestamp,
		creationTimestamp:   creationTimestamp,
		flavour:             "test-flavour",
		addons:              []string{"test-addon1", "test-addon2"},
	}

	if !reflect.DeepEqual(definedCluster, *builtCluster) {
		t.Errorf("cluster made through builder and cluster defined normally are not equal")
	}
}
