package fedramp

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFedrampSuite(t *testing.T) {
	RegisterFailHandler(Fail)

	suiteConfig, reporterConfig := GinkgoConfiguration()
	reporterConfig.JUnitReport = os.Getenv("REPORT_DIR") + "/" + "junit.xml"
	suiteConfig.Timeout = 3 * time.Hour

	labelFilter := os.Getenv("GINKGO_LABEL_FILTER")
	if labelFilter != "" {
		suiteConfig.LabelFilter = labelFilter
	}

	if suiteConfig.LabelFilter == "" {
		suiteConfig.LabelFilter = "Fedramp"
	}

	RunSpecs(t, "Fedramp Suite", suiteConfig, reporterConfig)
}
