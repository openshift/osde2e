package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	commonslack "github.com/openshift/osde2e/pkg/common/slack"
)

const (
	// Slack webhook message limits
	maxWebhookLength    = 2900
	minReservedSpace    = 500
	truncationNoticeLen = 150
	minStdoutSpace      = 200

	// Test output truncation thresholds
	fullOutputThreshold = 250
	initialContextLines = 20
	finalSummaryLines   = 80

	// Failure block extraction
	maxFailureBlocks    = 3
	failureContextLines = 5
	failureBlockLines   = 30

	// File collection defaults
	defaultMaxLogFiles  = 15
	defaultMaxLogSizeMB = 50
)

// SlackReporter implements Reporter interface for Slack webhook notifications
type SlackReporter struct {
	client *commonslack.Client
}

// NewSlackReporter creates a new Slack reporter
func NewSlackReporter() *SlackReporter {
	return &SlackReporter{
		client: commonslack.NewClient(),
	}
}

// Name returns the reporter identifier
func (s *SlackReporter) Name() string {
	return "slack"
}

// Report sends the analysis result to Slack via webhook
func (s *SlackReporter) Report(ctx context.Context, result *AnalysisResult, config *ReporterConfig) error {
	if !config.Enabled {
		return nil
	}

	webhookURL, ok := config.Settings["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook_url is required and must be a string")
	}

	message := s.formatMessage(result, config)

	botToken, hasBotToken := config.Settings["bot_token"].(string)
	channel, hasChannel := config.Settings["channel"].(string)
	reportDir, hasReportDir := config.Settings["report_dir"].(string)

	if hasBotToken && botToken != "" && hasChannel && channel != "" && hasReportDir && reportDir != "" {
		logFiles := s.collectLogFilesWithConfig(reportDir, config)
		if err := s.client.PostMessageWithFiles(ctx, botToken, channel, message, logFiles); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to post message with files, falling back to webhook: %v\n", err)
			// Include stdout in fallback message since we can't attach files
			messageWithStdout := s.formatMessageWithStdout(result, config)
			if err := s.client.SendWebhook(ctx, webhookURL, messageWithStdout); err != nil {
				return fmt.Errorf("failed to send to Slack: %w", err)
			}
		}
		return nil
	}

	messageWithStdout := s.formatMessageWithStdout(result, config)
	if err := s.client.SendWebhook(ctx, webhookURL, messageWithStdout); err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}

	return nil
}

// Message represents a simple Slack message payload
type Message struct {
	Text    string `json:"text"`              // Combined text for webhooks
	Channel string `json:"channel,omitempty"` // Optional channel override

	// Internal fields for PostMessageWithFiles (not sent via JSON)
	summary  string // Used by PostMessageWithFiles
	analysis string // Used by PostMessageWithFiles
}

// formatMessage creates a simple text message for Slack
func (s *SlackReporter) formatMessage(result *AnalysisResult, config *ReporterConfig) *Message {
	// Build message components in priority order
	// Priority: Header > Cluster Info > Analysis > Error > Test Suite Info

	statusEmoji := ":failed:"
	header := fmt.Sprintf("%s Pipeline Failed at E2E Test\n", statusEmoji)

	analysis := s.buildAnalysisSection(result)
	clusterInfo := s.buildClusterInfoSection(config)

	errorMsg := ""
	if result.Error != "" {
		errorMsg = fmt.Sprintf("\n\nError: %s", result.Error)
	}

	testSuiteInfo := s.buildTestSuiteSection(config)

	fullSummary := header + clusterInfo + testSuiteInfo
	fullAnalysis := analysis + errorMsg

	_, hasBotToken := config.Settings["bot_token"].(string)

	truncatedText := s.buildTruncatedMessage(
		header,
		clusterInfo,
		analysis,
		errorMsg,
		testSuiteInfo,
		maxWebhookLength,
		hasBotToken,
	)

	message := &Message{
		Text:     truncatedText,
		summary:  fullSummary,
		analysis: fullAnalysis,
	}

	if channel, ok := config.Settings["channel"].(string); ok && channel != "" {
		message.Channel = channel
	}

	return message
}

