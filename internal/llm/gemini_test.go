package llm

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestGeminiClient_Analyze(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := NewGeminiClient(ctx, apiKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("simple prompt", func(t *testing.T) {
		result, err := client.Analyze(ctx, "What is 2+2?")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Content == "" {
			t.Fatal("Expected content, got empty string")
		}

		if result.Provider != "gemini" {
			t.Errorf("Expected provider 'gemini', got '%s'", result.Provider)
		}

		t.Logf("Response: %s", result.Content)
	})

	t.Run("with system instruction", func(t *testing.T) {
		systemInstruction := "You are a helpful math tutor. Always show your work step by step."
		userPrompt := "What is 15 * 23?"

		config := &AnalysisConfig{
			SystemInstruction: &systemInstruction,
		}

		result, err := client.Analyze(ctx, userPrompt, config)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Content == "" {
			t.Fatal("Expected content, got empty string")
		}

		t.Logf("Response with system instruction: %s", result.Content)
	})

	t.Run("with temperature and topP", func(t *testing.T) {
		temp := float32(0.7)
		topP := float32(0.9)

		config := &AnalysisConfig{
			Temperature: &temp,
			TopP:        &topP,
		}

		result, err := client.Analyze(ctx, "Tell me a creative story about a robot.", config)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Content == "" {
			t.Fatal("Expected content, got empty string")
		}

		t.Logf("Response with temperature/topP: %s", result.Content)
	})

	t.Run("with max tokens", func(t *testing.T) {
		maxTokens := 50

		config := &AnalysisConfig{
			MaxTokens: &maxTokens,
		}

		result, err := client.Analyze(ctx, "Write a long essay about artificial intelligence.", config)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Content == "" {
			t.Fatal("Expected content, got empty string")
		}

		t.Logf("Response with max tokens: %s", result.Content)
	})

	t.Run("full config", func(t *testing.T) {
		systemInstruction := "You are a concise technical writer."
		temp := float32(0.3)
		topP := float32(0.8)
		maxTokens := 100

		config := &AnalysisConfig{
			SystemInstruction: &systemInstruction,
			Temperature:       &temp,
			TopP:              &topP,
			MaxTokens:         &maxTokens,
		}

		result, err := client.Analyze(ctx, "Explain what a REST API is.", config)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Content == "" {
			t.Fatal("Expected content, got empty string")
		}

		t.Logf("Response with full config: %s", result.Content)
	})
}

func TestGeminiClient_HealthCheck(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewGeminiClient(ctx, apiKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	err = client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

func TestGeminiClient_InvalidAPIKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewGeminiClient(ctx, "invalid-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	_, err = client.Analyze(ctx, "test")
	if err == nil {
		t.Fatal("Expected error with invalid API key")
	}
}

// Test that GeminiClient implements LLMClient interface
func TestGeminiClient_ImplementsInterface(t *testing.T) {
	var _ LLMClient = (*GeminiClient)(nil)
}
