package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SlackReporter implements Reporter interface for Slack webhook notifications
type SlackReporter struct{}

// NewSlackReporter creates a new Slack reporter
func NewSlackReporter() *SlackReporter {
	return &SlackReporter{}
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

	// Send to Slack
	if err := s.sendToSlack(ctx, webhookURL, message); err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}

	return nil
}

// SlackMessage represents a simple Slack webhook payload
type SlackMessage struct {
	Text string `json:"text"`
}

// formatMessage creates a simple text message for Slack
func (s *SlackReporter) formatMessage(result *AnalysisResult, config *ReporterConfig) *SlackMessage {
	// Create simple text message
	statusEmoji := "✅"
	if result.Status != "completed" || result.Error != "" {
		statusEmoji = "❌"
	}

	text := fmt.Sprintf("%s OSDE2E Analysis Report\n\n", statusEmoji)

	// Try to parse and format JSON analysis
	if formattedAnalysis := s.formatAnalysisContent(result.Content); formattedAnalysis != "" {
		text += formattedAnalysis
	} else {
		text += fmt.Sprintf("Analysis:\n%s", result.Content)
	}

	if result.Error != "" {
		text += fmt.Sprintf("\n\n❌ Error: %s", result.Error)
	}

	return &SlackMessage{
		Text: text,
	}
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
		formatted.WriteString("🔍 Root Cause:\n")
		formatted.WriteString(rootCause)
		formatted.WriteString("\n\n")
	}

	// Format recommendations
	if recommendations, ok := analysis["recommendations"].([]interface{}); ok && len(recommendations) > 0 {
		formatted.WriteString("💡 Recommendations:\n")
		for i, rec := range recommendations {
			if recStr, ok := rec.(string); ok {
				formatted.WriteString(fmt.Sprintf("%d. %s\n", i+1, recStr))
			}
		}
	}

	return formatted.String()
}

// sendToSlack sends the formatted message to Slack webhook
func (s *SlackReporter) sendToSlack(ctx context.Context, webhookURL string, message *SlackMessage) error {
	// Marshal message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "osde2e-analysis/1.0")

	// Create HTTP client for this request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
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
