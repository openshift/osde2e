package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultTimeout is the default HTTP client timeout for Slack webhook requests
	DefaultTimeout = 30 * time.Second
)

// Client provides methods for interacting with Slack webhooks
type Client struct {
	timeout time.Duration
}

// NewClient creates a new Slack client with default settings
func NewClient() *Client {
	return &Client{
		timeout: DefaultTimeout,
	}
}

// NewClientWithTimeout creates a new Slack client with a custom timeout
func NewClientWithTimeout(timeout time.Duration) *Client {
	return &Client{
		timeout: timeout,
	}
}

// SendWebhook sends a JSON payload to a Slack webhook URL
// payload can be any struct that will be marshaled to JSON
func (c *Client) SendWebhook(ctx context.Context, webhookURL string, payload interface{}) error {
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "osde2e/1.0")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: c.timeout,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// SendMessage is a convenience method to send a simple text message to Slack
func (c *Client) SendMessage(ctx context.Context, webhookURL string, text string) error {
	msg := text
	return c.SendWebhook(ctx, webhookURL, msg)
}

// SendWebhook is a package-level convenience function that sends a payload using the default client
func SendWebhook(ctx context.Context, webhookURL string, payload interface{}) error {
	client := NewClient()
	return client.SendWebhook(ctx, webhookURL, payload)
}

// PostMessageWithFiles posts a message to Slack with file attachments using chat.postMessage + files.upload
// This provides a richer experience than webhooks by attaching files to the message
func (c *Client) PostMessageWithFiles(ctx context.Context, botToken string, channel string, message interface{}, filePaths []string) error {
	// First, post the message to get a timestamp
	ts, err := c.postMessage(ctx, botToken, channel, message)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	// Then upload files as threaded replies to that message
	if len(filePaths) > 0 {
		for _, filePath := range filePaths {
			if err := c.uploadFileToThread(ctx, botToken, channel, ts, filePath); err != nil {
				// Log error but continue with other files
				fmt.Fprintf(os.Stderr, "Warning: Failed to upload %s: %v\n", filePath, err)
			}
		}
	}

	return nil
}

// postMessage posts a message using chat.postMessage and returns the message timestamp
func (c *Client) postMessage(ctx context.Context, botToken string, channel string, message interface{}) (string, error) {
	// Extract text content from message
	var text string
	if msg, ok := message.(map[string]interface{}); ok {
		if summary, ok := msg["summary"].(string); ok {
			text = summary
		}
		if analysis, ok := msg["analysis"].(string); ok {
			if text != "" {
				text += "\n\n"
			}
			text += analysis
		}
	}

	// Create payload for chat.postMessage
	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+botToken)

	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if ok, exists := result["ok"].(bool); !exists || !ok {
		if errMsg, exists := result["error"].(string); exists {
			return "", fmt.Errorf("slack API error: %s", errMsg)
		}
		return "", fmt.Errorf("slack API returned ok=false")
	}

	// Extract timestamp
	ts, _ := result["ts"].(string)
	return ts, nil
}

// uploadFileToThread uploads a file to a Slack thread
func (c *Client) uploadFileToThread(ctx context.Context, botToken string, channel string, threadTs string, filePath string) error {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add channel field
	if err := writer.WriteField("channels", channel); err != nil {
		return fmt.Errorf("failed to write channel field: %w", err)
	}

	// Add thread_ts to attach to message thread
	if threadTs != "" {
		if err := writer.WriteField("thread_ts", threadTs); err != nil {
			return fmt.Errorf("failed to write thread_ts field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/files.upload", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+botToken)

	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if ok, exists := result["ok"].(bool); !exists || !ok {
		if errMsg, exists := result["error"].(string); exists {
			return fmt.Errorf("slack API error: %s", errMsg)
		}
		return fmt.Errorf("slack API returned ok=false")
	}

	return nil
}