// buildClusterInfoSection creates the full cluster information section
func (s *SlackReporter) buildClusterInfoSection(config *ReporterConfig) string {
	clusterInfo, ok := config.Settings["cluster_info"].(*ClusterInfo)
	if !ok || clusterInfo == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("\n# Cluster Info\n")
	builder.WriteString(fmt.Sprintf("Cluster ID: %s\n", clusterInfo.ID))
	if clusterInfo.Name != "" {
		builder.WriteString(fmt.Sprintf("Name: %s\n", clusterInfo.Name))
	}
	if clusterInfo.Version != "" {
		builder.WriteString(fmt.Sprintf("Version: %s\n", clusterInfo.Version))
	}
	if clusterInfo.Provider != "" {
		builder.WriteString(fmt.Sprintf("Provider: %s\n", clusterInfo.Provider))
	}
	if clusterInfo.Expiration != "" {
		builder.WriteString(fmt.Sprintf("Expiration: %s\n", clusterInfo.Expiration))
	}
	builder.WriteString("\n")

	return builder.String()
}

// buildTestSuiteSection creates the test suite information section
func (s *SlackReporter) buildTestSuiteSection(config *ReporterConfig) string {
	image, ok := config.Settings["image"].(string)
	if !ok || image == "" {
		return ""
	}

	imageInfo := strings.Split(image, ":")
	if len(imageInfo) < 2 {
		return ""
	}

	env, _ := config.Settings["env"].(string)
	return fmt.Sprintf("Test suite: %s \nCommit: %s \nEnvironment: %s\n\n", imageInfo[0], imageInfo[1], env)
}

// buildAnalysisSection creates the analysis content section
func (s *SlackReporter) buildAnalysisSection(result *AnalysisResult) string {
	if formattedAnalysis := s.formatAnalysisContent(result.Content); formattedAnalysis != "" {
		return formattedAnalysis
	}
	return fmt.Sprintf("Analysis:\n%s", result.Content)
}

// buildTruncatedMessage strategically truncates message components to fit Slack's limit.
// Priority: Header > Cluster Info > Analysis (truncate if needed) > Error > Test Suite
func (s *SlackReporter) buildTruncatedMessage(header, clusterInfo, analysis, errorMsg, testSuiteInfo string, maxLength int, hasBotToken bool) string {
	// Always include header
	baseContent := header + clusterInfo

	// Try to fit everything
	fullMessage := baseContent + analysis + errorMsg + testSuiteInfo
	if len(fullMessage) <= maxLength {
		return fullMessage
	}

	// Drop test suite info
	withoutTestSuite := baseContent + analysis + errorMsg
	if len(withoutTestSuite) <= maxLength {
		suffix := s.buildTruncationNotice("test suite info", hasBotToken)
		return withoutTestSuite + suffix
	}

	// Drop error message
	withoutError := baseContent + analysis
	if len(withoutError) <= maxLength {
		suffix := s.buildTruncationNotice("error message and test suite info", hasBotToken)
		return withoutError + suffix
	}

	// Truncate analysis to fit
	availableSpace := maxLength - len(baseContent) - truncationNoticeLen
	if availableSpace > 100 {
		truncatedAnalysis := analysis[:availableSpace]
		suffix := s.buildTruncationNotice("partial analysis, error message, and test suite info", hasBotToken)
		return baseContent + truncatedAnalysis + suffix
	}

	// Last resort: just header and cluster info
	if len(baseContent) <= maxLength {
		suffix := s.buildTruncationNotice("analysis, error message, and test suite info", hasBotToken)
		return baseContent + suffix
	}

	// Cluster info alone is too big (unlikely)
	return header + "\n\n[WARNING: Cluster information too large for Slack message limits]"
}

// buildTruncationNotice creates an appropriate truncation notice
func (s *SlackReporter) buildTruncationNotice(omittedContent string, hasBotToken bool) string {
	if hasBotToken {
		return fmt.Sprintf("\n\n... (%s omitted due to length, see attached files for full details)", omittedContent)
	}
	return fmt.Sprintf("\n\n... (%s omitted due to Slack message length limit)", omittedContent)
}

