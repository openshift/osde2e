package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/osde2e/pkg/common/util"
)

// ArtifactLink represents a single uploaded artifact with its presigned URL.
type ArtifactLink struct {
	Name string
	URL  string
	Size int64
}

const (
	maxWorkflowFieldLength = 30000

	fullOutputThreshold = 250
	finalSummaryLines   = 80

	maxFailureBlocks    = 3
	failureContextLines = 5
	failureBlockLines   = 30
)

// SlackReporter implements Reporter interface for Slack webhook notifications
type SlackReporter struct {
	client *Client
}

// NewSlackReporter creates a new Slack reporter
func NewSlackReporter() *SlackReporter {
	return &SlackReporter{
		client: NewClient(),
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

	payload := s.buildWorkflowPayload(result, config)

	if err := s.client.SendWebhook(ctx, webhookURL, payload); err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}

	return nil
}

// WorkflowPayload represents the Slack workflow webhook payload
type WorkflowPayload struct {
	Channel        string `json:"channel"`
	Summary        string `json:"summary"`
	Analysis       string `json:"analysis"`
	ExtendedLogs   string `json:"extended_logs,omitempty"`
	ClusterDetails string `json:"cluster_details,omitempty"`
	Image          string `json:"image,omitempty"`
	Env            string `json:"env,omitempty"`
	Commit         string `json:"commit,omitempty"`
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

func (s *SlackReporter) buildWorkflowPayload(result *AnalysisResult, config *ReporterConfig) *WorkflowPayload {
	payload := &WorkflowPayload{}

	if channel, ok := config.Settings["channel"].(string); ok && channel != "" {
		payload.Channel = channel
	}

	payload.Summary = s.buildSummaryField(config)
	payload.Analysis = s.buildAnalysisField(result)

	if links, ok := config.Settings["artifact_links"].([]ArtifactLink); ok && len(links) > 0 {
		payload.ExtendedLogs = s.enforceFieldLimit(s.buildArtifactLinksSection(links), maxWorkflowFieldLength)
	} else if reportDir, ok := config.Settings["report_dir"].(string); ok && reportDir != "" {
		if testOutput := s.readTestOutput(reportDir); testOutput != "" {
			payload.ExtendedLogs = s.enforceFieldLimit(testOutput, maxWorkflowFieldLength)
		} else {
			payload.ExtendedLogs = "No test failure logs found in the report directory."
		}
	} else {
		payload.ExtendedLogs = "Test output logs not available (no report directory configured)."
	}

	if clusterDetails := s.buildClusterInfoSection(config); clusterDetails != "" {
		payload.ClusterDetails = clusterDetails
	} else {
		payload.ClusterDetails = "Cluster information not available."
	}

	if image, ok := config.Settings["image"].(string); ok && image != "" {
		payload.Image = image
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

func (s *SlackReporter) buildSummaryField(config *ReporterConfig) string {
	var builder strings.Builder
	builder.WriteString(":failed: Pipeline Failed at E2E Test\n\n")
	builder.WriteString(s.buildTestSuiteSection(config))
	return s.enforceFieldLimit(builder.String(), maxWorkflowFieldLength)
}

func (s *SlackReporter) buildAnalysisField(result *AnalysisResult) string {
	var builder strings.Builder

	if formattedAnalysis := s.formatAnalysisContent(result.Content); formattedAnalysis != "" {
		builder.WriteString(formattedAnalysis)
	} else if result.Content != "" {
		builder.WriteString(result.Content)
	}

	if result.Error != "" {
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString("====== ⚠️ Error ======\n")
		builder.WriteString(result.Error)
	}

	return s.enforceFieldLimit(builder.String(), maxWorkflowFieldLength)
}

func (s *SlackReporter) buildClusterInfoSection(config *ReporterConfig) string {
	clusterInfo, ok := config.Settings["cluster_info"].(*ClusterInfo)
	if !ok || clusterInfo == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("====== ☸️ Cluster Information ======\n")
	builder.WriteString(fmt.Sprintf("• Cluster ID: `%s`\n", clusterInfo.ID))
	if clusterInfo.Expiration != "" {
		builder.WriteString(fmt.Sprintf("• Expiration: `%s`\n", clusterInfo.Expiration))
	}
	if clusterInfo.Version != "" {
		builder.WriteString(fmt.Sprintf("• Version: `%s`\n", clusterInfo.Version))
	}
	if clusterInfo.Provider != "" {
		builder.WriteString(fmt.Sprintf("• Provider: `%s`\n", clusterInfo.Provider))
	}
	builder.WriteString("\n")

	return builder.String()
}

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
	builder.WriteString(fmt.Sprintf("• Image: `%s`\n", image))
	if env, ok := config.Settings["env"].(string); ok && env != "" {
		builder.WriteString(fmt.Sprintf("• Environment: `%s`\n", env))
	}
	builder.WriteString("\n")

	return builder.String()
}

// buildArtifactLinksSection formats S3 artifact URLs for the Slack message.
// Presigned URLs are valid for 7 days after upload.
func (s *SlackReporter) buildArtifactLinksSection(links []ArtifactLink) string {
	var builder strings.Builder
	builder.WriteString("====== 🔗 Artifacts (links expire in 7 days) ======\n")
	for _, link := range links {
		builder.WriteString(fmt.Sprintf("▸ %s\n%s\n\n", link.Name, link.URL))
	}
	return strings.TrimRight(builder.String(), "\n")
}

func (s *SlackReporter) enforceFieldLimit(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	truncated := content[:maxLength-100]
	return truncated + "\n\n... (content truncated due to length)"
}

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

func (s *SlackReporter) readTestOutput(reportDir string) string {
	filePath := filepath.Join(reportDir, "test_output.log")
	if content, err := os.ReadFile(filepath.Clean(filePath)); err == nil {
		lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
		totalLines := len(lines)

		if totalLines <= fullOutputThreshold {
			return string(content)
		}

		failureBlocks := s.extractFailureBlocks(lines, totalLines)
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

		lastN := finalSummaryLines
		var result strings.Builder
		result.WriteString("No [FAILED] or ERROR markers found. Showing final output:\n\n")
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
	return ""
}

func (s *SlackReporter) extractFailureBlocks(lines []string, endIdx int) []string {
	var blocks []string

	for i := 0; i < endIdx && len(blocks) < maxFailureBlocks; i++ {
		line := lines[i]
		if util.ContainsFailureMarker(line) || util.ContainsErrorMarker(line) {
			var block strings.Builder

			start := i - failureContextLines
			if start < 0 {
				start = 0
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
