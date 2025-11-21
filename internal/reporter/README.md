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
