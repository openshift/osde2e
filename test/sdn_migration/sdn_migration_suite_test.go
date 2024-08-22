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

	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.Timeout = 10 * time.Hour

	labelFilter := os.Getenv("GINKGO_LABEL_FILTER")
	if labelFilter != "" {
		suiteConfig.LabelFilter = labelFilter
	}

	if suiteConfig.LabelFilter == "" {
		suiteConfig.LabelFilter = "CreateRosaCluster || PostMigrationCheck || " +
			"RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster"
	}

	prefix := "sdn_migration-"

	// Get the current date and time
	now := time.Now()

	// Format the date and time
	dateTime := now.Format("20060102_150405")

	// Construct the filename
	filename := fmt.Sprintf("%s%s.xml", prefix, dateTime)
	reporterConfig.JUnitReport = filename //"junit.xml"

	RunSpecs(t, "SdnMigration Suite", suiteConfig, reporterConfig)
}
