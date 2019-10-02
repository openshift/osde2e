// Package osde2e launches an OSD cluster, performs tests on it, destroys it, and reports results to TestGrid.
package osde2e

import (
	"context"
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
	"k8s.io/test-infra/testgrid/metadata"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/osd"
	"github.com/openshift/osde2e/pkg/testgrid"
)

// OSD is used to deploy and manage clusters.
var OSD *osd.OSD

const (
	// metadata key holding build-version
	buildVersionKey = "build-version"
)

// RunE2ETests runs the osde2e test suite using the given cfg.
func RunE2ETests(t *testing.T, cfg *config.Config) {
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
		cfg.ClusterUpTimeout = 135 * time.Minute
	}

	// setup OSD client
	var err error
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

	// setup reporter
	err = os.Mkdir(cfg.ReportDir, os.ModePerm)
	if err != nil {
		log.Printf("Could not create reporter directory: %v", err)
	}
	reportPath := path.Join(cfg.ReportDir, fmt.Sprintf("junit_%v.xml", cfg.Suffix))
	reporter := reporters.NewJUnitReporter(reportPath)

	// setup testgrid
	if !cfg.NoTestGrid {
		var buildNum int
		ctx := context.Background()
		tg, err := testgrid.NewTestGrid(cfg.TestGridBucket, cfg.TestGridPrefix, cfg.TestGridServiceAccount)
		if err != nil {
			log.Printf("Failed to setup TestGrid support: %v", err)
		} else {
			// check if new run should be performed
			if !doBuild(ctx, cfg, tg) {
				t.SkipNow()
			}

			now := time.Now().UTC().Unix()
			started := metadata.Started{
				Timestamp: now,
			}
			if buildNum, err = tg.StartBuild(ctx, &started); err != nil {
				log.Printf("Failed to start TestGrid build: %v", err)
			} else {
				log.Printf("Started TestGrid build '%d'", buildNum)
			}
		}
		defer reportToTestGrid(t, cfg, tg, buildNum)
	} else {
		log.Print("NO_TESTGRID is set, skipping submitting to TestGrid...")
	}

	if !cfg.DryRun {
		log.Println("Running e2e tests...")
		ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "OSD e2e suite", []ginkgo.Reporter{reporter})
	}
}

func reportToTestGrid(t *testing.T, cfg *config.Config, tg *testgrid.TestGrid, buildNum int) {
	if tg != nil {
		end := time.Now().UTC().Unix()
		passed := !t.Failed()
		result := "FAILURE"
		if passed {
			result = "SUCCESS"
		}

		// create metadata from config and set build version
		meta := cfg.TestGrid()
		meta[buildVersionKey] = buildVersion(cfg)

		finished := metadata.Finished{
			Timestamp: &end,
			Passed:    &passed,
			Result:    result,
			Metadata:  meta,
		}

		ctx := context.Background()
		if err := tg.FinishBuild(ctx, buildNum, &finished, cfg.ReportDir); err != nil {
			log.Printf("Failed to report results to TestGrid for build '%d': %v", buildNum, err)
		} else {
			log.Printf("Successfully reported results to TestGrid for build '%d'", buildNum)
		}
	} else {
		log.Print("Skipping reporting to TestGrid...")
	}
}

// doBuild checks if this run should be performed.
func doBuild(ctx context.Context, cfg *config.Config, tg *testgrid.TestGrid) bool {
	if cfg.CleanRuns > 0 {
		if finished, buildNum, err := tg.LatestFinished(ctx); err == nil && finished.Metadata != nil {
			// record build-version of current suite
			curVersion := buildVersion(cfg)

			// check if enough clean runs have been performed
			for i := 0; i < cfg.CleanRuns; i++ {
				if i != 0 {
					if finished, err = tg.Finished(ctx, buildNum-i); err != nil || finished.Metadata == nil {
						log.Printf("Could not get finished for build '%d', running build", buildNum)
						return true
					}
				}

				if finished.Passed == nil || !*finished.Passed {
					return true
				} else if bVersion, ok := finished.Metadata.String(buildVersionKey); ok {
					if *bVersion != curVersion {
						log.Printf("CLEAN_RUNS set, need %d more clean runs before skipping.", cfg.CleanRuns-i)
						return true
					}
				}
			}
			log.Printf("Skipping, CLEAN_RUNS set and build-version '%s' same for %d builds", curVersion, cfg.CleanRuns)
			return false
		} else if err != nil {
			log.Println("Error getting latest finished, running tests")
		}
	}
	return true
}
