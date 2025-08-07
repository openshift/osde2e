package tools

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type addNumbersTool struct{}

func (t *addNumbersTool) name() string {
	return "add_numbers"
}

func (t *addNumbersTool) description() string {
	return "Adds two numbers together and returns the result"
}

func (t *addNumbersTool) schema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"a": {
				Type:        genai.TypeNumber,
				Description: "The first number to add",
			},
			"b": {
				Type:        genai.TypeNumber,
				Description: "The second number to add",
			},
		},
		Required: []string{"a", "b"},
	}
}

func (t *addNumbersTool) execute(ctx context.Context, params map[string]any) (any, error) {
	a, err := extractNumber(params, "a")
	if err != nil {
		return nil, err
	}

	b, err := extractNumber(params, "b")
	if err != nil {
		return nil, err
	}

	return a + b, nil
}

func extractNumber(params map[string]any, key string) (float64, error) {
	val, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("parameter '%s' is required", key)
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("parameter '%s' must be a number, got %T", key, val)
	}
}