// formatMessageWithStdout creates a message with stdout content included for webhook fallback.
// Note: Stdout has lowest priority and will be omitted from webhook if space is limited.
// The readTestOutput already extracts failure blocks, so we're prioritizing the most important content.
func (s *SlackReporter) formatMessageWithStdout(result *AnalysisResult, config *ReporterConfig) *Message {
	message := s.formatMessage(result, config)

	_, hasBotToken := config.Settings["bot_token"].(string)

	if reportDir, ok := config.Settings["report_dir"].(string); ok && reportDir != "" {
		if stdout := s.readTestOutput(reportDir); stdout != "" {
			// readTestOutput extracts only failure blocks - the most important content
			stdoutSection := "\n\n# Test Failures\n" + stdout

			message.analysis += stdoutSection

			currentLength := len(message.Text)
			if currentLength+len(stdoutSection) <= maxWebhookLength {
				message.Text += stdoutSection
			} else if currentLength < maxWebhookLength-minReservedSpace {
				availableSpace := maxWebhookLength - currentLength - truncationNoticeLen
				if availableSpace > minStdoutSpace {
					// Truncate from beginning since readTestOutput now only returns failure blocks
					// Show the start of the first failure block
					endPos := availableSpace
					if endPos > len(stdout) {
						endPos = len(stdout)
					}
					truncatedStdout := stdout[:endPos]
					fileMsg := ""
					if hasBotToken {
						fileMsg = " (see attached files for full output)"
					}
					message.Text += "\n\n# Test Failures\n" + truncatedStdout + "..." + fileMsg
				} else {
					if hasBotToken {
						message.Text += "\n\n... (test failures omitted, see attached files)"
					} else {
						message.Text += "\n\n... (test failures omitted due to length)"
					}
				}
			}
		}
	}

	return message
}

// formatAnalysisContent tries to parse JSON and format it nicely for Slack
func (s *SlackReporter) formatAnalysisContent(content string) string {
	lines := strings.Split(content, "\n")
	var jsonContent strings.Builder
	inJSONBlock := false

	for _, line := range lines {
		if strings.Contains(line, "```json") {
			inJSONBlock = true
			continue
		}
		if strings.Contains(line, "```") && inJSONBlock {
			break
		}
		if inJSONBlock {
			jsonContent.WriteString(line + "\n")
		}
	}

	if jsonContent.Len() == 0 {
		return ""
	}

	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent.String()), &analysis); err != nil {
		return ""
	}

	var formatted strings.Builder

	if rootCause, ok := analysis["root_cause"].(string); ok && rootCause != "" {
		formatted.WriteString("# Possible Cause\n")
		formatted.WriteString(rootCause)
		formatted.WriteString("\n\n")
	}

	if recommendations, ok := analysis["recommendations"].([]interface{}); ok && len(recommendations) > 0 {
		formatted.WriteString("# Recommendations\n")
		for i, rec := range recommendations {
			if recStr, ok := rec.(string); ok {
				formatted.WriteString(fmt.Sprintf("%d. %s\n", i+1, recStr))
			}
		}
	}

	return formatted.String()
}

// collectLogFiles collects log files from the report directory with filtering.
// File collection is limited by:
// - Allowlist patterns (configurable via "log_file_patterns" setting)
// - Maximum file count (configurable via "max_log_files" setting, default: 15)
// - Maximum total size (configurable via "max_log_size_mb" setting, default: 50MB)
func (s *SlackReporter) collectLogFiles(reportDir string) []string {
	return s.collectLogFilesWithConfig(reportDir, nil)
}

// collectLogFilesWithConfig collects log files with optional config override (for testing)
func (s *SlackReporter) collectLogFilesWithConfig(reportDir string, config *ReporterConfig) []string {
	defaultPatterns := []string{
		"test_output.log",
		"test_output.txt",
		"junit*.xml",
	}

	maxFiles := defaultMaxLogFiles
	maxSizeMB := defaultMaxLogSizeMB

	var patterns []string
	if config != nil {
		if configPatterns, ok := config.Settings["log_file_patterns"].([]string); ok && len(configPatterns) > 0 {
			patterns = configPatterns
		}
		if configMaxFiles, ok := config.Settings["max_log_files"].(int); ok && configMaxFiles > 0 {
			maxFiles = configMaxFiles
		}
		if configMaxSizeMB, ok := config.Settings["max_log_size_mb"].(int); ok && configMaxSizeMB > 0 {
			maxSizeMB = configMaxSizeMB
		}
	}

	if len(patterns) == 0 {
		patterns = defaultPatterns
	}

	var files []string
	var totalSize int64
	maxSizeBytes := int64(maxSizeMB * 1024 * 1024)

	err := filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		if info.IsDir() {
			return nil
		}

		if len(files) >= maxFiles {
			return filepath.SkipAll
		}

		if totalSize+info.Size() > maxSizeBytes {
			fmt.Fprintf(os.Stderr, "Warning: Skipping remaining files, size limit reached (%dMB)\n", maxSizeMB)
			return filepath.SkipAll
		}

		filename := filepath.Base(path)
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Invalid pattern %q: %v\n", pattern, err)
				continue
			}
			if matched {
				files = append(files, path)
				totalSize += info.Size()
				break
			}
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error collecting log files: %v\n", err)
	}

	return files
}

