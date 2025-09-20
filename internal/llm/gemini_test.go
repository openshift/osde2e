package llm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestGeminiClient_ImplementsInterface(t *testing.T) {
	var _ LLMClient = (*GeminiClient)(nil)
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
