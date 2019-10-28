// Package osde2e launches an OSD cluster, performs tests on it, and destroys it.
package osde2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/osd"
)

// OSD is used to deploy and manage clusters.
var OSD *osd.OSD

const (
	// metadata key holding build-version
	buildVersionKey = "build-version"
)

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

	// ensure to wait longer than infra alerting rules thresholds
	// otherwise startup failures won't trigger alerts
	if cfg.ClusterUpTimeout == 0 {
		cfg.ClusterUpTimeout = time.Duration(135) * time.Minute
	}

	// setup OSD unless Kubeconfig is present
	if len(cfg.Kubeconfig) > 0 {
		log.Print("Found an existing Kubeconfig!")
	} else {
		if OSD, err = osd.New(cfg.UHCToken, cfg.OSDEnv, cfg.DebugOSD); err != nil {
			t.Fatalf("could not setup OSD: %v", err)
		}

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
	err = os.Mkdir(cfg.ReportDir, os.ModePerm)
	if err != nil {
		log.Printf("Could not create reporter directory: %v", err)
	}
	reportPath := path.Join(cfg.ReportDir, fmt.Sprintf("junit_%v.xml", cfg.Suffix))
	reporter := reporters.NewJUnitReporter(reportPath)

	if !cfg.DryRun {
		log.Println("Running e2e tests...")
		ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "OSD e2e suite", []ginkgo.Reporter{reporter})
	}
}
