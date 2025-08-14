package aggregator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Collect(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	reportDir := filepath.Join(tempDir, "report")
	require.NoError(t, os.MkdirAll(reportDir, 0o755))

	// Create minimal test files
	buildLog := "Test build log content"
	require.NoError(t, os.WriteFile(filepath.Join(reportDir, "build-log.txt"), []byte(buildLog), 0o644))

	// Create service
	logger := logr.Discard()
	service := NewService(logger)

	// // Create test metadata
	// metadata := ClusterMetadata{
	// 	ClusterID: "test-cluster",
	// 	Provider:  "rosa",
	// }

	// Test collection
	ctx := context.Background()
	data, err := service.Collect(ctx, reportDir)

	require.NoError(t, err)
	require.NotNil(t, data)

	// Verify logs were collected (build-log.txt should be in there)
	assert.Greater(t, len(data.Logs), 0)

	// Check that our test file was collected
	found := false
	for _, log := range data.Logs {
		if strings.Contains(log.Source, "build-log.txt") {
			found = true
			break
		}
	}
	assert.True(t, found, "build-log.txt should be in collected logs")
}

func TestService_NonExistentDirectory(t *testing.T) {
	logger := logr.Discard()
	service := NewService(logger)

	// metadata := ClusterMetadata{
	// 	ClusterID: "test-cluster",
	// }

	ctx := context.Background()
	_, err := service.Collect(ctx, "/non/existent/path")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "report directory does not exist")
}
