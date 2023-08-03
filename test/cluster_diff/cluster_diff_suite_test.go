package cluster_diff_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClusterDiff(t *testing.T) {
	RegisterFailHandler(Fail)

	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.Timeout = 3 * time.Hour

	RunSpecs(t, "Cluster Diff Suite", suiteConfig)
}
