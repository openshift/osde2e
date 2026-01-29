# Reporter System

The reporter system handles notification delivery after LLM analysis completion, providing a flexible and extensible way to send analysis results to external systems.

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

The workflow creates three messages in a thread:

1. **Initial Message** - Failure summary with cluster and test suite info
2. **First Reply** - AI-powered analysis with root cause and recommendations
3. **Second Reply** - Extracted test failure logs (only failure blocks, not full stdout)

### Setup Instructions

#### 1. Add Workflow to Your Slack Channel

Each team adds the shared workflow to their channel:

1. Open the workflow link: https://slack.com/shortcuts/Ft09RL7M2AMV/60f07b46919da20d103806a8f5bba094
2. Click **Add to Slack**
3. Select your destination channel
4. **Copy the webhook URL** (starts with `https://hooks.slack.com/workflows/...`)

#### 2. Get Your Channel ID

The workflow requires a Slack **channel ID** (not channel name).

**To find your channel ID:**
1. Right-click the channel name in Slack
2. Select **View channel details**
3. Scroll to bottom and **copy the channel ID** (starts with `C`)

**Example:** `C06HQR8HN0L`

#### 3. Configure Pipeline

Set these environment variables in your CI/CD pipeline or Vault:

```bash
LOG_ANALYSIS_SLACK_WEBHOOK=https://hooks.slack.com/workflows/T.../A.../...
LOG_ANALYSIS_SLACK_CHANNEL=C06HQR8HN0L  # Channel ID, not #channel-name
```

#### 4. Enable in Config

```yaml
tests:
  enableSlackNotify: true
logAnalysis:
  enableAnalysis: true
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `LOG_ANALYSIS_SLACK_WEBHOOK` | Yes | Workflow webhook URL from step 1 |
| `LOG_ANALYSIS_SLACK_CHANNEL` | Yes | Channel ID (starts with `C`) |

### Message Format

**Summary (Initial Message):**
```
:failed: Pipeline Failed at E2E Test

====== ☸️ Cluster Information ======
• Cluster ID: `abc-123`
• Name: `my-cluster`
• Version: `4.20`
• Provider: `aws`
• Expiration: `2026-01-28T10:00:00Z`

====== 🧪 Test Suite Information ======
• Image: `quay.io/openshift/osde2e-tests`
• Commit: `abc123`
• Environment: `stage`
```

**Analysis (First Reply):**
```
====== 🔍 Possible Cause ======
<AI-generated root cause analysis>

====== 💡 Recommendations ======
1. <recommendation 1>
2. <recommendation 2>
```

**Extended Logs (Second Reply):**
```
Found 3 test failure(s):

[FAILED] test description
<failure context lines>
...
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
  "summary": "Pipeline Failed at E2E Test\n\n# Cluster Info...",
  "analysis": "# Possible Cause\n...",
  "extended_logs": "Found 3 test failure(s):\n...",
  "image": "quay.io/openshift/osde2e:abc123",
  "env": "stage",
  "commit": "abc123"
}
```

### Troubleshooting

**Workflow not posting threaded messages:**
- Verify webhook URL is from the workflow (not a legacy incoming webhook)
- Workflow URLs contain `/workflows/` in the path
- Legacy incoming webhook URLs contain `/services/` instead

**Channel not receiving messages:**
- Ensure you're using the channel ID (starts with `C`), not channel name
- Channel ID is case-sensitive

**Missing fields in Slack message:**
- Check that all required fields are present: `channel`, `summary`, `analysis`
- Verify environment variables are set correctly

**Analysis too long:**
- The workflow handles message splitting automatically
- Payload limits: 30KB per field (enforced by code)
