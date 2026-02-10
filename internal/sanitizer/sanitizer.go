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
	SkipAuditOnNoMatch bool   `json:"skip_audit_on_no_match"` // Skip audit logging when no matches found
}

// Sanitizer provides data sanitization with configurable rules and audit logging.
type Sanitizer struct {
	auditLog      *AuditLogger
	config        *Config
	compiledRules []*compiledRule // Pre-compiled regex patterns for performance
}

// compiledRule holds a compiled regex pattern for faster matching
type compiledRule struct {
	rule  Rule
	regex *regexp.Regexp
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
			SkipAuditOnNoMatch: true, // Skip audit logging when no matches found (performance optimization)
		}
	}

	s := &Sanitizer{
		config: config,
	}

	// Pre-compile all regex patterns for better performance
	if err := s.compileRules(); err != nil {
		return nil, fmt.Errorf("failed to compile rules: %w", err)
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

// compileRules pre-compiles all enabled regex patterns for optimal performance
func (s *Sanitizer) compileRules() error {
	rules := getDefaultRules()
	s.compiledRules = make([]*compiledRule, 0, len(rules))

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			if s.config.StrictMode {
				return fmt.Errorf("invalid regex pattern for rule %s: %w", rule.ID, err)
			}
			continue // Skip invalid rules in graceful mode
		}

		s.compiledRules = append(s.compiledRules, &compiledRule{
			rule:  rule,
			regex: regex,
		})
	}

	return nil
}

// SanitizeBatch efficiently processes multiple content strings with batch optimizations
func (s *Sanitizer) SanitizeBatch(contents []string, sources []string) ([]*Result, error) {
	if len(contents) != len(sources) {
		return nil, fmt.Errorf("contents and sources length mismatch: %d vs %d", len(contents), len(sources))
	}

	// Pre-allocate results slice with exact capacity
	results := make([]*Result, len(contents))
	timestamp := time.Now() // Use single timestamp for the entire batch

	for i, content := range contents {
		source := sources[i] // Safe since we validated lengths match

		result, err := s.sanitizeContent(content, source, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to sanitize content %d: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// sanitizeContent processes content with pre-compiled rules and shared timestamp for batch efficiency
func (s *Sanitizer) sanitizeContent(content, source string, timestamp time.Time) (*Result, error) {
	if s.config.MaxContentSize > 0 && int64(len(content)) > s.config.MaxContentSize {
		return nil, fmt.Errorf("content size exceeds limit: %d > %d", len(content), s.config.MaxContentSize)
	}

	// Pre-allocate slices with estimated capacity
	rulesApplied := make([]string, 0, 4) // Most content has 0-4 matches
	matchCount := 0
	currentContent := content

	for _, compiledRule := range s.compiledRules {
		matches := compiledRule.regex.FindAllStringSubmatchIndex(currentContent, -1)
		if len(matches) == 0 {
			continue
		}

		matchCount += len(matches)
		rulesApplied = append(rulesApplied, compiledRule.rule.ID)
		currentContent = compiledRule.regex.ReplaceAllString(currentContent, compiledRule.rule.Replacement)
	}

	result := &Result{
		Content:      currentContent,
		Source:       source,
		Timestamp:    timestamp,
		RulesApplied: rulesApplied,
		MatchesFound: matchCount,
	}

	// Perform audit logging asynchronously to avoid blocking
	if s.auditLog != nil && (matchCount > 0 || !s.config.SkipAuditOnNoMatch) {
		go func() {
			_ = s.auditLog.Log(AuditEntry{
				Timestamp:    timestamp,
				Source:       source,
				RulesApplied: rulesApplied,
				MatchCount:   matchCount,
			})
		}()
	}

	return result, nil
}

// SanitizeText removes sensitive information using enabled rules.
func (s *Sanitizer) SanitizeText(content, source string) (*Result, error) {
	return s.sanitizeContent(content, source, time.Now())
}

// CleanupAuditLogs removes old audit logs based on retention policy.
func (s *Sanitizer) CleanupAuditLogs() error {
	if s.auditLog == nil {
		return nil // No audit logger configured
	}
	return s.auditLog.Cleanup(s.config.AuditRetentionDays)
}
