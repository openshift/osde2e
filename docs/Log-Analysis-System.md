# Log Analysis System

The Log Analysis System provides intelligent failure analysis and notification capabilities for osde2e test runs. It automatically analyzes test failures using LLM-powered insights and can send notifications to configured channels.

## Architecture

The system consists of three main components:

```
╔════════════════════╗         ╔════════════════════╗         ╔════════════════════╗
║  Data Sanitizer    ║────────▶║  Analysis Engine   ║────────▶║  Reporter System   ║
╠════════════════════╣         ╠════════════════════╣         ╠════════════════════╣
║                    ║         ║                    ║         ║                    ║
║ • Remove secrets   ║         ║ • LLM analysis     ║         ║ • Slack notifier   ║
║ • Audit logging    ║         ║ • Root cause       ║         ║ • Extensible       ║
║ • PII protection   ║         ║ • Recommendations  ║         ║ • Multi-channel    ║
║                    ║         ║                    ║         ║                    ║
╚════════════════════╝         ╚════════════════════╝         ╚════════════════════╝
```

### 1. Data Sanitizer (`internal/sanitizer`)

A high-performance, configurable data sanitization library that removes sensitive information from CI/CD artifacts before LLM analysis.

**Key Features:**
- **Enterprise Security**: 14 built-in rules covering AWS, GitHub, JWT, OpenShift, Docker tokens
- **Highly Configurable**: Enable/disable rules, custom patterns, size limits, retention policies
- **Compliance Ready**: JSON audit logging with automatic cleanup and data retention
- **High Performance**: Efficient regex processing with async audit logging
- **Production Ready**: Strict/graceful error modes, content size limits, memory protection

**Quick Example:**

```go
s, err := sanitizer.New(nil)
if err != nil {
    log.Fatal(err)
}

result, err := s.SanitizeText(content, "config.yaml")
fmt.Printf("Sanitized: %s\n", result.Content)
fmt.Printf("Secrets Found: %d\n", result.MatchesFound)
```

**Configuration:**

```go
config := &sanitizer.Config{
    EnableAudit:        true,                    // Enable compliance audit logging
    AuditLogDir:        "/var/log/sanitizer",    // Custom audit log directory
    AuditRetentionDays: 90,                      // Keep audit logs for 90 days
    MaxContentSize:     50 * 1024 * 1024,       // 50MB content size limit
    StrictMode:         true,                    // Fail fast on any rule errors
}
```

**Supported Patterns:**

High Priority (Always Enabled):
- AWS Access Key: `AKIA[0-9A-Z]{16}` → `[AWS-ACCESS-KEY-REDACTED]`
- GitHub Token: `ghp_[A-Za-z0-9]{34,40}` → `[GITHUB-TOKEN-REDACTED]`
- JWT Token → `[JWT-TOKEN-REDACTED]`
- OpenShift Token: `sha256~[A-Za-z0-9_-]{43}` → `[OPENSHIFT-TOKEN-REDACTED]`
- Bearer Token → `Authorization: Bearer [TOKEN-REDACTED]`

Medium Priority (Enabled by Default):
- API Keys → `[API-KEY-REDACTED]`
- Database Passwords → `[PASSWORD-REDACTED]`
- Connection Strings → `[USER]:[PASSWORD-REDACTED]@`
- Kubernetes Secrets → `[SECRET-REDACTED]`
- Private Keys → `[PRIVATE-KEY-REDACTED]`

### 2. Analysis Engine (`internal/analysisengine`)

Analyzes test failures and artifacts using LLM (Gemini) to provide intelligent insights and root cause analysis.

**Workflow:**

```
╔═══════════════╗        ╔═══════════════╗        ╔═══════════════╗        ╔═══════════════╗
║     Test      ║        ║  Aggregator   ║        ║  PromptStore  ║        ║  LLM Client   ║
║   Artifacts   ║───────▶║   Collects:   ║───────▶║    Renders    ║───────▶║   (Gemini)    ║
╟───────────────╢        ╟───────────────╢        ╟───────────────╢        ╟───────────────╢
║               ║        ║ • JUnit XML   ║        ║ • Templates   ║        ║ • Analyzes    ║
║ • Logs        ║        ║ • Log files   ║        ║ • Variables   ║        ║ • Root cause  ║
║ • Results     ║        ║ • Failed      ║        ║ • Context     ║        ║ • AI insights ║
║ • Failures    ║        ║   tests       ║        ║   injection   ║        ║               ║
╚═══════════════╝        ╚═══════════════╝        ╚═══════════════╝        ╚═══════╤═══════╝
                                                                                    │
                                                                                    ▼
╔════════════════════════════════════════════════════════════════════════════════════════╗
║                         Output: llm-analysis/summary.yaml                              ║
╠════════════════════════════════════════════════════════════════════════════════════════╣
║  • Analysis results  • Cluster info  • Metadata  • Original prompt  • Recommendations  ║
╚════════════════════════════════════════════════════════════════════════════════════════╝
```

