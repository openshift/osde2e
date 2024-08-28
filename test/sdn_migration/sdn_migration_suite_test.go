package sdn_migration_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSdnMigration(t *testing.T) {
	RegisterFailHandler(Fail)
	// Get the current date and time
	now := time.Now()

	// Format the date and time
	dateTime := now.Format("20060102_150405")

	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.Timeout = 10 * time.Hour

	labelFilter := os.Getenv("GINKGO_LABEL_FILTER")

	// Define the filter values in a map
	labelFilters := map[string]string{
		"DefaultBuild":            "CreateRosaCluster || PostMigrationCheck || RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster",
		"DefaultBuildWithProxy":   "CreateRosaCluster || PostMigrationCheck || RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster || EnableClusterProxy",
		"AutoScaleBuild":          "CreateRosaCluster || PostMigrationCheck || RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster || EnableAutoScaling",
		"AutoScaleBuildWithProxy": "CreateRosaCluster || PostMigrationCheck || RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster || EnableAutoScaling || EnableClusterProxy",
	}

	if filter, exists := labelFilters[labelFilter]; exists {
		suiteConfig.LabelFilter = filter
	} else if suiteConfig.LabelFilter == "" {
		suiteConfig.LabelFilter = labelFilters["DefaultBuild"]
	} else {
		suiteConfig.LabelFilter = labelFilter
	}

	prefix := "sdn_migration-"

	// Construct the filename
	filename := fmt.Sprintf("%s%s.xml", prefix, dateTime)
	reporterConfig.JUnitReport = filename //"junit.xml"

	RunSpecs(t, "SdnMigration Suite", suiteConfig, reporterConfig)
}
