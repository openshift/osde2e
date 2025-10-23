// Package sanitizer removes sensitive information from CI/CD artifacts before LLM analysis.
package sanitizer

import (
	"fmt"
	"regexp"
	"time"
)

// Rule defines a sanitization rule with regex pattern and replacement.
type Rule struct {
	ID          string `json:"id"`
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	Category    string `json:"category"`
	Enabled     bool   `json:"enabled"`
}

// Result contains sanitization outcome and metadata.
type Result struct {
	Content      string    `json:"content"`
	MatchesFound int       `json:"matches_found"`
	RulesApplied []string  `json:"rules_applied"`
	Source       string    `json:"source"`
	Timestamp    time.Time `json:"timestamp"`
}

// Config holds sanitizer configuration.
type Config struct {
	EnableAudit        bool   `json:"enable_audit"`
	AuditLogDir        string `json:"audit_log_dir"`
	AuditRetentionDays int    `json:"audit_retention_days"` // 0 disables cleanup
	MaxContentSize     int64  `json:"max_content_size"`
	StrictMode         bool   `json:"strict_mode"`
}

// Sanitizer provides data sanitization with configurable rules and audit logging.
type Sanitizer struct {
	rules    []Rule
	auditLog *AuditLogger
	config   *Config
}

// New creates a sanitizer with default config if nil provided.
func New(config *Config) (*Sanitizer, error) {
	if config == nil {
		config = &Config{
			EnableAudit:        true,
			AuditLogDir:        "./logs",
			AuditRetentionDays: 30,               // Keep logs for 30 days
			MaxContentSize:     10 * 1024 * 1024, // 10MB
			StrictMode:         false,
		}
	}

	s := &Sanitizer{
		config: config,
		rules:  getDefaultRules(),
	}

	if config.EnableAudit {
		auditLog, err := NewAuditLogger(config.AuditLogDir)
		if err != nil && config.StrictMode {
			return nil, fmt.Errorf("failed to initialize audit logger: %w", err)
		}
		s.auditLog = auditLog
	}

	return s, nil
}

// SanitizeText removes sensitive information using enabled rules.
func (s *Sanitizer) SanitizeText(content, source string) (*Result, error) {
	if s.config.MaxContentSize > 0 && int64(len(content)) > s.config.MaxContentSize {
		return nil, fmt.Errorf("content size exceeds limit: %d > %d", len(content), s.config.MaxContentSize)
	}

	result := &Result{
		Content:      content,
		Source:       source,
		Timestamp:    time.Now(),
		RulesApplied: []string{},
	}

	matchCount := 0
	currentContent := content

	for _, rule := range s.rules {
		if !rule.Enabled {
			continue
		}

		matches, sanitized, err := s.applyRule(rule, currentContent)
		if err != nil {
			if s.config.StrictMode {
				return nil, fmt.Errorf("rule %s failed: %w", rule.ID, err)
			}
			continue // Skip failed rules in graceful mode
		}

		if matches > 0 {
			matchCount += matches
			result.RulesApplied = append(result.RulesApplied, rule.ID)
			currentContent = sanitized
		}
	}

	result.Content = currentContent
	result.MatchesFound = matchCount

	// Perform audit logging asynchronously to avoid blocking
	if s.auditLog != nil {
		go s.auditLog.Log(AuditEntry{
			Timestamp:    result.Timestamp,
			Source:       source,
			RulesApplied: result.RulesApplied,
			MatchCount:   matchCount,
		})
	}

	return result, nil
}

// applyRule applies a single rule and returns matches count, sanitized content, and any error.
func (s *Sanitizer) applyRule(rule Rule, content string) (int, string, error) {
	regex, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return 0, content, fmt.Errorf("invalid regex pattern for rule %s: %w", rule.ID, err)
	}

	matches := regex.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return 0, content, nil
	}

	result := regex.ReplaceAllString(content, rule.Replacement)
	return len(matches), result, nil
}

// CleanupAuditLogs removes old audit logs based on retention policy.
func (s *Sanitizer) CleanupAuditLogs() error {
	if s.auditLog == nil {
		return nil // No audit logger configured
	}
	return s.auditLog.Cleanup(s.config.AuditRetentionDays)
}
