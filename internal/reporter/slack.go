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
	// Slack workflow payload limits (conservative estimate)
	// Slack workflows can handle much larger payloads than webhooks
	maxWorkflowFieldLength = 30000 // 30KB per field

	// Test output truncation thresholds
	fullOutputThreshold = 250
	initialContextLines = 20
	finalSummaryLines   = 80

	// Failure block extraction
	maxFailureBlocks    = 3
	failureContextLines = 5
	failureBlockLines   = 30
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

	// Build workflow payload
	payload := s.buildWorkflowPayload(result, config)

	// Send to Slack workflow webhook
	if err := s.client.SendWebhook(ctx, webhookURL, payload); err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}

	return nil
}

// WorkflowPayload represents the Slack workflow webhook payload
type WorkflowPayload struct {
	Channel        string `json:"channel"`                   // Required - Slack channel ID
	Summary        string `json:"summary"`                   // Required - Initial message (test suite info)
	Analysis       string `json:"analysis"`                  // Required - AI analysis (posted as reply 1)
	ExtendedLogs   string `json:"extended_logs,omitempty"`   // Optional - Test failures (posted as reply 2)
	ClusterDetails string `json:"cluster_details,omitempty"` // Optional - Cluster info for debugging (posted as reply 3)
	Image          string `json:"image,omitempty"`           // Optional - Test image
	Env            string `json:"env,omitempty"`             // Optional - Environment
	Commit         string `json:"commit,omitempty"`          // Optional - Commit hash
}

// ClusterInfo holds cluster information for reporting
type ClusterInfo struct {
	ID            string
	Name          string
	Provider      string
	Region        string
	CloudProvider string
	Version       string
	Expiration    string
}

// buildWorkflowPayload constructs the JSON payload for the Slack Workflow
func (s *SlackReporter) buildWorkflowPayload(result *AnalysisResult, config *ReporterConfig) *WorkflowPayload {
	payload := &WorkflowPayload{}

	// Required: channel ID
	if channel, ok := config.Settings["channel"].(string); ok && channel != "" {
		payload.Channel = channel
	}

	// Required: summary (initial message)
	payload.Summary = s.buildSummaryField(config)

	// Required: analysis (AI response)
	payload.Analysis = s.buildAnalysisField(result)

	// Optional: extended_logs (test failures)
	if reportDir, ok := config.Settings["report_dir"].(string); ok && reportDir != "" {
		if testOutput := s.readTestOutput(reportDir); testOutput != "" {
			payload.ExtendedLogs = s.enforceFieldLimit(testOutput, maxWorkflowFieldLength)
		} else {
			// Provide fallback when logs exist but couldn't be read
			payload.ExtendedLogs = "No test failure logs found in the report directory."
		}
	} else {
		// Provide fallback when no report directory is configured
		payload.ExtendedLogs = "Test output logs not available (no report directory configured)."
	}

	// Optional: cluster_details (for debugging)
	if clusterDetails := s.buildClusterInfoSection(config); clusterDetails != "" {
		payload.ClusterDetails = clusterDetails
	} else {
		// Provide fallback when no cluster info is configured
		payload.ClusterDetails = "Cluster information not available."
	}

	// Optional metadata
	if image, ok := config.Settings["image"].(string); ok && image != "" {
		payload.Image = image
		// Extract commit from image tag if present
		parts := strings.Split(image, ":")
		if len(parts) == 2 {
			payload.Commit = parts[1]
		}
	}

	if env, ok := config.Settings["env"].(string); ok && env != "" {
		payload.Env = env
	}

	return payload
}

// buildSummaryField creates the initial message content
func (s *SlackReporter) buildSummaryField(config *ReporterConfig) string {
	var builder strings.Builder

	// Header
	builder.WriteString(":failed: Pipeline Failed at E2E Test\n\n")

	// Test suite info (what failed)
	builder.WriteString(s.buildTestSuiteSection(config))

	return s.enforceFieldLimit(builder.String(), maxWorkflowFieldLength)
}

// buildAnalysisField formats the AI analysis
func (s *SlackReporter) buildAnalysisField(result *AnalysisResult) string {
	var builder strings.Builder

	// Format the analysis content (handles JSON parsing)
	if formattedAnalysis := s.formatAnalysisContent(result.Content); formattedAnalysis != "" {
		builder.WriteString(formattedAnalysis)
	} else if result.Content != "" {
		builder.WriteString(result.Content)
	}

	// Add error if present
	if result.Error != "" {
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString("====== ⚠️ Error ======\n")
		builder.WriteString(result.Error)
	}

	return s.enforceFieldLimit(builder.String(), maxWorkflowFieldLength)
}

// buildClusterInfoSection creates the cluster information section
func (s *SlackReporter) buildClusterInfoSection(config *ReporterConfig) string {
	clusterInfo, ok := config.Settings["cluster_info"].(*ClusterInfo)
	if !ok || clusterInfo == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("====== ☸️ Cluster Information ======\n")
	builder.WriteString(fmt.Sprintf("• Cluster ID: `%s`\n", clusterInfo.ID))
	if clusterInfo.Name != "" {
		builder.WriteString(fmt.Sprintf("• Name: `%s`\n", clusterInfo.Name))
	}
	if clusterInfo.Version != "" {
		builder.WriteString(fmt.Sprintf("• Version: `%s`\n", clusterInfo.Version))
	}
	if clusterInfo.Provider != "" {
		builder.WriteString(fmt.Sprintf("• Provider: `%s`\n", clusterInfo.Provider))
	}
	if clusterInfo.Expiration != "" {
		builder.WriteString(fmt.Sprintf("• Expiration: `%s`\n", clusterInfo.Expiration))
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

	var builder strings.Builder
	builder.WriteString("====== 🧪 Test Suite Information ======\n")
	builder.WriteString(fmt.Sprintf("• Image: `%s`\n", imageInfo[0]))
	builder.WriteString(fmt.Sprintf("• Commit: `%s`\n", imageInfo[1]))
	if env, ok := config.Settings["env"].(string); ok && env != "" {
		builder.WriteString(fmt.Sprintf("• Environment: `%s`\n", env))
	}
	builder.WriteString("\n")

	return builder.String()
}

// enforceFieldLimit truncates a field to the maximum allowed length
func (s *SlackReporter) enforceFieldLimit(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	// Truncate and add notice
	truncated := content[:maxLength-100]
	return truncated + "\n\n... (content truncated due to length)"
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
		formatted.WriteString("====== 🔍 Possible Cause ======\n")
		formatted.WriteString(rootCause)
		formatted.WriteString("\n\n")
	}

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

// readTestOutput reads the test stdout from test_output.txt, test_output.log, or build-log.txt
func (s *SlackReporter) readTestOutput(reportDir string) string {
	for _, filename := range []string{"test_output.txt", "test_output.log", "build-log.txt"} {
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
				result.WriteString("======  Log Extract ======\n")
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
func BuildNotificationConfig(webhook string, channel string, clusterInfo interface{}, reportDir string) *NotificationConfig {
	if webhook == "" || channel == "" {
		return nil
	}

	slackConfig := SlackReporterConfig(webhook, true)
	slackConfig.Settings["channel"] = channel
	slackConfig.Settings["cluster_info"] = clusterInfo
	slackConfig.Settings["report_dir"] = reportDir

	return &NotificationConfig{
		Enabled:   true,
		Reporters: []ReporterConfig{slackConfig},
	}
}
