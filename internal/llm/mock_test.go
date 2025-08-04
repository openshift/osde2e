package llm

import (
	"context"
	"testing"
)

func TestMockClient_Analyze(t *testing.T) {
	t.Run("single response", func(t *testing.T) {
		client := NewMockClient([]string{"Hello, world!"})
		defer client.Close()

		result, err := client.Analyze(context.Background(), "test prompt")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Content != "Hello, world!" {
			t.Errorf("Expected 'Hello, world!', got '%s'", result.Content)
		}

		if result.Provider != "mock" {
			t.Errorf("Expected provider 'mock', got '%s'", result.Provider)
		}
	})

	t.Run("multiple responses cycle", func(t *testing.T) {
		client := NewMockClient([]string{"Response 1", "Response 2"})
		defer client.Close()

		// First call
		result1, err := client.Analyze(context.Background(), "test 1")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if result1.Content != "Response 1" {
			t.Errorf("Expected 'Response 1', got '%s'", result1.Content)
		}

		// Second call
		result2, err := client.Analyze(context.Background(), "test 2")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if result2.Content != "Response 2" {
			t.Errorf("Expected 'Response 2', got '%s'", result2.Content)
		}

		// Third call should cycle back to first
		result3, err := client.Analyze(context.Background(), "test 3")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if result3.Content != "Response 1" {
			t.Errorf("Expected 'Response 1', got '%s'", result3.Content)
		}
	})

	t.Run("with system instruction", func(t *testing.T) {
		client := NewMockClient([]string{"Mock response"})
		defer client.Close()

		config := &AnalysisConfig{
			SystemInstruction: StringPtr("You are a helpful assistant"),
		}

		result, err := client.Analyze(context.Background(), "test", config)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		expected := "[System: You are a helpful assistant] Mock response"
		if result.Content != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result.Content)
		}
	})

	t.Run("with error", func(t *testing.T) {
		client := NewMockClient([]string{"Response"}).WithError("test error")
		defer client.Close()

		_, err := client.Analyze(context.Background(), "test")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if err.Error() != "mock error: test error" {
			t.Errorf("Expected 'mock error: test error', got '%s'", err.Error())
		}
	})

	t.Run("closed client", func(t *testing.T) {
		client := NewMockClient([]string{"Response"})
		client.Close()

		_, err := client.Analyze(context.Background(), "test")
		if err == nil {
			t.Fatal("Expected error for closed client, got nil")
		}
	})
}

func TestMockClient_HealthCheck(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		client := NewMockClient([]string{})
		defer client.Close()

		err := client.HealthCheck(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("with error", func(t *testing.T) {
		client := NewMockClient([]string{}).WithError("health check failed")
		defer client.Close()

		err := client.HealthCheck(context.Background())
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("closed client", func(t *testing.T) {
		client := NewMockClient([]string{})
		client.Close()

		err := client.HealthCheck(context.Background())
		if err == nil {
			t.Fatal("Expected error for closed client, got nil")
		}
	})
}

func TestMockClient_Reset(t *testing.T) {
	client := NewMockClient([]string{"Response 1", "Response 2"}).WithError("test error")
	
	// Use up first response
	_, _ = client.Analyze(context.Background(), "test")
	
	// Reset
	client.Reset()
	
	// Should start from first response again and not error
	result, err := client.Analyze(context.Background(), "test")
	if err != nil {
		t.Fatalf("Expected no error after reset, got: %v", err)
	}
	
	if result.Content != "Response 1" {
		t.Errorf("Expected 'Response 1' after reset, got '%s'", result.Content)
	}
}

func TestMockClient_AddResponse(t *testing.T) {
	client := NewMockClient([]string{"Original"})
	defer client.Close()
	
	client.AddResponse("Added")
	
	// First call gets original
	result1, _ := client.Analyze(context.Background(), "test")
	if result1.Content != "Original" {
		t.Errorf("Expected 'Original', got '%s'", result1.Content)
	}
	
	// Second call gets added
	result2, _ := client.Analyze(context.Background(), "test")
	if result2.Content != "Added" {
		t.Errorf("Expected 'Added', got '%s'", result2.Content)
	}
}

// Test that MockClient implements LLMClient interface
func TestMockClient_ImplementsInterface(t *testing.T) {
	var _ LLMClient = (*MockClient)(nil)
}