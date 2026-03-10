# Test Fixtures for Krkn-AI Analysis

This directory contains sample krkn-ai chaos test results used for testing the analysis feature.

## Files

### `all.csv`
Complete results from 3 generations of chaos testing with 10 scenarios:
- 9 successful scenarios with varying fitness scores
- 1 failed scenario (dns-outage with krkn_failure_score = -1.0)
- Covers multiple chaos types: node-cpu-hog, node-memory-hog, pod-scenarios, network-chaos, dns-outage, container-kill

### `health_check_report.csv`
Health check metrics for each scenario showing:
- Component response times (console, api, oauth)
- Success and failure counts
- Correlation with chaos scenarios

### `krkn-ai.yaml`
Krkn-ai configuration file showing:
- Genetic algorithm parameters (3 generations, population size 10)
- Fitness function configuration
- Enabled chaos scenario types and their parameters

## Usage

Tests copy these files to temporary test directories using `t.TempDir()`:

```go
func setupTestData(t *testing.T) string {
    t.Helper()

    reportDir := t.TempDir()
    reportsDir := filepath.Join(reportDir, "reports")
    require.NoError(t, os.MkdirAll(reportsDir, 0o755))

    // Copy testdata files
    copyTestFile(t, "all.csv", filepath.Join(reportsDir, "all.csv"))
    copyTestFile(t, "health_check_report.csv", filepath.Join(reportsDir, "health_check_report.csv"))
    copyTestFile(t, "krkn-ai.yaml", filepath.Join(reportDir, "krkn-ai.yaml"))

    return reportDir
}
```

This approach:
- ✅ Uses standard Go testing patterns (`testdata/` directory)
- ✅ Uses `t.TempDir()` for test output directories
- ✅ Keeps test data in version control for reproducibility
- ✅ Automatically cleans up temp directories after tests
