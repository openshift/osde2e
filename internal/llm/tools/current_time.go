package tools

import (
	"context"
	"time"

	"google.golang.org/genai"
)

type currentTimeTool struct{}

func (t *currentTimeTool) name() string {
	return "get_current_time"
}

func (t *currentTimeTool) description() string {
	return "Returns the current date and time in RFC3339 format"
}

func (t *currentTimeTool) schema() *genai.Schema {
	return &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	}
}

func (t *currentTimeTool) execute(ctx context.Context, params map[string]any) (any, error) {
	return time.Now().Format(time.RFC3339), nil
}
