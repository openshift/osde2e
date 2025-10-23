# Data Sanitizer

A high-performance, configurable data sanitization library for removing sensitive information from CI/CD artifacts before LLM analysis. Designed for enterprise security and compliance requirements.

## Overview

The Data Sanitizer automatically detects and redacts sensitive information including authentication tokens, passwords, API keys, and personally identifiable information (PII) from text content. It's optimized for CI/CD pipelines where log files and configuration data need to be cleaned before analysis or storage.

## Key Features

- **Enterprise Security**: 14 built-in rules covering AWS, GitHub, JWT, OpenShift, Docker tokens
- **Highly Configurable**: Enable/disable rules, custom patterns, size limits, retention policies
- **Compliance Ready**: JSON audit logging with automatic cleanup and data retention
- **High Performance**: Efficient regex processing with async audit logging
- **Production Ready**: Strict/graceful error modes, content size limits, memory protection
- **Simple Integration**: Clean API without context dependencies for fast processing

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/openshift/osde2e/internal/sanitizer"
)

func main() {
    // Create sanitizer with default configuration
    s, err := sanitizer.New(nil)
    if err != nil {
        log.Fatal(err)
    }

    // Sanitize sensitive content
    input := "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE and password=secret123"
    result, err := s.SanitizeText(input, "config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Original: %s\n", input)
    fmt.Printf("Sanitized: %s\n", result.Content)
    fmt.Printf("Rules Applied: %v\n", result.RulesApplied)
    fmt.Printf("Matches Found: %d\n", result.MatchesFound)
}
```

**Output:**
```
Original: AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE and password=secret123
Sanitized: AWS_ACCESS_KEY_ID=[AWS-ACCESS-KEY-REDACTED] and password=[PASSWORD-REDACTED]
Rules Applied: [aws-access-key db-password]
Matches Found: 2
```

### Custom Configuration

```go
config := &sanitizer.Config{
    EnableAudit:        true,                    // Enable compliance audit logging
    AuditLogDir:        "/var/log/sanitizer",    // Custom audit log directory
    AuditRetentionDays: 90,                      // Keep audit logs for 90 days
    MaxContentSize:     50 * 1024 * 1024,       // 50MB content size limit
    StrictMode:         true,                    // Fail fast on any rule errors
}

s, err := sanitizer.New(config)
```

## Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `EnableAudit` | `bool` | `true` | Enable audit logging for compliance tracking |
| `AuditLogDir` | `string` | `"./logs"` | Directory for audit log files |
| `AuditRetentionDays` | `int` | `30` | Days to keep audit logs (0 = no cleanup) |
| `MaxContentSize` | `int64` | `10MB` | Maximum content size in bytes (0 = unlimited) |
| `StrictMode` | `bool` | `false` | Fail on rule errors vs graceful degradation |

## Supported Patterns

### High Priority (Always Enabled)
Critical authentication tokens that pose immediate security risks:

- **AWS Access Key**: `AKIA[0-9A-Z]{16}` → `[AWS-ACCESS-KEY-REDACTED]`
- **AWS Secret Key**: 40-character base64 secrets → `[AWS-SECRET-REDACTED]`
- **GitHub Token**: `ghp_[A-Za-z0-9]{34,40}` → `[GITHUB-TOKEN-REDACTED]`
- **JWT Token**: Standard JWT format → `[JWT-TOKEN-REDACTED]`
- **Bearer Token**: Authorization headers → `Authorization: Bearer [TOKEN-REDACTED]`
- **OpenShift Token**: `sha256~[A-Za-z0-9_-]{43}` → `[OPENSHIFT-TOKEN-REDACTED]`
- **Docker Auth**: Docker authentication tokens → `[DOCKER-AUTH-REDACTED]`
- **Generic Token**: Generic access tokens → `[TOKEN-REDACTED]`

### Medium Priority (Enabled by Default)
Common credentials in configuration files:

- **API Keys**: Generic API key patterns → `[API-KEY-REDACTED]`
- **Database Passwords**: Password field patterns → `[PASSWORD-REDACTED]`
- **Connection Strings**: DB URLs with credentials → `[USER]:[PASSWORD-REDACTED]@`
- **Kubernetes Secrets**: Base64 encoded values → `[SECRET-REDACTED]`
- **Private Keys**: PEM format keys → `[PRIVATE-KEY-REDACTED]`

### Optional (Disabled by Default)
May cause false positives in CI logs:

- **Email Addresses**: Standard email format → `[EMAIL-REDACTED]`

## Integration Examples

### With Aggregator

```go
// In your data aggregator
func (a *Aggregator) ProcessLogFile(filepath string) error {
    content, err := os.ReadFile(filepath)
    if err != nil {
        return err
    }

    // Sanitize before processing
    if a.sanitizer != nil {
        result, err := a.sanitizer.SanitizeText(string(content), filepath)
        if err != nil {
            log.Printf("Sanitization warning for %s: %v", filepath, err)
        } else {
            content = []byte(result.Content)
            log.Printf("Sanitized %s: %d secrets found", filepath, result.MatchesFound)
        }
    }

    return a.processContent(content)
}
```

### Processing Multiple Files

```go
// Process multiple files with a single sanitizer instance
s, err := sanitizer.New(nil)
if err != nil {
    return err
}

