package reporter

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Send to Slack using common package
	if err := s.client.SendWebhook(ctx, webhookURL, message); err != nil {
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
