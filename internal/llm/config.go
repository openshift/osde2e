package llm

type AnalysisConfig struct {
	SystemInstruction *string  `json:"systemInstruction,omitempty"`
	Temperature       *float32 `json:"temperature,omitempty"`
	TopP              *float32 `json:"topP,omitempty"`
	MaxTokens         *int     `json:"maxTokens,omitempty"`
}

type AnalysisResult struct {
	Content string
}
