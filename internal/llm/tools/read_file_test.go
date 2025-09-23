package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift/osde2e/internal/aggregator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFileTool_Name(t *testing.T) {
	tool := &readFileTool{}
	assert.Equal(t, "read_file", tool.Name())
}

func TestReadFileTool_Description(t *testing.T) {
	tool := &readFileTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "Reads a specific file")
	assert.Contains(t, desc, "line range")
}

func TestReadFileTool_Schema(t *testing.T) {
	tool := &readFileTool{}
	schema := tool.Schema()

	require.NotNil(t, schema)
	assert.Equal(t, "OBJECT", string(schema.Type))

	// Check required parameters
	assert.Contains(t, schema.Required, "path")
	assert.Len(t, schema.Required, 1)

	// Check properties
	assert.Contains(t, schema.Properties, "path")
	assert.Contains(t, schema.Properties, "start")
	assert.Contains(t, schema.Properties, "stop")
}

func TestReadFileTool_Execute(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	testContent := `line 1
line 2
line 3
line 4
line 5`

	err := os.WriteFile(testFile, []byte(testContent), 0o644)
	require.NoError(t, err)

	// Create aggregated data with the test file
	data := &aggregator.AggregatedData{
		LogArtifacts: []aggregator.LogEntry{
			{Source: testFile},
		},
	}

	tool := &readFileTool{}
	ctx := context.Background()

	t.Run("read entire file", func(t *testing.T) {
		params := map[string]any{
			"path": testFile,
		}

		result, err := tool.Execute(ctx, params, data)
		require.NoError(t, err)

		content := result.(string)
		assert.Contains(t, content, "1\tline 1")
		assert.Contains(t, content, "5\tline 5")
	})

	t.Run("read with start line", func(t *testing.T) {
		params := map[string]any{
			"path":  testFile,
			"start": 3,
		}

		result, err := tool.Execute(ctx, params, data)
		require.NoError(t, err)

		content := result.(string)
		assert.Contains(t, content, "3\tline 3")
		assert.Contains(t, content, "5\tline 5")
		assert.NotContains(t, content, "1\tline 1")
		assert.NotContains(t, content, "2\tline 2")
	})

	t.Run("read with start and stop lines", func(t *testing.T) {
		params := map[string]any{
			"path":  testFile,
			"start": 2,
			"stop":  4,
		}

		result, err := tool.Execute(ctx, params, data)
		require.NoError(t, err)

		content := result.(string)
		assert.Contains(t, content, "2\tline 2")
		assert.Contains(t, content, "3\tline 3")
		assert.Contains(t, content, "4\tline 4")
		assert.NotContains(t, content, "1\tline 1")
		assert.NotContains(t, content, "5\tline 5")
	})

	t.Run("read with stop line only", func(t *testing.T) {
		params := map[string]any{
			"path": testFile,
			"stop": 3,
		}

		result, err := tool.Execute(ctx, params, data)
		require.NoError(t, err)

		content := result.(string)
		assert.Contains(t, content, "1\tline 1")
		assert.Contains(t, content, "2\tline 2")
		assert.Contains(t, content, "3\tline 3")
		assert.NotContains(t, content, "4\tline 4")
		assert.NotContains(t, content, "5\tline 5")
	})

	t.Run("invalid file path", func(t *testing.T) {
		params := map[string]any{
			"path": "/nonexistent/file.log",
		}

		_, err := tool.Execute(ctx, params, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in the collected artifacts")
	})

	t.Run("missing path parameter", func(t *testing.T) {
		params := map[string]any{}

		_, err := tool.Execute(ctx, params, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parameter 'path' is required")
	})

	t.Run("invalid start line", func(t *testing.T) {
		params := map[string]any{
			"path":  testFile,
			"start": 0,
		}

		_, err := tool.Execute(ctx, params, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start line must be >= 1")
	})

	t.Run("invalid stop line", func(t *testing.T) {
		params := map[string]any{
			"path": testFile,
			"stop": 0,
		}

		_, err := tool.Execute(ctx, params, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stop line must be >= 1")
	})

	t.Run("start greater than stop", func(t *testing.T) {
		params := map[string]any{
			"path":  testFile,
			"start": 5,
			"stop":  3,
		}

		_, err := tool.Execute(ctx, params, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start line (5) cannot be greater than stop line (3)")
	})

	t.Run("line range beyond file", func(t *testing.T) {
		params := map[string]any{
			"path":  testFile,
			"start": 10,
			"stop":  15,
		}

		result, err := tool.Execute(ctx, params, data)
		require.NoError(t, err)

		content := result.(string)
		assert.Contains(t, content, "No lines found in range 10-15")
	})

	t.Run("nil data", func(t *testing.T) {
		params := map[string]any{
			"path": testFile,
		}

		_, err := tool.Execute(ctx, params, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no data provided to tool")
	})
}

func TestReadFileTool_ExtractInteger(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		paramName string
		expected  int
		wantError bool
	}{
		{"float64", float64(42), "test", 42, false},
		{"int", int(42), "test", 42, false},
		{"int64", int64(42), "test", 42, false},
		{"string", "42", "test", 0, true},
		{"bool", true, "test", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractInteger(tt.value, tt.paramName)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.paramName)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestReadFileWithLineRange(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	testContent := `first line
second line
third line
fourth line
fifth line`

	err := os.WriteFile(testFile, []byte(testContent), 0o644)
	require.NoError(t, err)

	t.Run("no range specified", func(t *testing.T) {
		content, err := readFileWithLineRange(testFile, nil, nil)
		require.NoError(t, err)

		assert.Contains(t, content, "1\tfirst line")
		assert.Contains(t, content, "5\tfifth line")
	})

	t.Run("start only", func(t *testing.T) {
		start := 3
		content, err := readFileWithLineRange(testFile, &start, nil)
		require.NoError(t, err)

		assert.Contains(t, content, "3\tthird line")
		assert.Contains(t, content, "5\tfifth line")
		assert.NotContains(t, content, "1\tfirst line")
	})

	t.Run("range specified", func(t *testing.T) {
		start := 2
		stop := 4
		content, err := readFileWithLineRange(testFile, &start, &stop)
		require.NoError(t, err)

		assert.Contains(t, content, "2\tsecond line")
		assert.Contains(t, content, "3\tthird line")
		assert.Contains(t, content, "4\tfourth line")
		assert.NotContains(t, content, "1\tfirst line")
		assert.NotContains(t, content, "5\tfifth line")
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := readFileWithLineRange("/nonexistent/file.log", nil, nil)
		assert.Error(t, err)
	})
}
