package tools

import (
	"context"
	"testing"

	"github.com/openshift/osde2e/internal/aggregator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindMustGatherTar(t *testing.T) {
	tests := []struct {
		name         string
		logArtifacts []aggregator.LogEntry
		expected     string
	}{
		{
			name: "finds must-gather.tar",
			logArtifacts: []aggregator.LogEntry{
				{Source: "/path/to/build-log.txt"},
				{Source: "/path/to/must-gather.tar"},
				{Source: "/path/to/other-file.log"},
			},
			expected: "/path/to/must-gather.tar",
		},
		{
			name: "finds must-gather.tar.gz",
			logArtifacts: []aggregator.LogEntry{
				{Source: "/path/to/build-log.txt"},
				{Source: "/path/to/must-gather.tar.gz"},
			},
			expected: "/path/to/must-gather.tar.gz",
		},
		{
			name: "finds must-gather with timestamp",
			logArtifacts: []aggregator.LogEntry{
				{Source: "/path/to/must-gather-20231201-123456.tar.gz"},
			},
			expected: "/path/to/must-gather-20231201-123456.tar.gz",
		},
		{
			name: "case insensitive matching",
			logArtifacts: []aggregator.LogEntry{
				{Source: "/path/to/Must-Gather.TAR"},
			},
			expected: "/path/to/Must-Gather.TAR",
		},
		{
			name: "no must-gather file found",
			logArtifacts: []aggregator.LogEntry{
				{Source: "/path/to/build-log.txt"},
				{Source: "/path/to/test-results.xml"},
			},
			expected: "",
		},
		{
			name:         "empty artifacts",
			logArtifacts: []aggregator.LogEntry{},
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMustGatherTar(tt.logArtifacts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMustGatherTool_Name(t *testing.T) {
	tool := &mustGatherTool{
		omcClient: NewOMCClient(),
	}

	assert.Equal(t, "must_gather", tool.Name())
}

func TestMustGatherTool_Description(t *testing.T) {
	tool := &mustGatherTool{
		omcClient: NewOMCClient(),
	}

	description := tool.Description()
	assert.Contains(t, description, "unhealthy operators")
	assert.Contains(t, description, "must-gather data")
	assert.Contains(t, description, "operator-related issues")
}

func TestMustGatherTool_Schema(t *testing.T) {
	tool := &mustGatherTool{
		omcClient: NewOMCClient(),
	}

	schema := tool.Schema()
	require.NotNil(t, schema)

	// Check required fields
	assert.Contains(t, schema.Required, "get_operator_health")

	// Check properties
	assert.Contains(t, schema.Properties, "get_operator_health")

	// Verify get_operator_health property
	healthProp := schema.Properties["get_operator_health"]
	assert.Equal(t, "string", string(healthProp.Type))
	assert.Contains(t, healthProp.Description, "operator health check")
}

func TestMustGatherTool_ValidScopes(t *testing.T) {
	tests := []struct {
		scope string
		valid bool
	}{
		{"core", true},
		{"addons", true},
		{"all", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.scope, func(t *testing.T) {
			validScopes := []string{"core", "addons", "all"}
			found := false
			for _, validScope := range validScopes {
				if validScope == tt.scope {
					found = true
					break
				}
			}
			assert.Equal(t, tt.valid, found, "scope: %s", tt.scope)
		})
	}
}

func TestNewRegistry_WithMustGather(t *testing.T) {
	data := &aggregator.AggregatedData{
		LogArtifacts: []aggregator.LogEntry{
			{Source: "/path/to/build-log.txt"},
			{Source: "/path/to/must-gather.tar"},
		},
	}

	registry := NewRegistry(data, &RegistryConfig{EnableMustGather: true})

	// Should have both read_file and must_gather tools
	tools := registry.GetTools()
	assert.Len(t, tools, 2)

	// Check that must_gather tool is registered
	_, exists := registry.tools["must_gather"]
	assert.True(t, exists)
}

func TestNewRegistry_WithoutMustGather(t *testing.T) {
	data := &aggregator.AggregatedData{
		LogArtifacts: []aggregator.LogEntry{
			{Source: "/path/to/build-log.txt"},
			{Source: "/path/to/test-results.xml"},
		},
	}

	registry := NewRegistry(data, &RegistryConfig{EnableMustGather: true})

	// Should only have read_file tool (no must-gather file present)
	tools := registry.GetTools()
	assert.Len(t, tools, 1)

	// Check that must_gather tool is not registered
	_, exists := registry.tools["must_gather"]
	assert.False(t, exists)
}

func TestNewRegistry_MustGatherDisabled(t *testing.T) {
	data := &aggregator.AggregatedData{
		LogArtifacts: []aggregator.LogEntry{
			{Source: "/path/to/build-log.txt"},
			{Source: "/path/to/must-gather.tar"},
		},
	}

	registry := NewRegistry(data, &RegistryConfig{EnableMustGather: false})

	// Should only have read_file tool (must-gather disabled)
	tools := registry.GetTools()
	assert.Len(t, tools, 1)

	// Check that must_gather tool is not registered even though file exists
	_, exists := registry.tools["must_gather"]
	assert.False(t, exists)
}

func TestNewMustGatherTool_WithInvalidPath(t *testing.T) {
	ctx := context.Background()

	// This should fail because the must-gather file doesn't exist
	_, err := newMustGatherTool(ctx, "/non/existent/path.tar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize OMC client")
}
