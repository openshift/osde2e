// Package ephemeral provides tooling for on-demand ROSA HCP cluster
// lifecycle management through the bonfire CLI and oc commands.
//
// bonfire (https://github.com/RedHatInsights/bonfire) is a Python CLI
// that interacts with the eng-prod management cluster's ephemeral
// namespace operator to provision and tear down ROSA HCP clusters.
// This package shells out to bonfire rather than reimplementing its
// qontract/CAPI template processing logic.
package ephemeral

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
)

// CommandRunner abstracts external process execution for testability.
type CommandRunner interface {
	// Run executes a command and returns its stdout as a trimmed string.
	// Extra environment variables (in "KEY=VALUE" form) are appended to
	// the current process environment.
	Run(ctx context.Context, env []string, name string, args ...string) (string, error)

	// LookPath searches for an executable named file in the directories
	// named by the PATH environment variable.
	LookPath(file string) (string, error)
}

// CLI wraps the bonfire and oc command-line tools used for managing
// ephemeral ROSA HCP cluster namespaces on the eng-prod management cluster.
//
// Call Cleanup when done to remove any temporary kubeconfig created by Login.
//
// All methods that invoke bonfire set BONFIRE_BOT=true to suppress
// interactive oc context switching.
type CLI struct {
	cmd        CommandRunner
	log        logr.Logger
	kubeconfig string // temp file written by Login; propagated via KUBECONFIG env
}

// NewCLI returns a CLI backed by real OS process execution.
func NewCLI(logger logr.Logger) *CLI {
	return &CLI{cmd: &osCommandRunner{}, log: logger}
}

// NewCLIWithRunner returns a CLI with a custom CommandRunner.
// Use this in tests to inject a mock.
func NewCLIWithRunner(logger logr.Logger, r CommandRunner) *CLI {
	return &CLI{cmd: r, log: logger}
}

// Cleanup removes the temporary kubeconfig written by Login.
// Safe to call multiple times or when Login was never called.
func (c *CLI) Cleanup() {
	if c.kubeconfig != "" {
		os.Remove(c.kubeconfig)
		c.kubeconfig = ""
	}
}

// InstallBonfire ensures the bonfire CLI is available on PATH.
// If already present this is a no-op; otherwise it runs
// `pip install crc-bonfire`.
func (c *CLI) InstallBonfire(ctx context.Context) error {
	if _, err := c.cmd.LookPath("bonfire"); err == nil {
		c.log.Info("bonfire CLI already on PATH")
		return nil
	}

	c.log.Info("installing crc-bonfire via pip")
	if _, err := c.cmd.Run(ctx, nil, "pip", "install", "--quiet", "crc-bonfire"); err != nil {
		return fmt.Errorf("pip install crc-bonfire: %w", err)
	}

	if _, err := c.cmd.LookPath("bonfire"); err != nil {
		return fmt.Errorf("bonfire not on PATH after install: %w", err)
	}
	c.log.Info("bonfire installed successfully")
	return nil
}

// Login authenticates to the eng-prod management cluster by writing a
// temporary kubeconfig file with the provided bearer token. The kubeconfig
// is propagated to subsequent bonfire and oc calls via the KUBECONFIG
// environment variable, keeping the token out of process argument lists.
func (c *CLI) Login(ctx context.Context, serverURL, token string) error {
	if serverURL == "" {
		return fmt.Errorf("management cluster URL must not be empty")
	}
	if token == "" {
		return fmt.Errorf("service account token must not be empty")
	}

	c.Cleanup() // remove any leftover from a previous call

	content := fmt.Sprintf(kubeconfigTemplate, serverURL, token)

	f, err := os.CreateTemp("", "ephemeral-kubeconfig-*")
	if err != nil {
		return fmt.Errorf("creating temp kubeconfig: %w", err)
	}
	path := f.Name()

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(path)
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	f.Close()

	c.log.Info("verifying management cluster connectivity", "server", serverURL)
	if _, err := c.cmd.Run(ctx, []string{"KUBECONFIG=" + path}, "oc", "whoami"); err != nil {
		os.Remove(path)
		return fmt.Errorf("verifying management cluster connectivity: %w", err)
	}

	c.kubeconfig = path
	c.log.Info("authenticated to management cluster", "server", serverURL)
	return nil
}

