package llm

import (
	"context"
	"os"
	"testing"

	"github.com/openshift/osde2e/internal/aggregator"
	"github.com/openshift/osde2e/internal/llm/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestGeminiClient_ImplementsInterface(t *testing.T) {
	var _ LLMClient = (*GeminiClient)(nil)
}

func TestGeminiClient_ModelSupported(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	require.NoError(t, err)

	model, err := client.Models.Get(ctx, DefaultModel, nil)
	require.NoError(t, err, "model %q is not available in the Gemini API", DefaultModel)
	assert.Equal(t, "models/"+DefaultModel, model.Name)
	t.Logf("Model %q is available", DefaultModel)
}

func TestGeminiClient_Integration(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	client, err := NewGeminiClient(ctx, apiKey)
	require.NoError(t, err)

	t.Run("with no config", func(t *testing.T) {
		result, err := client.Analyze(ctx, "What is 2+2?", nil, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Content)
		t.Logf("Response: %s", result.Content)
	})

	t.Run("with basic config", func(t *testing.T) {
		config := &AnalysisConfig{
			SystemInstruction: genai.Ptr("You are a helpful math assistant."),
			Temperature:       genai.Ptr[float32](0.1),
		}

		result, err := client.Analyze(ctx, "What is 5+5?", config, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Content)
		t.Logf("Response with config: %s", result.Content)
	})
}

type dummyTool struct{}

func (d *dummyTool) Name() string { return "get_next_number" }
func (d *dummyTool) Description() string {
	return "Returns the next number in a sequence. Must be called repeatedly to get all numbers."
}

func (d *dummyTool) Schema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"current": {Type: genai.TypeInteger, Description: "The current number"},
		},
		Required: []string{"current"},
	}
}

func (d *dummyTool) Execute(_ context.Context, params map[string]any, _ []aggregator.LogEntry) (any, error) {
	current, _ := params["current"].(float64)
	return map[string]any{"next": current + 1}, nil
}

func TestGeminiClient_ForcedFinalResponse(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	client, err := NewGeminiClient(ctx, apiKey)
	require.NoError(t, err)

	registry := &tools.Registry{}
	*registry = *tools.NewRegistry(nil)
	registry.Register(&dummyTool{})

	config := &AnalysisConfig{
		SystemInstruction: genai.Ptr(
			"You have a tool called get_next_number. " +
				"Always call it to get the next number before responding. " +
				"Keep calling it to count up. " +
				"When you can no longer call tools, summarize all the numbers you collected."),
		Temperature: genai.Ptr[float32](0.0),
	}

	result, err := client.Analyze(ctx, "Start counting from 1 using the get_next_number tool. Call it each turn.", config, registry)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Content, "expected final forced response to contain text")
	assert.NotEmpty(t, result.ToolCalls, "expected tool calls to be tracked")
	t.Logf("Tool calls made: %d", len(result.ToolCalls))
	t.Logf("Final response: %s", result.Content)
}
