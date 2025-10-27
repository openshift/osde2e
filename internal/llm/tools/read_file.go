package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/openshift/osde2e/internal/aggregator"
	"github.com/openshift/osde2e/internal/sanitizer"
	"google.golang.org/genai"
)

type readFileTool struct {
	sanitizer *sanitizer.Sanitizer
}

// newReadFileTool creates a new read file tool with sanitizer
func newReadFileTool() *readFileTool {
	// Initialize sanitizer with default config
	s, err := sanitizer.New(nil)
	if err != nil {
		// If sanitizer fails to initialize, create tool without it
		// This ensures the tool still works even if sanitizer has issues
		return &readFileTool{sanitizer: nil}
	}

	return &readFileTool{sanitizer: s}
}

func (t *readFileTool) Name() string {
	return "read_file"
}

func (t *readFileTool) Description() string {
	return "Reads a specific file from the collected artifacts, optionally specifying a line range. Sensitive information is sanitized by default for security. Use start and stop parameters to read specific line ranges."
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
			"sanitize": {
				Type:        genai.TypeBoolean,
				Description: "Whether to sanitize sensitive information from the content (default: true). Set to false only for debugging or testing purposes.",
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

	// Extract parameters with defaults
	start := extractIntPtr(params, "start")
	stop := extractIntPtr(params, "stop")
	shouldSanitize := extractBool(params, "sanitize", true)

	// Log when sanitization is disabled for security awareness
	if !shouldSanitize {
		fmt.Printf("⚠️  WARNING: Sanitization disabled for file %s - sensitive information may be exposed\n", path)
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
	content, err := t.readFileWithLineRange(path, start, stop, shouldSanitize)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return content, nil
}

// readFileWithLineRange reads a file and returns content within the specified line range
func (t *readFileTool) readFileWithLineRange(filePath string, start, stop *int, shouldSanitize bool) (string, error) {
	// Read lines from file within the specified range
	rawLines, lineNumbers, err := t.readLinesInRange(filePath, start, stop)
	if err != nil {
		return "", err
	}

	if len(rawLines) == 0 {
		if start != nil {
			startLine := 1
			if start != nil {
				startLine = *start
			}
			return fmt.Sprintf("No lines found in range %d-%s", startLine, formatStopLine(stop)), nil
		}
		return "File is empty", nil
	}

	// Process lines (with or without sanitization)
	formattedLines := t.processLines(rawLines, lineNumbers, filePath, shouldSanitize)

	// Join all lines with newlines
	return joinLines(formattedLines), nil
}

// readLinesInRange reads lines from file within the specified range
func (t *readFileTool) readLinesInRange(filePath string, start, stop *int) ([]string, []int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var rawLines []string
	var lineNumbers []int
	lineNum := 1

	startLine := 1
	if start != nil {
		startLine = *start
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip lines before start
		if lineNum < startLine {
			lineNum++
			continue
		}

		// Stop if we've reached the stop line
		if stop != nil && lineNum > *stop {
			break
		}

		rawLines = append(rawLines, line)
		lineNumbers = append(lineNumbers, lineNum)
		lineNum++
	}

	return rawLines, lineNumbers, scanner.Err()
}

// processLines applies sanitization and formatting to lines
func (t *readFileTool) processLines(rawLines []string, lineNumbers []int, filePath string, shouldSanitize bool) []string {
	if !shouldSanitize || t.sanitizer == nil {
		return t.formatLinesWithoutSanitization(rawLines, lineNumbers)
	}

	return t.sanitizeAndFormatLines(rawLines, lineNumbers, filePath)
}

// formatLinesWithoutSanitization formats lines without sanitization
func (t *readFileTool) formatLinesWithoutSanitization(rawLines []string, lineNumbers []int) []string {
	lines := make([]string, len(rawLines))
	for i, line := range rawLines {
		lines[i] = fmt.Sprintf("%d\t%s", lineNumbers[i], line)
	}
	return lines
}

// sanitizeAndFormatLines applies batch sanitization and formats lines
func (t *readFileTool) sanitizeAndFormatLines(rawLines []string, lineNumbers []int, filePath string) []string {
	const batchSize = 1000 // Fixed batch size for optimal performance
	var formattedLines []string

	// Process in batches
	for batchStart := 0; batchStart < len(rawLines); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > len(rawLines) {
			batchEnd = len(rawLines)
		}

		batchLines := rawLines[batchStart:batchEnd]
		batchLineNumbers := lineNumbers[batchStart:batchEnd]

		// Create sources for batch
		sources := make([]string, len(batchLines))
		for i := range sources {
			sources[i] = fmt.Sprintf("%s:line_%d", filePath, batchLineNumbers[i])
		}

		// Try batch sanitization first
		if results, err := t.sanitizer.SanitizeBatch(batchLines, sources); err == nil {
			// Batch sanitization succeeded
			for i, result := range results {
				formattedLines = append(formattedLines, fmt.Sprintf("%d\t%s", batchLineNumbers[i], result.Content))
			}
		} else {
			// Fallback to line-by-line sanitization
			for i, line := range batchLines {
				sanitizedLine := t.sanitizeSingleLine(line, sources[i])
				formattedLines = append(formattedLines, fmt.Sprintf("%d\t%s", batchLineNumbers[i], sanitizedLine))
			}
		}
	}

	return formattedLines
}

// sanitizeSingleLine sanitizes a single line with error handling
func (t *readFileTool) sanitizeSingleLine(line, source string) string {
	result, err := t.sanitizer.SanitizeText(line, source)
	if err != nil {
		return fmt.Sprintf("%s [SANITIZATION-ERROR: %v]", line, err)
	}
	return result.Content
}

// joinLines joins formatted lines with newlines
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}
	return result
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

// extractIntPtr extracts an optional integer parameter and returns a pointer
func extractIntPtr(params map[string]any, key string) *int {
	val, exists := params[key]
	if !exists {
		return nil
	}

	switch v := val.(type) {
	case float64:
		result := int(v)
		return &result
	case int:
		return &v
	case int64:
		result := int(v)
		return &result
	default:
		return nil
	}
}

// extractBool extracts a boolean parameter with a default value
func extractBool(params map[string]any, key string, defaultValue bool) bool {
	val, exists := params[key]
	if !exists {
		return defaultValue
	}

	if boolVal, ok := val.(bool); ok {
		return boolVal
	}

	return defaultValue
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
