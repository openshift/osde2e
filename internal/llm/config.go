package llm

import (
	"google.golang.org/genai"
)

// AnalysisConfig contains configuration for LLM requests
type AnalysisConfig struct {
	SystemInstruction *string       `json:"systemInstruction,omitempty"`
	Temperature       *float32      `json:"temperature,omitempty"`
	TopP              *float32      `json:"topP,omitempty"`
	ResponseSchema    *genai.Schema `json:"responseSchema,omitempty"`
	Tools             []any         `json:"tools,omitempty"`
	MaxTokens         *int          `json:"maxTokens,omitempty"`
}

// AnalysisResult contains the response from the LLM
type AnalysisResult struct {
	Content  string
	Provider string
	Model    string
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *AnalysisConfig {
	return &AnalysisConfig{
		Temperature: genai.Ptr[float32](0.7),
		TopP:        genai.Ptr[float32](0.9),
		MaxTokens:   genai.Ptr(1000),
	}
}

// NewConfigWithSystem creates a config with just a system instruction
func NewConfigWithSystem(systemInstruction string) *AnalysisConfig {
	config := DefaultConfig()
	config.SystemInstruction = genai.Ptr(systemInstruction)
	return config
}

// NewConfigWithTemperature creates a config with custom temperature
func NewConfigWithTemperature(temp float32) *AnalysisConfig {
	config := DefaultConfig()
	config.Temperature = genai.Ptr(temp)
	return config
}