**Components:**
- **Engine**: Orchestrates the analysis workflow
- **Config**: Analysis configuration (API keys, templates, cluster info)
- **ClusterInfo**: Cluster metadata (ID, provider, version, etc.)
- **Result**: Analysis output with summary and metadata

**Usage:**

```go
engine, err := analysisengine.New(ctx, &analysisengine.Config{
    ArtifactsDir:   "/path/to/artifacts",
    PromptTemplate: "default",
    APIKey:         os.Getenv("GEMINI_API_KEY"),
    ClusterInfo:    clusterInfo,
})

result, err := engine.Run(ctx)
```

**Output:**

Creates `llm-analysis/summary.yaml` with:
- LLM analysis and recommendations
- Cluster and failure context
- Examined artifacts count
- Complete prompt and response data

### 3. Reporter System (`internal/reporter`)

Handles notification delivery after LLM analysis completion, providing a flexible and extensible way to send analysis results to external systems.

**Architecture:**

```
╔═══════════════════╗        ╔═══════════════════════╗        ╔═══════════════════════╗
║ Analysis Engine   ║        ║ NotificationConfig    ║        ║  ReporterRegistry     ║
║ completes LLM     ║───────▶║                       ║───────▶║                       ║
║ analysis          ║        ║ • Enabled: bool       ║        ║ • Manages reporters   ║
╚═══════════════════╝        ║ • Reporters: []       ║        ║ • Get by type         ║
                             ╚═══════════════════════╝        ╚═══════════╤═══════════╝
                                                                           │
                             ╔═════════════════════════════════════════════▼═════════╗
                             ║         Reporter Processing Loop                      ║
                             ╟═══════════════════════════════════════════════════════╣
                             ║  For each ReporterConfig in Reporters array:          ║
                             ║    1. Check if config.Enabled = true                  ║
                             ║    2. Get reporter by config.Type from registry       ║
                             ║    3. Call reporter.Report(ctx, result, config)       ║
                             ╚═══════════════════════════╤═══════════════════════════╝
                                                         │
                             ╔═══════════════════════════▼═══════════════════════════╗
                             ║         SlackReporter Implementation                  ║
                             ╟═══════════════════════════════════════════════════════╣
                             ║  • Extracts webhook_url from config                   ║
                             ║  • Formats analysis message with results              ║
                             ║  • Sends notification to Slack via HTTP POST          ║
                             ╚═══════════════════════════════════════════════════════╝
```

**Configuration:**

```go
// E2E usage example (pkg/e2e/e2e.go)
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

## Environment Variables

The following environment variables control the log analysis system:

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `GEMINI_API_KEY` | Google Gemini API key for LLM analysis | Yes (for analysis) | - |
| `ENABLE_LLM_ANALYSIS` | Enable/disable LLM-based failure analysis | No | `false` |
| `ENABLE_SLACK_NOTIFY` | Enable Slack notifications | No | `false` |
| `SLACK_WEBHOOK_URL` | Slack webhook URL for notifications | Yes (if Slack enabled) | - |
| `SANITIZER_AUDIT_DIR` | Directory for sanitizer audit logs | No | `./logs` |
| `SANITIZER_RETENTION_DAYS` | Days to keep audit logs | No | `30` |
| `SANITIZER_STRICT_MODE` | Fail fast on sanitization errors | No | `false` |

## CLI Flags

Log analysis can also be configured via command-line flags:

```bash
./osde2e test \
  --enable-llm-analysis \
  --enable-slack-notify \
  --slack-webhook-url="https://hooks.slack.com/services/YOUR/WEBHOOK/URL" \
  --configs=prod,e2e-suite
