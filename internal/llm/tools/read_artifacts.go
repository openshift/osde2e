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

func (t *readArtifactsTool) Name() string {
	return "read_artifacts"
}

func (t *readArtifactsTool) Description() string {
	return "Reads and returns specific artifacts from the current osde2e test run. Supports reading by artifact type: failed_tests or read_file"
}

func (t *readArtifactsTool) Schema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"artifact_type": {
				Type:        genai.TypeString,
				Description: "The type of artifacts to return. Options: failed_tests, read_file",
				Enum:        []string{"failed_tests", "read_file"},
			},
			"file_path": {
				Type:        genai.TypeString,
				Description: "File path to read (required for read_file)",
			},
		},
		Required: []string{"artifact_type"},
	}
}

func (t *readArtifactsTool) Execute(ctx context.Context, params map[string]any, data *aggregator.AggregatedData) (any, error) {
	artifactType, err := extractString(params, "artifact_type")
	if err != nil {
		return nil, err
	}
	fmt.Println("Executing read_artifacts with params: ", params)

	if data == nil {
		return nil, fmt.Errorf("no data provided to tool")
	}

	// Handle file reading case
	if artifactType == "read_file" {
		filePath, err := extractString(params, "file_path")
		if err != nil {
			return nil, err
		}

		// Validate that the file path exists in the collected logs
		if !isValidLogFile(filePath, data.LogArtifacts) {
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
	case "failed_tests":
		result = data.FailedTests
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s. Supported types are: failed_tests, read_file", artifactType)
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
