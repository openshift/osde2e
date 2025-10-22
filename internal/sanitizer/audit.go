package sanitizer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// AuditEntry represents a sanitization audit log entry.
type AuditEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Source       string    `json:"source"`
	RulesApplied []string  `json:"rules_applied"`
	MatchCount   int       `json:"match_count"`
}

// AuditLogger handles audit logging for compliance.
type AuditLogger struct {
	logPath string
}

// NewAuditLogger creates an audit logger and ensures log directory exists.
func NewAuditLogger(logDir string) (*AuditLogger, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	return &AuditLogger{
		logPath: filepath.Join(logDir, "sanitizer-audit.log"),
	}, nil
}

// Log writes audit entry as JSON line.
func (a *AuditLogger) Log(entry AuditEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(a.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(string(data) + "\n")
	return err
}

// Cleanup removes audit logs older than retentionDays (0 disables cleanup).
func (a *AuditLogger) Cleanup(retentionDays int) error {
	if retentionDays <= 0 {
		return nil // Cleanup disabled
	}

	info, err := os.Stat(a.logPath)
	if os.IsNotExist(err) {
		return nil // No log file to clean
	}
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	if info.ModTime().Before(cutoff) {
		return os.Remove(a.logPath)
	}

	return nil
}