// DeployROSA runs `bonfire deploy rosa` which:
//  1. Reserves a namespace from the given pool.
//  2. Fetches and applies the ROSA HCP cluster template.
//  3. Waits up to timeoutSeconds for all resources to become ready.
//
// Returns the reserved namespace name on success.
func (c *CLI) DeployROSA(ctx context.Context, pool, duration string, timeoutSeconds int) (string, error) {
	args := []string{
		"deploy", "rosa",
		"--reserve",
		"--pool", pool,
		"--duration", duration,
		"--timeout", fmt.Sprintf("%d", timeoutSeconds),
	}

	c.log.Info("deploying ROSA HCP cluster", "pool", pool, "duration", duration, "timeout", timeoutSeconds)
	out, err := c.cmd.Run(ctx, c.mergeEnv("BONFIRE_BOT=true"), "bonfire", args...)
	if err != nil {
		return "", fmt.Errorf("bonfire deploy rosa: %w", err)
	}

	ns := parseNamespace(out)
	if ns == "" {
		return "", fmt.Errorf("could not parse namespace from bonfire output:\n%s", out)
	}
	c.log.Info("ROSA cluster deploying", "namespace", ns)
	return ns, nil
}

// ReleaseNamespace tells the ephemeral-namespace-operator to tear down
// the reserved namespace and deprovision any ROSA clusters within it.
func (c *CLI) ReleaseNamespace(ctx context.Context, namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace must not be empty")
	}

	c.log.Info("releasing namespace", "namespace", namespace)
	if _, err := c.cmd.Run(ctx, c.mergeEnv("BONFIRE_BOT=true"),
		"bonfire", "namespace", "release", namespace, "--force",
	); err != nil {
		return fmt.Errorf("bonfire namespace release %s: %w", namespace, err)
	}
	c.log.Info("namespace released", "namespace", namespace)
	return nil
}

// GetSecretValue retrieves a single base64-encoded field from a Kubernetes
// Secret using `oc get secret -o jsonpath=...`. The raw jsonpath output
// is returned without decoding.
func (c *CLI) GetSecretValue(ctx context.Context, secretName, namespace, dataKey string) (string, error) {
	out, err := c.cmd.Run(ctx, c.mergeEnv(),
		"oc", "get", "secret", secretName,
		"-n", namespace,
		"-o", fmt.Sprintf("jsonpath={.data.%s}", dataKey),
	)
	if err != nil {
		return "", fmt.Errorf("reading secret %s/%s key %q: %w", namespace, secretName, dataKey, err)
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mergeEnv builds the environment slice for a command, prepending the
// management-cluster KUBECONFIG (if Login was called) to any extras.
func (c *CLI) mergeEnv(extra ...string) []string {
	var env []string
	if c.kubeconfig != "" {
		env = append(env, "KUBECONFIG="+c.kubeconfig)
	}
	env = append(env, extra...)
	if len(env) == 0 {
		return nil
	}
	return env
}

// ---------------------------------------------------------------------------
// Output parsing
// ---------------------------------------------------------------------------

// parseNamespace extracts the namespace name from bonfire deploy stdout.
//
// bonfire routes all logging and Rich console output to stderr; the only
// content written to stdout is a bare `click.echo(ns)` call — the namespace
// name alone (e.g. "ephemeral-abc1234f"). The log-prefix filter below is a
// defensive fallback in case that assumption ever changes.
func parseNamespace(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "INFO") ||
			strings.HasPrefix(upper, "WARNING") ||
			strings.HasPrefix(upper, "ERROR") ||
			strings.HasPrefix(upper, "DEBUG") {
			continue
		}
		return line
	}
	return ""
}

// ---------------------------------------------------------------------------
// Kubeconfig template
// ---------------------------------------------------------------------------

// kubeconfigTemplate is a minimal kubeconfig that authenticates via bearer
// token. Written to a temp file by Login to avoid passing the token in
// process argument lists.
//
// Format args: serverURL (%s), token (%s).
const kubeconfigTemplate = `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: %s
    insecure-skip-tls-verify: true
  name: ephemeral-mgmt
contexts:
- context:
    cluster: ephemeral-mgmt
    user: ephemeral-mgmt
  name: ephemeral-mgmt
current-context: ephemeral-mgmt
users:
- name: ephemeral-mgmt
  user:
    token: %s
`

// ---------------------------------------------------------------------------
// Default CommandRunner implementation (os/exec)
// ---------------------------------------------------------------------------

type osCommandRunner struct{}

func (r *osCommandRunner) Run(ctx context.Context, env []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("%s failed: %w\n%s", name, err, detail)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (r *osCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
