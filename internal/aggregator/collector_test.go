package aggregator

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtifactCollector_collectFromReportDir(t *testing.T) {
	// Create a temporary test directory structure
	tempDir := t.TempDir()

	// Create test report directory structure
	reportDir := filepath.Join(tempDir, "report")
	installDir := filepath.Join(reportDir, "install")
	upgradeDir := filepath.Join(reportDir, "upgrade")
	mustGatherDir := filepath.Join(reportDir, "must-gather")
	clusterLogsDir := filepath.Join(reportDir, "cluster-logs")

	require.NoError(t, os.MkdirAll(installDir, 0o755))
	require.NoError(t, os.MkdirAll(upgradeDir, 0o755))
	require.NoError(t, os.MkdirAll(mustGatherDir, 0o755))
	require.NoError(t, os.MkdirAll(clusterLogsDir, 0o755))

	// Create test files
	createTestFiles(t, reportDir, installDir, upgradeDir, mustGatherDir, clusterLogsDir)

	// Create collector
	logger := logr.Discard()
	collector := newArtifactCollector(logger)

	// Test collection
	data, err := collector.collectFromReportDir(reportDir)

	require.NoError(t, err)
	require.NotNil(t, data)

	// Verify test results
	assert.Equal(t, 3, data.TestResults.TotalTests)
	assert.Equal(t, 1, data.TestResults.PassedTests)
	assert.Equal(t, 2, data.TestResults.FailedTests)
	assert.Equal(t, 2, len(data.FailedTests))

	// Verify failed test details
	failedTest := data.FailedTests[0]
	assert.Equal(t, "TestOperatorInstallation", failedTest.Name)
	assert.Equal(t, "operators.test", failedTest.ClassName)
	assert.Contains(t, failedTest.ErrorMsg, "operator installation timed out")

	// Verify logs were collected
	assert.Greater(t, len(data.Logs), 0)

	// Verify collection time
	assert.WithinDuration(t, time.Now(), data.CollectionTime, time.Minute)
}

func TestArtifactCollector_nonExistentDirectory(t *testing.T) {
	logger := logr.Discard()
	collector := newArtifactCollector(logger)

	_, err := collector.collectFromReportDir("/non/existent/path")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "report directory does not exist")
}

func TestArtifactCollector_emptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	logger := logr.Discard()
	collector := newArtifactCollector(logger)

	data, err := collector.collectFromReportDir(tempDir)

	require.NoError(t, err)
	require.NotNil(t, data)

	// Should have empty results but valid structure
	assert.Equal(t, 0, data.TestResults.TotalTests)
	assert.Equal(t, 0, len(data.FailedTests))
	assert.Equal(t, 0, len(data.Logs))
}

// Helper function to create test files
func createTestFiles(t *testing.T, reportDir, installDir, upgradeDir, mustGatherDir, clusterLogsDir string) {
	// Create build log
	buildLog := `2024-01-15 10:00:00 Starting osde2e test run
2024-01-15 10:01:00 Provisioning cluster
2024-01-15 10:05:00 Cluster provisioned successfully
2024-01-15 10:06:00 Running install tests
2024-01-15 10:10:00 Test failure detected in operator installation
2024-01-15 10:15:00 Running must-gather
2024-01-15 10:20:00 Test run completed with failures`

	require.NoError(t, os.WriteFile(filepath.Join(reportDir, "build-log.txt"), []byte(buildLog), 0o644))

	// Create JUnit XML files
	junitInstall := `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="install-tests" tests="2" failures="1" errors="0" time="120.5">
  <testcase name="TestClusterProvisioning" classname="cluster.test" time="60.0">
  </testcase>
  <testcase name="TestOperatorInstallation" classname="operators.test" time="60.5">
    <failure type="InstallationError" message="operator installation timed out after 300 seconds">
      <![CDATA[
Error: operator installation timed out after 300 seconds
Stack trace:
  at operator_test.go:45
  at test_runner.go:123
      ]]>
    </failure>
  </testcase>
</testsuite>`

	junitUpgrade := `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="upgrade-tests" tests="1" failures="1" errors="0" time="180.0">
  <testcase name="TestUpgradeProcess" classname="upgrade.test" time="180.0">
    <failure type="UpgradeError" message="cluster upgrade failed">
      <![CDATA[
Error: upgrade process failed during node update
Stack trace:
  at upgrade_test.go:67
  at test_runner.go:156
      ]]>
    </failure>
  </testcase>
</testsuite>`

	require.NoError(t, os.WriteFile(filepath.Join(installDir, "junit_install.xml"), []byte(junitInstall), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(upgradeDir, "junit_upgrade.xml"), []byte(junitUpgrade), 0o644))

	// Create must-gather files
	mustGatherLog := `2024-01-15T10:15:00Z Collecting cluster information
2024-01-15T10:15:30Z Gathering node information
2024-01-15T10:16:00Z Collecting pod logs from all namespaces
2024-01-15T10:17:00Z Must-gather collection completed`

	require.NoError(t, os.WriteFile(filepath.Join(mustGatherDir, "gather.log"), []byte(mustGatherLog), 0o644))

	// Create cluster logs
	clusterLog := `2024-01-15T10:00:00Z cluster-version-operator: Starting cluster version reconciliation
2024-01-15T10:01:00Z cluster-version-operator: Error updating cluster operators
2024-01-15T10:02:00Z cluster-version-operator: Retrying operator update`

	require.NoError(t, os.WriteFile(filepath.Join(clusterLogsDir, "cluster-version-operator.log"), []byte(clusterLog), 0o644))
}
