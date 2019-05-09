// Package osde2e launches an OSD cluster, performs tests on it, destroys it, and reports results to TestGrid.
package osde2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
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

// RunE2ETests runs the osde2e test suite using the given cfg.
func RunE2ETests(t *testing.T, cfg *config.Config) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	// set defaults
	if cfg.ClusterVersion == "" {
		cfg.ClusterVersion = DefaultVersion
	}

	if cfg.Suffix == "" {
		cfg.Suffix = randomStr(5)
	}

	if cfg.ClusterName == "" {
		safeVersion := strings.Replace(cfg.ClusterVersion, ".", "-", -1)
		cfg.ClusterName = "ci-cluster-" + safeVersion + "-" + cfg.Suffix
	}

	if cfg.ReportDir == "" {
		if dir, err := ioutil.TempDir("", "osde2e"); err == nil {
			cfg.ReportDir = dir
		}
	}

	// setup reporter
	os.Mkdir(cfg.ReportDir, os.ModePerm)
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
			start := time.Now().UTC().Unix()
			if buildNum, err = tg.StartBuild(ctx, start); err != nil {
				log.Printf("Failed to start TestGrid build: %v", err)
			} else {
				log.Printf("Started TestGrid build '%d'", buildNum)
			}
		}
		defer reportToTestGrid(t, tg, buildNum, cfg.ReportDir)
	} else {
		log.Print("NO_TESTGRID is set, skipping submitting to TestGrid...")
	}

	log.Println("Running e2e tests...")
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "OSD e2e suite", []ginkgo.Reporter{reporter})
}

func reportToTestGrid(t *testing.T, tg *testgrid.TestGrid, buildNum int, reportDir string) {
	if tg != nil {
		end := time.Now().UTC().Unix()
		passed := !t.Failed()
		result := "FAILURE"
		if passed {
			result = "SUCCESS"
		}

		finished := metadata.Finished{
			Timestamp: &end,
			Passed:    &passed,
			Result:    result,
		}

		ctx := context.Background()
		if err := tg.FinishBuild(ctx, buildNum, finished, reportDir); err != nil {
			log.Printf("Failed to report results to TestGrid for build '%d': %v", buildNum, err)
		} else {
			log.Printf("Successfully reported results to TestGrid for build '%d'", buildNum)
		}
	} else {
		log.Print("Skipping reporting to TestGrid...")
	}
}
