package llm

func StringPtr(s string) *string {
	return &s
}

func Float32Ptr(f float32) *float32 {
	return &f
}

func IntPtr(i int) *int {
	return &i
}

// AnalysisConfig contains configuration for LLM requests
type AnalysisConfig struct {
	SystemInstruction *string  `json:"systemInstruction,omitempty"`
	Temperature       *float32 `json:"temperature,omitempty"`
	TopP              *float32 `json:"topP,omitempty"`
	ResponseSchema    any      `json:"responseSchema,omitempty"`
	Tools             []any    `json:"tools,omitempty"`
	MaxTokens         *int     `json:"maxTokens,omitempty"`
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
		Temperature: Float32Ptr(0.7),
		TopP:        Float32Ptr(0.9),
		MaxTokens:   IntPtr(1000),
	}
}

// NewConfigWithSystem creates a config with just a system instruction
func NewConfigWithSystem(systemInstruction string) *AnalysisConfig {
	config := DefaultConfig()
	config.SystemInstruction = StringPtr(systemInstruction)
	return config
}

// NewConfigWithTemperature creates a config with custom temperature
func NewConfigWithTemperature(temp float32) *AnalysisConfig {
	config := DefaultConfig()
	config.Temperature = Float32Ptr(temp)
	return config
}
