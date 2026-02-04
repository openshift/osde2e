# Reporter System (Developer Documentation)

The reporter system handles notification delivery after LLM analysis completion. This document covers the internal architecture and implementation details for developers working on the reporter system.

**For user setup instructions, see the [root README](../../README.md#slack-notifications).**

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│ Analysis Engine │    │ NotificationConfig│    │ ReporterRegistry    │
│ completes LLM   ├───▶│ • Enabled: bool   ├───▶│ • Manages reporters │
│ analysis        │    │ • Reporters: []   │    │ • Get by type       │
└─────────────────┘    └──────────────────┘    └─────────┬───────────┘
                                                         │
                       ┌─────────────────────────────────▼───────────────┐
                       │ For each ReporterConfig in Reporters array:     │
                       │ 1. Check if config.Enabled = true               │
                       │ 2. Get reporter by config.Type from registry    │
                       │ 3. Call reporter.Report(ctx, result, config)    │
                       └─────────────────┬───────────────────────────────┘
                                         │
                       ┌─────────────────▼───────────────┐
                       │ SlackReporter Implementation    │
                       │ ┌─────────────────────────────┐ │
                       │ │ Report() method:            │ │
                       │ │ 1. Extract webhook_url      │ │
                       │ │ 2. Format analysis message  │ │
                       │ │ 3. Send to Slack via HTTP   │ │
                       │ └─────────────────────────────┘ │
                       └─────────────────────────────────┘
                                         │
                       ┌─────────────────▼───────────────┐
                       │ sendToSlack() method:           │
                       │ ┌─────────────────────────────┐ │
                       │ │ 1. Marshal message to JSON  │ │
                       │ │ 2. Create HTTP POST request │ │
                       │ │ 3. Create HTTP client       │ │
                       │ │ 4. Send request to webhook  │ │
                       │ │ 5. Validate response status │ │
                       │ └─────────────────────────────┘ │
                       └─────────────────────────────────┘
                                         │
                       ┌─────────────────▼───────────────┐
                       │ Slack Webhook Endpoint          │
                       │ • Receives formatted message    │
                       │ • Posts to configured channel   │
                       │ • Returns HTTP 200 on success   │
                       └─────────────────────────────────┘
```

## Configuration Flow

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│ User Config     │    │ Application      │    │ NotificationConfig  │
│ • Slack webhook ├───▶│ Creates config   ├───▶│ Created when needed │
│ • Enable flags  │    │ conditionally    │    │ (not by default)    │
└─────────────────┘    └──────────────────┘    └─────────────────────┘
```

**E2E usage** (`pkg/e2e/e2e.go`):
```go
if enableSlackNotify && slackWebhook != "" {
    reporters = append(reporters, reporter.SlackReporterConfig(slackWebhook, true))
}
if len(reporters) > 0 {
    notificationConfig = &reporter.NotificationConfig{
        Enabled: true,
        Reporters: reporters,
    }
}
```

## Slack Workflow Integration

The Slack reporter sends test failure notifications using a **Slack Workflow** that creates threaded messages. This allows teams to add the shared workflow to their channels and receive structured failure notifications.

### How It Works

The workflow creates four messages in a thread:

1. **Initial Message** - Test suite information (what failed)
2. **First Reply** - AI-powered analysis with root cause and recommendations (briefly why)
3. **Second Reply** - Extracted test failure logs (evidence - only failure blocks, not full stdout)
4. **Third Reply** - Cluster information for debugging (least important - cluster is ephemeral)

**Note:** The code sends fallback messages (e.g., "Test output logs not available") when data is unavailable. This ensures the workflow is resilient to version drift between code and workflow changes.

### Message Format

**Summary (Initial Message - What Failed):**
```
:failed: Pipeline Failed at E2E Test

====== 🧪 Test Suite Information ======
• Image: `quay.io/openshift/osde2e-tests`
• Commit: `abc123`
• Environment: `stage`
```

**Analysis (First Reply - Briefly Why):**
```
====== 🔍 Possible Cause ======
<AI-generated root cause analysis>

====== 💡 Recommendations ======
1. <recommendation 1>
2. <recommendation 2>
```

**Extended Logs (Second Reply - Evidence):**
```
Found 3 test failure(s):

[FAILED] test description
<failure context lines>
...
```

**Cluster Details (Third Reply - For Debugging):**
```
====== ☸️ Cluster Information ======
• Cluster ID: `abc-123`
• Name: `my-cluster`
• Version: `4.20`
• Provider: `aws`
• Expiration: `2026-01-28T10:00:00Z`
```

### Testing

#### Unit Tests
```bash
# Run all reporter tests
go test -v github.com/openshift/osde2e/internal/reporter

# Run specific workflow tests
go test -v -run TestSlackReporter_buildWorkflowPayload
go test -v -run TestSlackReporter_extractFailureBlocks
```

#### Integration Test (with real Slack)
```bash
# Set environment variables
export LOG_ANALYSIS_SLACK_WEBHOOK="https://hooks.slack.com/workflows/..."
export LOG_ANALYSIS_SLACK_CHANNEL="C06HQR8HN0L"

# Run integration test
go test -v -run TestSlackReporter_Integration github.com/openshift/osde2e/pkg/e2e
```

**Note:** Integration test automatically skips if environment variables are not set.

### Workflow Payload Structure

The reporter sends this JSON payload to the Slack Workflow:

```json
{
  "channel": "C06HQR8HN0L",
  "summary": "Pipeline Failed at E2E Test\n\n# Test Suite Info...",
  "analysis": "# Possible Cause\n...",
  "extended_logs": "Found 3 test failure(s):\n...",
  "cluster_details": "# Cluster Information\nCluster ID: abc-123\n...",
  "image": "quay.io/openshift/osde2e:abc123",
  "env": "stage",
  "commit": "abc123"
}
```

## Implementation Notes

**Workflow vs Legacy Webhooks:**
- Workflow webhooks use `/workflows/` in the URL path
- Legacy incoming webhooks use `/services/` instead
- The code uses workflow webhooks to support threaded messages

**Payload Limits:**
- Maximum field length: 30KB per field (enforced by `maxWorkflowFieldLength` constant)
- Content exceeding limits is truncated with a notice
- Slack workflows handle much larger payloads than legacy webhooks

**Fallback Behavior:**
- All optional fields provide fallback messages when data is unavailable
- This ensures resilience to version drift between code and workflow changes
- Required fields: `channel`, `summary`, `analysis`

**Log Extraction Strategy:**
- For logs ≤250 lines: return full content
- For logs >250 lines: extract up to 3 failure blocks (max 30 lines each)
- Failure detection: `[FAILED]` markers and `ERROR`/`Error`/`error` strings
- Block deduplication: skip-ahead logic prevents overlapping extractions
