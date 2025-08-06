package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestGeminiClient_ImplementsInterface(t *testing.T) {
	var _ LLMClient = (*GeminiClient)(nil)
}

func TestAnalysisConfig(t *testing.T) {
	config := &AnalysisConfig{
		SystemInstruction: genai.Ptr("test system"),
		Temperature:       genai.Ptr[float32](0.5),
		TopP:              genai.Ptr[float32](0.8),
		MaxTokens:         genai.Ptr(100),
	}

	assert.Equal(t, "test system", *config.SystemInstruction)
	assert.Equal(t, float32(0.5), *config.Temperature)
	assert.Equal(t, float32(0.8), *config.TopP)
	assert.Equal(t, 100, *config.MaxTokens)
}

func TestAnalysisResult(t *testing.T) {
	result := &AnalysisResult{
		Content: "test response",
	}

	assert.Equal(t, "test response", result.Content)
}

// Integration tests require GEMINI_API_KEY environment variable
func TestGeminiClient_Integration(t *testing.T) {
	// Skip if no API key - these would be integration tests
	t.Skip("Integration tests require GEMINI_API_KEY")

	ctx := context.Background()
	client, err := NewGeminiClient(ctx, "test-key")
	require.NoError(t, err)

	config := &AnalysisConfig{
		SystemInstruction: genai.Ptr("You are helpful"),
		Temperature:       genai.Ptr[float32](0.1),
	}

	result, err := client.Analyze(ctx, "What is 2+2?", config)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Content)
}
