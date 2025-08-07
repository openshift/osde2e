package tools

import (
	"fmt"
	"time"

	"google.golang.org/genai"
)

// HandleToolCall processes a function call and returns the appropriate content
func HandleToolCall(functionCall *genai.FunctionCall) (*genai.Content, error) {
	switch functionCall.Name {
	case "GetCurrentTime":
		currentTime := time.Now().Format(time.RFC3339)
		return genai.NewContentFromText(fmt.Sprintf("The current time is: %s", currentTime), genai.RoleUser), nil
	default:
		return nil, fmt.Errorf("unknown function: %s", functionCall.Name)
	}
}
