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
		return nil // Skip disabled reporters
	}

	webhookURL, ok := config.Settings["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook_url is required and must be a string")
	}

	// Create simple message
	message := s.formatMessage(result, config)

	// Check if we have bot token for enhanced functionality
	botToken, hasBotToken := config.Settings["bot_token"].(string)
	channel, hasChannel := config.Settings["channel"].(string)
	reportDir, hasReportDir := config.Settings["report_dir"].(string)

	// If we have bot token, use chat.postMessage with file attachments
	if hasBotToken && botToken != "" && hasChannel && channel != "" && hasReportDir && reportDir != "" {
		logFiles := s.collectLogFiles(reportDir)
		if err := s.client.PostMessageWithFiles(ctx, botToken, channel, message, logFiles); err != nil {
			// Fall back to webhook on error
			fmt.Fprintf(os.Stderr, "Warning: Failed to post message with files, falling back to webhook: %v\n", err)
			// Include stdout in fallback message since we can't attach files
			messageWithStdout := s.formatMessageWithStdout(result, config)
			if err := s.client.SendWebhook(ctx, webhookURL, messageWithStdout); err != nil {
				return fmt.Errorf("failed to send to Slack: %w", err)
			}
		}
		return nil
	}

	// Fall back to webhook with stdout included (no files)
	messageWithStdout := s.formatMessageWithStdout(result, config)
	if err := s.client.SendWebhook(ctx, webhookURL, messageWithStdout); err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}

	return nil
}

// Message represents a simple Slack message payload
type Message struct {
	Analysis string `json:"analysis"`
	Summary  string `json:"summary,omitempty"`
	Channel  string `json:"channel,omitempty"`
}

// formatMessage creates a simple text message for Slack
func (s *SlackReporter) formatMessage(result *AnalysisResult, config *ReporterConfig) *Message {
	// Create simple text message
	statusEmoji := ":failed:"
	summary := fmt.Sprintf("%s Pipeline Failed at E2E Test\n", statusEmoji)
	text := ""

	// Add cluster information to summary
	if clusterInfo, ok := config.Settings["cluster_info"].(*ClusterInfo); ok && clusterInfo != nil {
		summary += "\n====== Cluster Information ======\n"
		summary += fmt.Sprintf("Cluster ID: %s\n", clusterInfo.ID)
		if clusterInfo.Expiration != "" {
			summary += fmt.Sprintf("Expiration: %s\n", clusterInfo.Expiration)
		}
		if clusterInfo.Name != "" {
			summary += fmt.Sprintf("Name: %s\n", clusterInfo.Name)
		}
		if clusterInfo.Version != "" {
			summary += fmt.Sprintf("Version: %s\n", clusterInfo.Version)
		}
		if clusterInfo.Provider != "" {
			summary += fmt.Sprintf("Provider: %s\n", clusterInfo.Provider)
		}
		summary += "\n"
	}

	if image, ok := config.Settings["image"].(string); ok && image != "" {
		imageInfo := strings.Split(image, ":")
		image := imageInfo[0]
		commit := imageInfo[1]
		env := config.Settings["env"].(string)
		summary += fmt.Sprintf("Test suite: %s \nCommit: %s \nEnvironment: %s\n", image, commit, env)
	}

	// Try to parse and format JSON analysis
	if formattedAnalysis := s.formatAnalysisContent(result.Content); formattedAnalysis != "" {
		text += formattedAnalysis
	} else {
		text += fmt.Sprintf("Analysis:\n%s", result.Content)
	}
	if result.Error != "" {
		text += fmt.Sprintf("\n\n Error: %s", result.Error)
	}
	message := &Message{
		Summary:  summary,
		Analysis: text,
	}
	// Add channel if specified (for workflow webhooks)
	if channel, ok := config.Settings["channel"].(string); ok && channel != "" {
		message.Channel = channel
	}

	return message
}

// formatMessageWithStdout creates a message with stdout content included for webhook fallback
func (s *SlackReporter) formatMessageWithStdout(result *AnalysisResult, config *ReporterConfig) *Message {
	message := s.formatMessage(result, config)

	// Add stdout content if report directory is available
	if reportDir, ok := config.Settings["report_dir"].(string); ok && reportDir != "" {
		if stdout := s.readTestOutput(reportDir); stdout != "" {
			message.Analysis += "\n\n====== Test Pod Stdout ======\n"
			message.Analysis += stdout
		}
	}

	return message
}

// formatAnalysisContent tries to parse JSON and format it nicely for Slack
func (s *SlackReporter) formatAnalysisContent(content string) string {
	// Look for JSON content in code blocks
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

	// Parse JSON
	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent.String()), &analysis); err != nil {
		return ""
	}

	var formatted strings.Builder

	// Format root cause
	if rootCause, ok := analysis["root_cause"].(string); ok && rootCause != "" {
		formatted.WriteString("====== 🔍 Possible Cause ======\n")
		formatted.WriteString(rootCause)
		formatted.WriteString("\n\n")
	}

	// Format recommendations
	if recommendations, ok := analysis["recommendations"].([]interface{}); ok && len(recommendations) > 0 {
		formatted.WriteString("====== 💡 Recommendations ======\n")
		for i, rec := range recommendations {
			if recStr, ok := rec.(string); ok {
				formatted.WriteString(fmt.Sprintf("%d. %s\n", i+1, recStr))
			}
		}
	}

	return formatted.String()
}

// collectLogFiles collects all log and XML files from the report directory
func (s *SlackReporter) collectLogFiles(reportDir string) []string {
	var files []string

	err := filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		if info.IsDir() {
			return nil
		}

		// Collect .log, .txt, and .xml files
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".log" || ext == ".txt" || ext == ".xml" {
			files = append(files, path)
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
	// Try test_output.txt first (main pod stdout), then test_output.log
	for _, filename := range []string{"test_output.txt", "test_output.log"} {
		filePath := filepath.Join(reportDir, filename)
		if content, err := os.ReadFile(filepath.Clean(filePath)); err == nil {
			lines := strings.Split(string(content), "\n")
			totalLines := len(lines)

			// If content is small enough, return it all
			if totalLines <= 150 {
				return string(content)
			}

			// Show first 50 lines (context) and last 100 lines (failure details)
			firstN := 50
			lastN := 100

			var result strings.Builder

			// First N lines
			for i := 0; i < firstN && i < totalLines; i++ {
				result.WriteString(lines[i])
				result.WriteString("\n")
			}

			// Omission notice
			omitted := totalLines - firstN - lastN
			result.WriteString(fmt.Sprintf("\n... (%d lines omitted) ...\n\n", omitted))

			// Last N lines
			startIdx := totalLines - lastN
			if startIdx < firstN {
				startIdx = firstN
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
