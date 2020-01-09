package latest

import (
	"flag"
	"log"
	"testing"

	"github.com/openshift/osde2e/common"
	"github.com/openshift/osde2e/pkg/config"
	// import suites to be tested
)

func init() {
	var filename string
	testing.Init()

	cfg := config.Cfg

	flag.StringVar(&filename, "e2e-config", ".osde2e.yaml", "Config file for osde2e")
	flag.Parse()

	cfg.LoadFromYAML(filename)

}
func TestE2E(t *testing.T) {
	// force overwriting the attributes to make sure the cluster with the specified cluster will be created.
	cfg := config.Cfg
	cfg.Kubeconfig.Path = ""
	cfg.Cluster.ID = ""
	cfg.Cluster.Version = ""
	cfg.Cluster.DestroyAfterTest = true

	versionList, err := common.GetNewestAndOldestVersions(cfg)
	if err != nil {
		log.Printf("Failed to do clusterImageSets testing: %+v\n", err)
	} else {
		if len(versionList) == 0 {
			log.Printf("Skip doing clusterImageSets testing: Default version is covered by the regular testing\n")
		} else if len(versionList) == 1 {
			log.Printf("Skip doing clusterImageSets testing: The version %s is covered by the oldest version testing\n", versionList[0])
		} else {
			log.Printf("Start doing clusterImageSets testing with the specified version %+v\n", versionList[1])
			cfg.Cluster.Version = versionList[1]
			common.RunE2ETests(t, cfg)
		}
	}
}
