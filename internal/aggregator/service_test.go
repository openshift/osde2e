package aggregator

import (
	"context"
	"os"
	"path/filepath"
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

	// Verify metadata was set
	assert.Equal(t, "test-cluster", data.Metadata.ClusterID)
	assert.Equal(t, "rosa", data.Metadata.Provider)

	// Verify build logs were collected
	assert.Greater(t, len(data.BuildLogs), 0)
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
