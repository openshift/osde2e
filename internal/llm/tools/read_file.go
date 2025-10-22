package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/openshift/osde2e/internal/aggregator"
	"google.golang.org/genai"
)

type readFileTool struct{}

func (t *readFileTool) Name() string {
	return "read_file"
}

func (t *readFileTool) Description() string {
	return "Reads a specific file from the collected artifacts, optionally specifying a line range. Use start and stop parameters to read specific line ranges."
}

func (t *readFileTool) Schema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"path": {
				Type:        genai.TypeString,
				Description: "Path to the file to read (must be from collected artifacts)",
			},
			"start": {
				Type:        genai.TypeInteger,
				Description: "Starting line number (1-based, optional). If not provided, reads from beginning.",
			},
			"stop": {
				Type:        genai.TypeInteger,
				Description: "Ending line number (1-based, optional). If not provided, reads to end.",
			},
		},
		Required: []string{"path"},
	}
}

func (t *readFileTool) Execute(ctx context.Context, params map[string]any, data *aggregator.AggregatedData) (any, error) {
	// Extract path parameter
	path, err := extractString(params, "path")
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, fmt.Errorf("no data provided to tool")
	}

	// Validate that the file path exists in the collected artifacts
	if !isValidLogFile(path, data.LogArtifacts) {
		return nil, fmt.Errorf("file path %s is not in the collected artifacts", path)
	}

	// Extract optional start and stop parameters
	var start, stop *int
	if startVal, exists := params["start"]; exists {
		if startInt, err := extractInteger(startVal, "start"); err != nil {
			return nil, err
		} else {
			start = &startInt
		}
	}

	if stopVal, exists := params["stop"]; exists {
		if stopInt, err := extractInteger(stopVal, "stop"); err != nil {
			return nil, err
		} else {
			stop = &stopInt
		}
	}

	// Validate line range parameters
	if start != nil && *start < 1 {
		return nil, fmt.Errorf("start line must be >= 1, got %d", *start)
	}
	if stop != nil && *stop < 1 {
		return nil, fmt.Errorf("stop line must be >= 1, got %d", *stop)
	}
	if start != nil && stop != nil && *start > *stop {
		return nil, fmt.Errorf("start line (%d) cannot be greater than stop line (%d)", *start, *stop)
	}

	// Read the file with line range support
	content, err := readFileWithLineRange(path, start, stop)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return content, nil
}

// extractInteger extracts an integer parameter from the params map
func extractInteger(val any, paramName string) (int, error) {
	switch v := val.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("parameter '%s' must be an integer, got %T", paramName, val)
	}
}

// readFileWithLineRange reads a file and returns content within the specified line range
func readFileWithLineRange(filePath string, start, stop *int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size to handle very long lines in CI logs (default is 64KB, set to 1MB)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var lines []string
	lineNum := 1

	// Determine actual start and stop values
	startLine := 1
	if start != nil {
		startLine = *start
	}

	for scanner.Scan() {
		line := scanner.Text()

		// If we haven't reached the start line yet, skip
		if lineNum < startLine {
			lineNum++
			continue
		}

		// If we have a stop line and we've exceeded it, break
		if stop != nil && lineNum > *stop {
			break
		}

		// Add line number prefix and collect the line
		lines = append(lines, fmt.Sprintf("%d\t%s", lineNum, line))
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if len(lines) == 0 {
		if start != nil {
			return fmt.Sprintf("No lines found in range %d-%s", startLine, formatStopLine(stop)), nil
		}
		return "File is empty", nil
	}

	// Join all lines with newlines
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}

	return result, nil
}

// formatStopLine formats the stop line for display purposes
func formatStopLine(stop *int) string {
	if stop == nil {
		return "end"
	}
	return fmt.Sprintf("%d", *stop)
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
