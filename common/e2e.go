// Package common launches an OSD cluster, performs tests on it, and destroys it.
package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/osd"
	"github.com/openshift/osde2e/pkg/upgrade"
)

const (
	customMetadataFile string = "custom-prow-metadata.json"
)

// OSD is used to deploy and manage clusters.
var OSD *osd.OSD

// RunE2ETests runs the osde2e test suite using the given cfg.
func RunE2ETests(t *testing.T, cfg *config.Config) {
	var err error
	gomega.RegisterFailHandler(ginkgo.Fail)

	// set defaults
	if cfg.Suffix == "" {
		cfg.Suffix = randomStr(3)
	}

	if cfg.ReportDir == "" {
		if dir, err := ioutil.TempDir("", "osde2e"); err == nil {
			cfg.ReportDir = dir
		}
	}

	// setup OSD unless Kubeconfig is present
	if len(cfg.Kubeconfig) > 0 {
		log.Print("Found an existing Kubeconfig!")
	} else {
		if OSD, err = osd.New(cfg.OCMToken, cfg.OSDEnv, cfg.DebugOSD); err != nil {
			t.Fatalf("could not setup OSD: %v", err)
		}

		metadata.Instance.Environment = cfg.OSDEnv

		// check that enough quota exists for this test if creating cluster
		if len(cfg.ClusterID) == 0 {
			if enoughQuota, err := OSD.CheckQuota(cfg); err != nil {
				log.Printf("Failed to check if enough quota is available: %v", err)
			} else if !enoughQuota {
				t.Fatal("Currently not enough quota exists to run this test, failing...")
			}
		}

		// configure cluster and upgrade versions
		if err = ChooseVersions(cfg, OSD); err != nil {
			t.Fatalf("failed to configure versions: %v", err)
		}
	}

	// setup reporter
	if err = os.Mkdir(cfg.ReportDir, os.ModePerm); err != nil {
		log.Printf("Could not create reporter directory: %v", err)
	}
	reportPath := path.Join(cfg.ReportDir, fmt.Sprintf("junit_%v.xml", cfg.Suffix))
	reporter := reporters.NewJUnitReporter(reportPath)

	if !cfg.DryRun {
		log.Println("Running e2e tests...")
		ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "OSD e2e suite", []ginkgo.Reporter{reporter})
		// upgrade cluster if requested
		if cfg.UpgradeImage != "" || cfg.UpgradeReleaseStream != "" {
			if cfg.Kubeconfig != nil {
				if err = upgrade.RunUpgrade(cfg, OSD); err != nil {
					t.Errorf("Error performing upgrade: %s", err.Error())
				}

				log.Println("Running e2e tests POST-UPGRADE...")
				ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "OSD e2e suite post-upgrade", []ginkgo.Reporter{reporter})
			} else {
				log.Println("No Kubeconfig found from initial cluster setup. Unable to run upgrade.")
			}
		}

		if cfg.ReportDir != "" {
			if err = metadata.Instance.WriteToJSON(filepath.Join(cfg.ReportDir, customMetadataFile)); err != nil {
				t.Errorf("Error while writing metadata: %s", err.Error())
			}
		}

		if OSD != nil {
			if cfg.DestroyClusterAfterTest {
				log.Printf("Destroying cluster '%s'...", cfg.ClusterID)
				if err = OSD.DeleteCluster(cfg.ClusterID); err != nil {
					t.Errorf("Error deleting cluster: %s", err.Error())
				}
			} else {
				log.Printf("For debugging, please look for cluster ID %s in environment %s", cfg.ClusterID, cfg.OSDEnv)
			}
		}
	}
}
