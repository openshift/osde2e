package sdn_migration_test

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSdnMigration(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.Timeout = 10 * time.Hour

	labelFilter := os.Getenv("GINKGO_LABEL_FILTER")

	// Define the filter values in a map
	labelFilters := map[string]string{
		"DefaultBuild":          "CreateRosaCluster || PostMigrationCheck || RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster",
		"DefaultBuildWithProxy": "CreateRosaCluster || PostMigrationCheck || RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster || EnableClusterProxy",
	}

	if filter, exists := labelFilters[labelFilter]; exists {
		suiteConfig.LabelFilter = filter
	} else if suiteConfig.LabelFilter == "" {
		suiteConfig.LabelFilter = labelFilters["DefaultBuild"]
	} else {
		suiteConfig.LabelFilter = labelFilter
	}

	reporterConfig.JUnitReport = os.Getenv("REPORT_DIR") + "/" + "junit.xml"

	RunSpecs(t, "SdnMigration Suite", suiteConfig, reporterConfig)
}
