package mcscupgrade_test

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMcscupgrade(t *testing.T) {
	RegisterFailHandler(Fail)

	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.Timeout = 10 * time.Hour

	labelFilter := os.Getenv("GINKGO_LABEL_FILTER")
	if labelFilter != "" {
		suiteConfig.LabelFilter = labelFilter
	}

	if suiteConfig.LabelFilter == "" {
		suiteConfig.LabelFilter = "ApplyHCPWorkloads || RemoveHCPWorkloads || " +
			"MCUpgrade || SCUpgrade || MCUpgradeHealthChecks || SCUpgradeHealthChecks"
	}

	reporterConfig.JUnitReport = "junit.xml"

	RunSpecs(t, "Mcscupgrade Suite", suiteConfig, reporterConfig)
}
