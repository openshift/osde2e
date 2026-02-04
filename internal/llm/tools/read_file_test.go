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

	// Create log artifacts with the test file
	logArtifacts := []aggregator.LogEntry{
		{Source: testFile},
	}

	tool := &readFileTool{}
	ctx := context.Background()

	t.Run("read entire file", func(t *testing.T) {
		params := map[string]any{
			"path": testFile,
		}

		result, err := tool.Execute(ctx, params, logArtifacts)
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

		result, err := tool.Execute(ctx, params, logArtifacts)
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

		result, err := tool.Execute(ctx, params, logArtifacts)
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

		result, err := tool.Execute(ctx, params, logArtifacts)
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

		_, err := tool.Execute(ctx, params, logArtifacts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in the collected artifacts")
	})

	t.Run("missing path parameter", func(t *testing.T) {
		params := map[string]any{}

		_, err := tool.Execute(ctx, params, logArtifacts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parameter 'path' is required")
	})

	t.Run("invalid start line", func(t *testing.T) {
		params := map[string]any{
			"path":  testFile,
			"start": 0,
		}

		_, err := tool.Execute(ctx, params, logArtifacts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start line must be >= 1")
	})

	t.Run("invalid stop line", func(t *testing.T) {
		params := map[string]any{
			"path": testFile,
			"stop": 0,
		}

		_, err := tool.Execute(ctx, params, logArtifacts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stop line must be >= 1")
	})

	t.Run("start greater than stop", func(t *testing.T) {
		params := map[string]any{
			"path":  testFile,
			"start": 5,
			"stop":  3,
		}

		_, err := tool.Execute(ctx, params, logArtifacts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start line (5) cannot be greater than stop line (3)")
	})

	t.Run("line range beyond file", func(t *testing.T) {
		params := map[string]any{
			"path":  testFile,
			"start": 10,
			"stop":  15,
		}

		result, err := tool.Execute(ctx, params, logArtifacts)
		require.NoError(t, err)

		content := result.(string)
		assert.Contains(t, content, "No lines found in range 10-15")
	})

	t.Run("nil log artifacts", func(t *testing.T) {
		params := map[string]any{
			"path": testFile,
		}

		_, err := tool.Execute(ctx, params, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no log artifacts provided to tool")
	})
}

func TestReadFileTool_ExtractIntPtr(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]any
		key      string
		expected *int
	}{
		{"float64", map[string]any{"test": float64(42)}, "test", func() *int { v := 42; return &v }()},
		{"int", map[string]any{"test": int(42)}, "test", func() *int { v := 42; return &v }()},
		{"int64", map[string]any{"test": int64(42)}, "test", func() *int { v := 42; return &v }()},
		{"missing", map[string]any{}, "test", nil},
		{"invalid", map[string]any{"test": "42"}, "test", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIntPtr(tt.params, tt.key)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
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

	tool := newReadFileTool()

	t.Run("no range specified", func(t *testing.T) {
		content, err := tool.readFileWithLineRange(testFile, nil, nil, false)
		require.NoError(t, err)

		assert.Contains(t, content, "1\tfirst line")
		assert.Contains(t, content, "5\tfifth line")
	})

	t.Run("start only", func(t *testing.T) {
		start := 3
		content, err := tool.readFileWithLineRange(testFile, &start, nil, false)
		require.NoError(t, err)

		assert.Contains(t, content, "3\tthird line")
		assert.Contains(t, content, "5\tfifth line")
		assert.NotContains(t, content, "1\tfirst line")
	})

	t.Run("range specified", func(t *testing.T) {
		start := 2
		stop := 4
		content, err := tool.readFileWithLineRange(testFile, &start, &stop, false)
		require.NoError(t, err)

		assert.Contains(t, content, "2\tsecond line")
		assert.Contains(t, content, "3\tthird line")
		assert.Contains(t, content, "4\tfourth line")
		assert.NotContains(t, content, "1\tfirst line")
		assert.NotContains(t, content, "5\tfifth line")
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := tool.readFileWithLineRange("/nonexistent/file.log", nil, nil, false)
		assert.Error(t, err)
	})
}
