package oldest

import (
	"log"
	"testing"

	"github.com/openshift/osde2e/common"
	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/state"
	// import suites to be tested
)

func TestE2E(t *testing.T) {
	// force overwriting the attributes to make sure the cluster with the specified version will be created.
	cfg := config.Instance
	state := state.Instance

	cfg.Kubeconfig.Path = ""
	cfg.Cluster.DestroyAfterTest = true
	state.Cluster.ID = ""
	state.Cluster.Version = ""

	versionList, err := common.GetEnabledNoDefaultVersions()
	if err != nil {
		log.Printf("Failed to do clusterImageSets testing: %+v\n", err)
	} else {
		if len(versionList) == 0 {
			log.Printf("Skip doing clusterImageSets testing: Default version is covered by the regular testing\n")
		} else {
			log.Printf("Start doing clusterImageSets testing with the specified version %+v\n", versionList[0])
			state.Cluster.Version = versionList[0]
			common.RunE2ETests(t)
		}
	}
}