```

## Integration Example

Here's a complete example of using the log analysis system in your test workflow:

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/openshift/osde2e/internal/analysisengine"
    "github.com/openshift/osde2e/internal/reporter"
    "github.com/openshift/osde2e/internal/sanitizer"
)

func main() {
    ctx := context.Background()

    // 1. Create sanitizer
    sanitizerConfig := &sanitizer.Config{
        EnableAudit:        true,
        AuditLogDir:        "/var/log/sanitizer",
        AuditRetentionDays: 30,
        MaxContentSize:     10 * 1024 * 1024, // 10MB
        StrictMode:         false,
    }

    s, err := sanitizer.New(sanitizerConfig)
    if err != nil {
        log.Fatal(err)
    }

    // 2. Run analysis engine
    engine, err := analysisengine.New(ctx, &analysisengine.Config{
        ArtifactsDir:   "/path/to/artifacts",
        PromptTemplate: "default",
        APIKey:         os.Getenv("GEMINI_API_KEY"),
        ClusterInfo: &analysisengine.ClusterInfo{
            ID:       "cluster-123",
            Provider: "rosa",
            Version:  "4.14.0",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    result, err := engine.Run(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // 3. Send notifications
    notificationConfig := &reporter.NotificationConfig{
        Enabled: true,
        Reporters: []reporter.ReporterConfig{
            reporter.SlackReporterConfig(
                os.Getenv("SLACK_WEBHOOK_URL"),
                true,
            ),
        },
    }

    // Process reporters
    registry := reporter.NewRegistry()
    for _, config := range notificationConfig.Reporters {
        if !config.Enabled {
            continue
        }

        rep := registry.Get(config.Type)
        if rep == nil {
            log.Printf("Unknown reporter type: %s", config.Type)
            continue
        }

        if err := rep.Report(ctx, result, config); err != nil {
            log.Printf("Failed to send notification: %v", err)
        }
    }
}
```

## Testing

Each component has comprehensive unit tests:

```bash
# Test sanitizer
go test ./internal/sanitizer -v

# Test analysis engine
go test ./internal/analysisengine -v

# Test reporter
go test ./internal/reporter -v

# Run all internal tests
go test ./internal/... -v
```

## Performance Considerations

### Sanitizer
- Thread-safe for concurrent use
- Compiled regex patterns for efficiency
- Async audit logging to maintain throughput
- Configurable content size limits to prevent memory exhaustion

### Analysis Engine
- Processes artifacts incrementally
- Caches prompt templates
- Handles large log files efficiently

### Reporter
- Non-blocking notification delivery
- Retry logic for failed deliveries
- Extensible for multiple notification channels

## Security Best Practices

1. **Always sanitize before analysis**: Ensure all artifacts are sanitized before sending to LLM
2. **Protect API keys**: Never commit API keys; use environment variables
3. **Audit logging**: Enable audit logging in production for compliance
4. **Content size limits**: Set appropriate limits to prevent resource exhaustion
5. **Webhook security**: Protect webhook URLs; rotate regularly

## Troubleshooting

### Sanitizer Issues

**Problem**: Too many false positives
- **Solution**: Disable optional rules or adjust patterns in custom configuration

**Problem**: Audit logs filling up disk
- **Solution**: Reduce `AuditRetentionDays` or disable audit logging

### Analysis Engine Issues

**Problem**: "API key not found" error
- **Solution**: Set `GEMINI_API_KEY` environment variable

**Problem**: Analysis timeout
- **Solution**: Reduce artifacts size or increase timeout in configuration

### Reporter Issues

**Problem**: Slack notifications not arriving
- **Solution**: Verify webhook URL is correct and Slack app is properly configured

**Problem**: Notification failures
- **Solution**: Check logs for detailed error messages; verify network connectivity

## Extending the System

### Adding Custom Sanitization Rules

```go
customRules := []sanitizer.Rule{
    {
        ID:          "custom-pattern",
        Pattern:     `MY_PATTERN_HERE`,
        Replacement: "[CUSTOM-REDACTED]",
        Category:    "token",
        Enabled:     true,
    },
}

// Merge with default rules during initialization
```

### Adding New Reporter Types

Implement the `Reporter` interface:

```go
type CustomReporter struct{}

func (r *CustomReporter) Report(ctx context.Context, result *Result, config ReporterConfig) error {
    // Your implementation here
    return nil
}

// Register in the registry
registry.Register("custom", &CustomReporter{})
```

## Related Documentation

- [Configuration Reference](/docs/Config.md) - Complete configuration options
- [Writing Tests](/docs/Writing-Tests.md) - Guidelines for writing tests
- [CI Jobs](/docs/CI-Jobs.md) - CI/CD integration patterns
