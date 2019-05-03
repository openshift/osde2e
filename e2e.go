package osde2e

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	"k8s.io/test-infra/testgrid/metadata"

	"github.com/openshift/osde2e/pkg/cluster"
	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/testgrid"
)

var (
	// Cfg is the configuration being used for end to end testing.
	Cfg = new(config.Config)

	// UHC is used to deploy and manage clusters.
	UHC *cluster.UHC

	// ClusterId is used to identify a cluster.
	ClusterId string

	// Kubeconfig is used to access a cluster.
	Kubeconfig []byte
)

func RunE2ETests(t *testing.T, cfg *config.Config) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	// setup reporter
	os.Mkdir(Cfg.ReportDir, os.ModePerm)
	reportPath := path.Join(cfg.ReportDir, fmt.Sprintf("junit_%v.xml", Cfg.Suffix))
	reporter := reporters.NewJUnitReporter(reportPath)

	// setup testgrid
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

	log.Println("Running e2e tests...")
	ginkgo.RunSpecsWithCustomReporters(t, "OSD e2e suite", []ginkgo.Reporter{reporter})
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
			log.Printf("Sucessfully reported results to TestGrid for build '%d'", buildNum)
		}
	} else {
		log.Print("Skipping reporting to TestGrid...")
	}
}
