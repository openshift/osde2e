package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// OMCClient manages must-gather data analysis using OMC binary
type OMCClient struct {
	workingDir     string
	mustGatherPath string
	initialized    bool
	omcBinaryPath  string
}

// NewOMCClient creates a new OMC client instance
func NewOMCClient() *OMCClient {
	return &OMCClient{}
}

// Initialize sets up the OMC client with the must-gather tar file
func (c *OMCClient) Initialize(ctx context.Context, mustGatherTarPath string) error {
	if c.initialized {
		return nil
	}

	// Create working directory
	workingDir, err := os.MkdirTemp("", "omc-analysis-*")
	if err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}
	c.workingDir = workingDir

	// Ensure OMC binary is available
	if err := c.ensureOMCBinary(ctx); err != nil {
		return fmt.Errorf("failed to ensure OMC binary: %w", err)
	}

	// Copy must-gather tar to our working directory to avoid polluting original location
	if err := c.copyMustGatherToWorkingDir(mustGatherTarPath); err != nil {
		return fmt.Errorf("failed to copy must-gather tar: %w", err)
	}

	// Initialize OMC with the copied tar file
	if err := c.initializeOMC(ctx); err != nil {
		return fmt.Errorf("failed to initialize OMC: %w", err)
	}

	c.initialized = true
	return nil
}

// ensureOMCBinary downloads OMC binary if not present
func (c *OMCClient) ensureOMCBinary(ctx context.Context) error {
	// Check if OMC is already in PATH
	if path, err := exec.LookPath("omc"); err == nil {
		c.omcBinaryPath = path
		return nil
	}

	c.omcBinaryPath = filepath.Join(c.workingDir, "omc")

	// Construct download URL with proper architecture mapping
	arch := c.mapArchitecture(runtime.GOARCH)
	osName, err := c.formatOSName(runtime.GOOS)
	if err != nil {
		return err
	}

	downloadURL := fmt.Sprintf("https://github.com/gmeghnag/omc/releases/latest/download/omc_%s_%s.tar.gz",
		osName, arch)

	return c.downloadAndExtractOMC(ctx, downloadURL)
}

// mapArchitecture maps Go architecture names to OMC release architecture names
func (c *OMCClient) mapArchitecture(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return goarch
	}
}

// formatOSName converts Go OS names to OMC release naming format
func (c *OMCClient) formatOSName(goos string) (string, error) {
	switch goos {
	case "linux":
		return "Linux", nil
	case "darwin":
		return "Darwin", nil
	default:
		return "", fmt.Errorf("operating system %q is not supported", goos)
	}
}

// downloadAndExtractOMC downloads and extracts the OMC binary
func (c *OMCClient) downloadAndExtractOMC(ctx context.Context, url string) error {
	// Create HTTP request with context and timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Download the tar.gz file
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download OMC: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download OMC: HTTP %d", resp.StatusCode)
	}

	// Save to temporary file
	tmpFile := filepath.Join(c.workingDir, "omc.tar.gz")
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save OMC archive: %w", err)
	}

	// Extract the binary
	return c.extractOMCBinary(tmpFile)
}

// extractOMCBinary extracts the OMC binary from tar.gz
func (c *OMCClient) extractOMCBinary(tarPath string) error {
	// Use tar command to extract
	cmd := exec.Command("tar", "-xzf", tarPath, "-C", c.workingDir, "omc")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract OMC binary: %w", err)
	}

	// Make binary executable
	if err := os.Chmod(c.omcBinaryPath, 0o755); err != nil {
		return fmt.Errorf("failed to make OMC executable: %w", err)
	}

	return nil
}

// copyMustGatherToWorkingDir copies the must-gather tar to our working directory
func (c *OMCClient) copyMustGatherToWorkingDir(originalPath string) error {
	// Generate destination path in our working directory
	tarFileName := filepath.Base(originalPath)
	copiedTarPath := filepath.Join(c.workingDir, tarFileName)

	// Copy the file
	src, err := os.Open(originalPath)
	if err != nil {
		return fmt.Errorf("failed to open source tar file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(copiedTarPath)
	if err != nil {
		return fmt.Errorf("failed to create destination tar file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy tar file: %w", err)
	}

	// Update mustGatherPath to point to the copied file in our working directory
	c.mustGatherPath = copiedTarPath

	return nil
}

// initializeOMC runs 'omc use' to initialize with the must-gather tar file
func (c *OMCClient) initializeOMC(ctx context.Context) error {
	// Run 'omc use <tar-file>' - OMC will handle extraction automatically
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.omcBinaryPath, "use", c.mustGatherPath)
	cmd.Dir = c.workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize OMC: %w, output: %s", err, string(output))
	}

	return nil
}

// ExecuteCommand runs an OMC command and returns the output
func (c *OMCClient) ExecuteCommand(ctx context.Context, command string) (string, error) {
	if !c.initialized {
		return "", fmt.Errorf("OMC client not initialized")
	}

	// Parse the command to ensure it's safe
	args := strings.Fields(command)
	if len(args) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Validate command starts with allowed operations
	allowedCommands := []string{"get", "describe", "logs", "explain", "version", "api-resources"}
	if !contains(allowedCommands, args[0]) {
		return "", fmt.Errorf("command '%s' not allowed", args[0])
	}

	// Execute the command with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.omcBinaryPath, args...)
	cmd.Dir = c.workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Cleanup removes temporary files and stops any running processes
func (c *OMCClient) Cleanup() error {
	if c.workingDir != "" {
		return os.RemoveAll(c.workingDir)
	}
	return nil
}
