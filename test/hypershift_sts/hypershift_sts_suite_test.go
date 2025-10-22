package hypershift_sts

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHyperShiftSTSSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	
	suiteConfig, reporterConfig := GinkgoConfiguration()
	reporterConfig.JUnitReport = os.Getenv("REPORT_DIR") + "/" + "junit.xml"
	suiteConfig.Timeout = 30 * time.Minute
	
	labelFilter := os.Getenv("GINKGO_LABEL_FILTER")
	if labelFilter != "" {
		suiteConfig.LabelFilter = labelFilter
	}
	if suiteConfig.LabelFilter == "" {
		suiteConfig.LabelFilter = "HyperShiftSTS"
	}
	
	RunSpecs(t, "HyperShift STS Suite", suiteConfig, reporterConfig)
}