// readTestOutput reads the test stdout from test_output.txt or test_output.log
func (s *SlackReporter) readTestOutput(reportDir string) string {
	for _, filename := range []string{"test_output.txt", "test_output.log"} {
		filePath := filepath.Join(reportDir, filename)
		if content, err := os.ReadFile(filepath.Clean(filePath)); err == nil {
			lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
			totalLines := len(lines)

			if totalLines <= fullOutputThreshold {
				return string(content)
			}

			// For large logs, extract only failure blocks - this is what matters
			failureBlocks := s.extractFailureBlocks(lines, 0, totalLines)
			if len(failureBlocks) > 0 {
				var result strings.Builder
				result.WriteString(fmt.Sprintf("Found %d test failure(s):\n\n", len(failureBlocks)))
				for i, block := range failureBlocks {
					if i > 0 {
						result.WriteString("\n---\n\n")
					}
					result.WriteString(block)
				}
				return result.String()
			}

			// No failures found, return summary section
			lastN := finalSummaryLines
			var result strings.Builder
			result.WriteString("No [FAILED] markers found. Showing final output:\n\n")
			startIdx := totalLines - lastN
			if startIdx < 0 {
				startIdx = 0
			}
			for i := startIdx; i < totalLines; i++ {
				result.WriteString(lines[i])
				result.WriteString("\n")
			}
			return result.String()
		}
	}
	return ""
}

// extractFailureBlocks finds [FAILED] test blocks and extracts them with context
func (s *SlackReporter) extractFailureBlocks(lines []string, startIdx, endIdx int) []string {
	var blocks []string

	for i := startIdx; i < endIdx && len(blocks) < maxFailureBlocks; i++ {
		if strings.Contains(lines[i], "[FAILED]") || strings.Contains(lines[i], "• [FAILED]") {
			var block strings.Builder

			start := i - failureContextLines
			if start < startIdx {
				start = startIdx
			}

			end := i + failureBlockLines
			if end > endIdx {
				end = endIdx
			}

			for j := start; j < end; j++ {
				block.WriteString(lines[j])
				if j < end-1 {
					block.WriteString("\n")
				}
			}

			blocks = append(blocks, block.String())

			// Skip ahead to avoid overlapping blocks
			i = end - 1
		}
	}

	return blocks
}

// ClusterInfo holds cluster information for reporting (mirrored from analysisengine to avoid import cycle)
type ClusterInfo struct {
	ID            string
	Name          string
	Provider      string
	Region        string
	CloudProvider string
	Version       string
	Expiration    string
}

// SlackReporterConfig creates a reporter config for Slack
func SlackReporterConfig(webhookURL string, enabled bool) ReporterConfig {
	return ReporterConfig{
		Type:    "slack",
		Enabled: enabled,
		Settings: map[string]interface{}{
			"webhook_url": webhookURL,
		},
	}
}

// BuildNotificationConfig creates notification configuration for log analysis.
func BuildNotificationConfig(webhook string, channel string, clusterInfo interface{}, reportDir string, botToken string) *NotificationConfig {
	if webhook == "" || channel == "" {
		return nil
	}

	slackConfig := SlackReporterConfig(webhook, true)
	slackConfig.Settings["channel"] = channel
	slackConfig.Settings["cluster_info"] = clusterInfo
	slackConfig.Settings["report_dir"] = reportDir
	if botToken != "" {
		slackConfig.Settings["bot_token"] = botToken
	}

	return &NotificationConfig{
		Enabled:   true,
		Reporters: []ReporterConfig{slackConfig},
	}
}
