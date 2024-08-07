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
	if labelFilter != "" {
		suiteConfig.LabelFilter = labelFilter
	}

	if suiteConfig.LabelFilter == "" {
		suiteConfig.LabelFilter = "CreateRosaCluster|| PostMigrationCheck || " +
			"RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster "
	}

	reporterConfig.JUnitReport = "junit.xml"

	RunSpecs(t, "SdnMigration Suite", suiteConfig, reporterConfig)
}
