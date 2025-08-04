package llm

import (
	"context"
	"fmt"
)

// MockClient implements LLMClient for testing
type MockClient struct {
	responses     []string
	responseIndex int
	shouldError   bool
	errorMessage  string
	closed        bool
}

// NewMockClient creates a new mock client with predefined responses
func NewMockClient(responses []string) *MockClient {
	return &MockClient{
		responses:     responses,
		responseIndex: 0,
		shouldError:   false,
	}
}

// WithError configures the mock to return an error
func (m *MockClient) WithError(errorMessage string) *MockClient {
	m.shouldError = true
	m.errorMessage = errorMessage
	return m
}

// Analyze returns the next predefined response or an error if configured
func (m *MockClient) Analyze(ctx context.Context, userPrompt string, config ...*AnalysisConfig) (*AnalysisResult, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	if m.shouldError {
		return nil, fmt.Errorf("mock error: %s", m.errorMessage)
	}

	if len(m.responses) == 0 {
		return nil, fmt.Errorf("no mock responses configured")
	}

	// Cycle through responses
	response := m.responses[m.responseIndex%len(m.responses)]
	m.responseIndex++

	// Add config info to response if system instruction is provided
	content := response
	if len(config) > 0 && config[0] != nil && config[0].SystemInstruction != nil {
		content = fmt.Sprintf("[System: %s] %s", *config[0].SystemInstruction, response)
	}

	return &AnalysisResult{
		Content:  content,
		Provider: "mock",
		Model:    "mock-model",
	}, nil
}

// HealthCheck always returns nil unless the client is closed or configured to error
func (m *MockClient) HealthCheck(ctx context.Context) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	if m.shouldError {
		return fmt.Errorf("mock health check error: %s", m.errorMessage)
	}

	return nil
}

// Close marks the client as closed
func (m *MockClient) Close() error {
	m.closed = true
	return nil
}

// Reset resets the response index and error state
func (m *MockClient) Reset() {
	m.responseIndex = 0
	m.shouldError = false
	m.errorMessage = ""
	m.closed = false
}

// AddResponse adds a new response to the mock
func (m *MockClient) AddResponse(response string) {
	m.responses = append(m.responses, response)
}