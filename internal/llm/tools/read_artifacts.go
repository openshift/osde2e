package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openshift/osde2e/internal/aggregator"
	"google.golang.org/genai"
)

type readArtifactsTool struct{}

func (t *readArtifactsTool) name() string {
	return "read_artifacts"
}

func (t *readArtifactsTool) description() string {
	return "Reads and returns specific artifacts and metadata from the current osde2e test run. Supports reading by artifact type: metadata, failed_tests, test_results, or read_file"
}

func (t *readArtifactsTool) schema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"artifact_type": {
				Type:        genai.TypeString,
				Description: "The type of artifacts to return. Options: metadata, failed_tests, test_results, read_file",
				Enum:        []string{"metadata", "failed_tests", "test_results", "read_file"},
			},
			"file_path": {
				Type:        genai.TypeString,
				Description: "File path to read (required for read_file)",
			},
		},
		Required: []string{"artifact_type"},
	}
}

func (t *readArtifactsTool) execute(ctx context.Context, params map[string]any) (any, error) {
	artifactType, err := extractString(params, "artifact_type")
	if err != nil {
		return nil, err
	}
	fmt.Println("Executing read_artifacts with params: ", params)

	if globalCollectedData == nil {
		return nil, fmt.Errorf("no data collected - call SetCollectedData first")
	}

	data := globalCollectedData

	// Handle file reading case
	if artifactType == "read_file" {
		filePath, err := extractString(params, "file_path")
		if err != nil {
			return nil, err
		}

		// Validate that the file path exists in the collected logs
		if !isValidLogFile(filePath, data.Logs) {
			content := fmt.Sprintf("file path %s is not in the collected artifacts, request valid path", filePath)
			fmt.Println(content)
			return content, nil
		}

		content, err := readFileAsString(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		return content, nil
	}

	// Handle standard artifact types
	var result any
	switch artifactType {
	case "metadata":
		result = data.Metadata
	case "test_results":
		result = data.TestResults
	case "failed_tests":
		result = data.FailedTests
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s. Supported types are: metadata, failed_tests, test_results, read_file", artifactType)
	}

	// Convert to JSON for clean serialization
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize artifacts data: %w", err)
	}
	return string(jsonData), nil
}

// extractString extracts a string parameter from the params map
func extractString(params map[string]any, key string) (string, error) {
	val, ok := params[key]
	if !ok {
		return "", fmt.Errorf("parameter '%s' is required", key)
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter '%s' must be a string, got %T", key, val)
	}

	return str, nil
}

// isValidLogFile checks if the given file path exists in the collected logs
func isValidLogFile(filePath string, logs []aggregator.LogEntry) bool {
	for _, log := range logs {
		if log.Source == filePath {
			return true
		}
	}
	return false
}

// readFileAsString reads a file and returns its content as a string
func readFileAsString(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
