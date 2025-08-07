package llm

import (
	"context"
	"os"
	"regexp"
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
		result, err := client.Analyze(ctx, "What is 2+2?", nil)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Content)
		t.Logf("Response: %s", result.Content)
	})

	t.Run("with basic config", func(t *testing.T) {
		config := &AnalysisConfig{
			SystemInstruction: genai.Ptr("You are a helpful math assistant."),
			Temperature:       genai.Ptr[float32](0.1),
		}

		result, err := client.Analyze(ctx, "What is 5+5?", config)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Content)
		t.Logf("Response with config: %s", result.Content)
	})

	t.Run("with tools available", func(t *testing.T) {
		config := &AnalysisConfig{
			SystemInstruction: genai.Ptr("If you don't know the answer call for the tool get_current_time. Once you have the time, return the time in the format YYYY-MM-DD HH:MM:SS."),
			Temperature:       genai.Ptr[float32](0.1),
			EnableTools:       true,
		}

		result, err := client.Analyze(ctx, "What is user system current time?", config)
		if err != nil {
			t.Fatalf("failed to analyze: %v", err)
		}

		// Check if the result contains a date in the expected format
		if !containsDateFormat(result.Content) {
			t.Errorf("Result does not contain date in format YYYY-MM-DD HH:MM:SS. Content: %s", result.Content)
		} else {
			t.Logf("Result contains date in correct format: %s", result.Content)
		}
	})
}

// containsDateFormat checks if the content contains a date in YYYY-MM-DD HH:MM:SS format
func containsDateFormat(content string) bool {
	// Regex pattern for YYYY-MM-DD HH:MM:SS format
	pattern := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`
	matched, err := regexp.MatchString(pattern, content)
	if err != nil {
		return false
	}
	return matched
}
