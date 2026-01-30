package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
		// Read response body for debugging
		bodyBytes := make([]byte, 1024)
		n, _ := resp.Body.Read(bodyBytes)
		bodyText := string(bodyBytes[:n])

		return fmt.Errorf("slack webhook returned status %d: %s\nResponse body: %s\nPayload sent: %s",
			resp.StatusCode, resp.Status, bodyText, string(jsonData))
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
