package llm

// Helper functions for creating pointers to primitive types
// These make it easier to create AnalysisConfig structs

func StringPtr(s string) *string {
	return &s
}

func Float32Ptr(f float32) *float32 {
	return &f
}

func IntPtr(i int) *int {
	return &i
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