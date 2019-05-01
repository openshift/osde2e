package osde2e

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
)

func RunE2ETests(t *testing.T, cfg Config) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	// setup reporter
	os.Mkdir(Cfg.ReportDir, os.ModePerm)
	reportPath := path.Join(cfg.ReportDir, fmt.Sprintf("junit_%v.xml", Cfg.Prefix))
	reporter := reporters.NewJUnitReporter(reportPath)

	log.Println("Running e2e tests...")
	ginkgo.RunSpecsWithCustomReporters(t, "OSD e2e suite", []ginkgo.Reporter{reporter})
}
