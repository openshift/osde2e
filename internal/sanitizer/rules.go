package sanitizer

// getDefaultRules returns built-in sanitization rules organized by priority.
func getDefaultRules() []Rule {
	return []Rule{
		// High Priority: Authentication Tokens

		{
			ID:          "aws-access-key",
			Pattern:     `AKIA[0-9A-Z]{16}`,
			Replacement: "[AWS-ACCESS-KEY-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "aws-secret-key",
			Pattern:     `(?i)(aws_secret_access_key|secret_key)["\s]*[:=]["\s]*[A-Za-z0-9/+=]{40}`,
			Replacement: "$1=[AWS-SECRET-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "github-token",
			Pattern:     `ghp_[A-Za-z0-9]{34,40}`,
			Replacement: "[GITHUB-TOKEN-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "jwt-token",
			Pattern:     `eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`,
			Replacement: "[JWT-TOKEN-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "bearer-token",
			Pattern:     `(?i)authorization:\s*bearer\s+[A-Za-z0-9_\-\.]{10,}`,
			Replacement: "Authorization: Bearer [TOKEN-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},

		// Medium Priority: Database & API Credentials
		{
			ID:          "api-key",
			Pattern:     `(?i)(api[_-]?key|apikey)["\s]*[:=]["\s]*[A-Za-z0-9_\-]{20,}`,
			Replacement: "$1=[API-KEY-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "db-password",
			Pattern:     `(?i)(password|pwd)["\s]*[:=]["\s]*[^\s"';]{8,}`,
			Replacement: "$1=[PASSWORD-REDACTED]",
			Category:    "password",
			Enabled:     true,
		},
		{
			ID:          "connection-string",
			Pattern:     `(?i)(mongodb|mysql|postgresql|postgres)://[^:]+:[^@]+@`,
			Replacement: "$1://[USER]:[PASSWORD-REDACTED]@",
			Category:    "password",
			Enabled:     true,
		},

		// OpenShift/Kubernetes Specific
		{
			ID:          "openshift-token",
			Pattern:     `sha256~[A-Za-z0-9_\-]{43}`,
			Replacement: "[OPENSHIFT-TOKEN-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "k8s-secret",
			Pattern:     `(?i)(secret|token):\s*[A-Za-z0-9+/=]{20,}`,
			Replacement: "$1: [SECRET-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},

		// Cryptographic Material
		{
			ID:          "private-key",
			Pattern:     `-----BEGIN[A-Z\s]*PRIVATE KEY-----[\s\S]*?-----END[A-Z\s]*PRIVATE KEY-----`,
			Replacement: "[PRIVATE-KEY-REDACTED]",
			Category:    "key",
			Enabled:     true,
		},

		// Additional High Priority Tokens
		{
			ID:          "docker-auth",
			Pattern:     `(?i)(docker[_-]?auth|dockercfg)["\s]*[:=]["\s]*[A-Za-z0-9+/=]{20,}`,
			Replacement: "$1=[DOCKER-AUTH-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "generic-token",
			Pattern:     `(?i)(token|access[_-]?token)["\s]*[:=]["\s]*[A-Za-z0-9_\-\.]{32,}`,
			Replacement: "$1=[TOKEN-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},

		// Additional Tokens
		{
			ID:          "slack-token",
			Pattern:     `xox[baprs]-[0-9]+-[0-9]+-[A-Za-z0-9]+`,
			Replacement: "[SLACK-TOKEN-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "azure-key",
			Pattern:     `(?i)(azure[_-]?key|subscription[_-]?key)["\s]*[:=]["\s]*[A-Za-z0-9+/=]{40,}`,
			Replacement: "$1=[AZURE-KEY-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},
		{
			ID:          "gcp-key",
			Pattern:     `(?i)(gcp[_-]?key|google[_-]?key)["\s]*[:=]["\s]*[A-Za-z0-9+/=]{40,}`,
			Replacement: "$1=[GCP-KEY-REDACTED]",
			Category:    "token",
			Enabled:     true,
		},

		// PII (Disabled by Default)
		{
			ID:          "email-address",
			Pattern:     `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
			Replacement: "[EMAIL-REDACTED]",
			Category:    "pii",
			Enabled:     false, // May cause false positives in CI logs
		},
	}
}
