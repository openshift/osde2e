package tools

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// tool represents an internal tool interface
type tool interface {
	name() string
	description() string
	schema() *genai.Schema
	execute(ctx context.Context, params map[string]any) (any, error)
}

// registry holds all available tools
var registry = make(map[string]tool)

// register adds a tool to the registry
func register(t tool) {
	registry[t.name()] = t
}

// GetTools returns all registered tools as genai.Tool slice
func GetTools() []*genai.Tool {
	tools := make([]*genai.Tool, 0, len(registry))
	for _, tool := range registry {
		tools = append(tools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        tool.name(),
					Description: tool.description(),
					Parameters:  tool.schema(),
				},
			},
		})
	}
	return tools
}

// execute runs a tool by name with given parameters
func execute(ctx context.Context, name string, params map[string]any) (any, error) {
	tool, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return tool.execute(ctx, params)
}

// HandleToolCall processes a function call and returns the appropriate content
func HandleToolCall(ctx context.Context, functionCall *genai.FunctionCall) (*genai.Content, error) {
	result, err := execute(ctx, functionCall.Name, functionCall.Args)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	response := fmt.Sprintf("Tool %s result: %v", functionCall.Name, result)
	return genai.NewContentFromText(response, genai.RoleUser), nil
}

// init registers default tools
func init() {
	register(&currentTimeTool{})
}
