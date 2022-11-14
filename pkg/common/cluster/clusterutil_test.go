package cluster

import (
	"testing"

	"github.com/openshift/osde2e/pkg/common/config"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

func TestRandomClusterName(t *testing.T) {
	var name string
	viper.Set(config.Cluster.Name, "random")
	for i := 0; i < 100; i++ {
		name = clusterName()
		if len(name) > 15 {
			t.Errorf("%s greater than 15 characters", name)
		}
	}
}
