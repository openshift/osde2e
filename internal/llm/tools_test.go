package llm

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/genai"
)

// loadEnv loads environment variables from .env file
func loadEnv() error {
	file, err := os.Open("../../.env")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func TestGetCurrentTime(t *testing.T) {
	// Load environment variables from .env file
	if err := loadEnv(); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
		// Fallback to test value if .env file is not available
		os.Setenv("GEMINI_API_KEY", "test")
	}

	// Check if we have a valid API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "test" {
		t.Skip("Using test API key, skipping actual API call")
	}

	tools := []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "GetCurrentTime",
					Description: "Get the current time",
					Parameters:  nil,
				},
			},
		},
	}

	ac := &AnalysisConfig{
		SystemInstruction: genai.Ptr("If you don't know the answer call for the tool GetCurrentTime. Once you have the time, return the time in the format YYYY-MM-DD HH:MM:SS."),
		Temperature:       genai.Ptr[float32](0.1),
		Tools:             tools,
	}

	ctx := context.Background()

	gc, err := NewGeminiClient(ctx, apiKey)
	if err != nil {
		t.Fatalf("failed to create Gemini client: %v", err)
	}

	result, err := gc.Analyze(ctx, "What is user system current time?", ac)
	if err != nil {
		t.Fatalf("failed to analyze: %v", err)
	}

	fmt.Println(result.Content)

	// Check if the result contains a date in the expected format
	if !containsDateFormat(result.Content) {
		t.Errorf("Result does not contain date in format YYYY-MM-DD HH:MM:SS. Content: %s", result.Content)
	} else {
		t.Log("Result contains date in correct format")
	}
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