for _, filename := range files {
    content, err := os.ReadFile(filename)
    if err != nil {
        continue
    }

    result, err := s.SanitizeText(string(content), filename)
    if err != nil {
        log.Printf("Sanitization failed for %s: %v", filename, err)
        continue
    }

    log.Printf("Sanitized %s: %d secrets found", filename, result.MatchesFound)
    // Use result.Content for further processing
}
```


## Audit Logging

### Log Format
Audit logs are written in JSON format to `{AuditLogDir}/sanitizer-audit.log`:

```json
{"timestamp":"2025-10-15T10:30:00Z","source":"deployment.yaml","rules_applied":["aws-access-key","jwt-token"],"match_count":2}
{"timestamp":"2025-10-15T10:31:15Z","source":"config.json","rules_applied":["api-key"],"match_count":1}
```

### Data Retention

```go
// Automatic cleanup of old audit logs
s, err := sanitizer.New(&sanitizer.Config{
    EnableAudit:        true,
    AuditLogDir:        "/var/log/sanitizer",
    AuditRetentionDays: 30, // Automatic cleanup after 30 days
})
if err != nil {
    return err
}

// Manual cleanup can also be triggered
err = s.CleanupAuditLogs()
```

## Performance Characteristics

- **Thread Safety**: Fully concurrent-safe for multi-goroutine usage
- **Memory Efficient**: Streaming processing with configurable size limits
- **Regex Processing**: Compiled patterns for efficient matching
- **Async Logging**: Non-blocking audit logging to maintain throughput

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/sanitizer -bench=. -benchmem
```

## Testing

```bash
# Run tests
go test ./internal/sanitizer -v

# Run with coverage
go test ./internal/sanitizer -cover

# Run benchmarks
go test ./internal/sanitizer -bench=.
```

## Security Considerations

### Content Size Limits
Always set appropriate `MaxContentSize` limits to prevent:
- Memory exhaustion attacks
- Processing of unexpectedly large files
- Resource consumption in production

### Strict vs Graceful Mode
- **Strict Mode**: Fails fast on any rule compilation errors (recommended for development)
- **Graceful Mode**: Continues processing even if some rules fail (recommended for production)

### Audit Log Security
- Audit logs contain metadata only (no actual sensitive content)
- Ensure appropriate file permissions on audit log directories
- Implement log rotation and secure archival for compliance

## API Reference

### Types

```go
type Rule struct {
    ID          string `json:"id"`          // Unique rule identifier
    Pattern     string `json:"pattern"`     // Regular expression pattern
    Replacement string `json:"replacement"` // Replacement text
    Category    string `json:"category"`    // Rule category (token, password, pii, key)
    Enabled     bool   `json:"enabled"`     // Whether rule is active
}

type Result struct {
    Content      string    `json:"content"`        // Sanitized content
    MatchesFound int       `json:"matches_found"`  // Number of matches found
    RulesApplied []string  `json:"rules_applied"`  // Applied rule IDs
    Source       string    `json:"source"`         // Source identifier
    Timestamp    time.Time `json:"timestamp"`      // Processing timestamp
}

type Config struct {
    EnableAudit        bool   `json:"enable_audit"`        // Enable audit logging
    AuditLogDir        string `json:"audit_log_dir"`       // Audit log directory
    AuditRetentionDays int    `json:"audit_retention_days"` // Days to keep logs (0 = no cleanup)
    MaxContentSize     int64  `json:"max_content_size"`    // Content size limit (bytes)
    StrictMode         bool   `json:"strict_mode"`         // Error handling mode
}
```

### Methods

```go
// Create new sanitizer instance
func New(config *Config) (*Sanitizer, error)

// Sanitize text content
func (s *Sanitizer) SanitizeText(content, source string) (*Result, error)

// Clean up old audit logs
func (s *Sanitizer) CleanupAuditLogs() error

// Create audit logger (internal use)
func NewAuditLogger(logDir string) (*AuditLogger, error)

// Write audit entry (internal use)
func (a *AuditLogger) Log(entry AuditEntry) error

// Clean up old logs (internal use)
func (a *AuditLogger) Cleanup(retentionDays int) error
```