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
	return "Reads one or more files from the collected artifacts, optionally specifying line ranges. " +
		"Use 'path' for a single file or 'files' array for multiple files in one call. " +
		"Sensitive information is sanitized by default for security."
}

func (t *readFileTool) Schema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"path": {
				Type:        genai.TypeString,
				Description: "Path to read a single file (convenience shorthand). Cannot be used together with 'files'.",
			},
			"start": {
				Type:        genai.TypeInteger,
				Description: "Starting line number for single-file mode (1-based, optional).",
			},
			"stop": {
				Type:        genai.TypeInteger,
				Description: "Ending line number for single-file mode (1-based, optional).",
			},
			"sanitize": {
				Type:        genai.TypeBoolean,
				Description: "Whether to sanitize sensitive information (default: true).",
			},
			"files": {
				Type:        genai.TypeArray,
				Description: "Array of file specifications for reading multiple files in one call. Each element must have 'path' and optionally 'start', 'stop'.",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"path": {
							Type:        genai.TypeString,
							Description: "Path to the file to read (must be from collected artifacts)",
						},
						"start": {
							Type:        genai.TypeInteger,
							Description: "Starting line number (1-based, optional)",
						},
						"stop": {
							Type:        genai.TypeInteger,
							Description: "Ending line number (1-based, optional)",
						},
					},
					Required: []string{"path"},
				},
			},
		},
	}
}

func (t *readFileTool) Execute(ctx context.Context, params map[string]any, logArtifacts []aggregator.LogEntry) (any, error) {
	if logArtifacts == nil {
		return nil, fmt.Errorf("no log artifacts provided to tool")
	}

	// Normalize input: convert single-file "path" param into a "files" array
	filesArray, err := normalizeParams(params)
	if err != nil {
		return nil, err
	}

	// Extract sanitize flag (applies to all files)
	shouldSanitize := extractBool(params, "sanitize", true)

	// Validate all files upfront before processing any
	if err := validateAllFiles(filesArray, logArtifacts); err != nil {
		return nil, err
	}

	// Process all files
	return t.processFiles(filesArray, shouldSanitize)
}

// normalizeParams converts single-file "path" parameter into a unified "files" array.
func normalizeParams(params map[string]any) ([]any, error) {
	_, hasPath := params["path"]
	filesArg, hasFiles := params["files"]

	if hasPath && hasFiles {
		return nil, fmt.Errorf("cannot use both 'path' and 'files' parameters; use 'path' for single file or 'files' for multiple")
	}

	if hasFiles {
		filesArray, ok := filesArg.([]any)
		if !ok {
			return nil, fmt.Errorf("'files' parameter must be an array")
		}
		if len(filesArray) == 0 {
			return nil, fmt.Errorf("'files' array must not be empty")
		}
		return filesArray, nil
	}

	if hasPath {
		// Convert single-file params to a files array entry
		fileSpec := map[string]any{"path": params["path"]}
		if start, ok := params["start"]; ok {
			fileSpec["start"] = start
		}
		if stop, ok := params["stop"]; ok {
			fileSpec["stop"] = stop
		}
		return []any{fileSpec}, nil
	}

	return nil, fmt.Errorf("must provide either 'path' for single file or 'files' for multiple files")
}

// validateAllFiles performs upfront validation of all file paths and line ranges.
func validateAllFiles(filesArray []any, logArtifacts []aggregator.LogEntry) error {
	for i, item := range filesArray {
		fileMap, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("files[%d]: each file specification must be an object", i)
		}

		path, err := extractString(fileMap, "path")
		if err != nil {
			return fmt.Errorf("files[%d]: %w", i, err)
		}

		if !isValidLogFile(path, logArtifacts) {
			return fmt.Errorf("files[%d]: file path %s is not in the collected artifacts", i, path)
		}

		start := extractIntPtr(fileMap, "start")
		stop := extractIntPtr(fileMap, "stop")

		if start != nil && *start < 1 {
			return fmt.Errorf("files[%d]: start line must be >= 1, got %d", i, *start)
		}
		if stop != nil && *stop < 1 {
			return fmt.Errorf("files[%d]: stop line must be >= 1, got %d", i, *stop)
		}
		if start != nil && stop != nil && *start > *stop {
			return fmt.Errorf("files[%d]: start line (%d) cannot be greater than stop line (%d)", i, *start, *stop)
		}
	}
	return nil
}

// processFiles reads all files and returns results.
// Single file: returns content directly as string.
// Multiple files: returns map[string]any with path -> content.
func (t *readFileTool) processFiles(filesArray []any, shouldSanitize bool) (any, error) {
	if len(filesArray) == 1 {
		fileMap := filesArray[0].(map[string]any)
		return t.processSingleFile(fileMap, shouldSanitize)
	}

	results := make(map[string]any, len(filesArray))
	for _, item := range filesArray {
		fileMap := item.(map[string]any)
		path, _ := extractString(fileMap, "path")

		content, err := t.processSingleFile(fileMap, shouldSanitize)
		if err != nil {
			results[path] = fmt.Sprintf("error: %v", err)
			continue
		}
		results[path] = content
	}
	return results, nil
}

// processSingleFile reads a single file based on its specification map.
func (t *readFileTool) processSingleFile(fileMap map[string]any, shouldSanitize bool) (any, error) {
	path, _ := extractString(fileMap, "path")
	start := extractIntPtr(fileMap, "start")
	stop := extractIntPtr(fileMap, "stop")

	if !shouldSanitize {
		fmt.Printf("⚠️  WARNING: Sanitization disabled for file %s - sensitive information may be exposed\n", path)
	}

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
