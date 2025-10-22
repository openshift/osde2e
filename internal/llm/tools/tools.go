package tools

import (
	"context"
	"fmt"

	"github.com/openshift/osde2e/internal/aggregator"
	"google.golang.org/genai"
)

// RegistryConfig holds configuration for the tool registry
type RegistryConfig struct {
	EnableMustGather bool
}

// Tool represents an internal tool interface
type Tool interface {
	Name() string
	Description() string
	Schema() *genai.Schema
	Execute(ctx context.Context, params map[string]any, data *aggregator.AggregatedData) (any, error)
}

// Registry manages available tools with their dependencies
type Registry struct {
	tools     map[string]Tool
	data      *aggregator.AggregatedData
	cleanupFn func() error
}

// NewRegistry creates a new tool registry with the provided data and config
func NewRegistry(data *aggregator.AggregatedData, config *RegistryConfig) *Registry {
	r := &Registry{
		tools: make(map[string]Tool),
		data:  data,
	}

	// Register production tools only
	r.Register(&readFileTool{})

	// Register must-gather tool if enabled and must-gather tar file is available
	if config != nil && config.EnableMustGather {
		if mustGatherPath := findMustGatherTar(data.LogArtifacts); mustGatherPath != "" {
			// Initialize the tool immediately during registry creation
			tool, err := newMustGatherTool(context.Background(), mustGatherPath)
			if err != nil {
				// Log warning but don't fail registry creation
				fmt.Printf("Warning: Failed to initialize must-gather tool: %v\n", err)
			} else {
				r.Register(tool)
				// Set up cleanup function for the registry
				r.cleanupFn = tool.Cleanup
			}
		}
	}

	return r
}

// Register adds a tool to the registry
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// GetTools returns all registered tools as genai.Tool slice
func (r *Registry) GetTools() []*genai.Tool {
	tools := make([]*genai.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  tool.Schema(),
				},
			},
		})
	}
	return tools
}

// Execute runs a tool by name with given parameters
func (r *Registry) Execute(ctx context.Context, name string, params map[string]any) (any, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return tool.Execute(ctx, params, r.data)
}

// HandleToolCall processes a function call and returns the appropriate content
func (r *Registry) HandleToolCall(ctx context.Context, functionCall *genai.FunctionCall) (*genai.Content, error) {
	result, err := r.Execute(ctx, functionCall.Name, functionCall.Args)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	response := fmt.Sprintf("Tool %s result: %q", functionCall.Name, result)
	return genai.NewContentFromText(response, genai.RoleUser), nil
}

// Cleanup cleans up resources used by tools in the registry
func (r *Registry) Cleanup() error {
	if r.cleanupFn != nil {
		return r.cleanupFn()
	}
	return nil
}
