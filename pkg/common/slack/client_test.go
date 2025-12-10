package common

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.timeout != DefaultTimeout {
		t.Errorf("expected timeout to be %v, got %v", DefaultTimeout, client.timeout)
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "custom timeout 10 seconds",
			timeout: 10 * time.Second,
		},
		{
			name:    "custom timeout 1 minute",
			timeout: 1 * time.Minute,
		},
		{
			name:    "custom timeout 5 seconds",
			timeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClientWithTimeout(tt.timeout)
			if client == nil {
				t.Fatal("expected non-nil client")
			}
			if client.timeout != tt.timeout {
				t.Errorf("expected timeout to be %v, got %v", tt.timeout, client.timeout)
			}
		})
	}
}

func TestClient_SendWebhook(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful webhook with map payload",
			payload: map[string]string{
				"text": "Test message",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("expected POST method, got %s", r.Method)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
					t.Errorf("expected Content-Type application/json; charset=utf-8, got %s", ct)
				}
				if ua := r.Header.Get("User-Agent"); ua != "osde2e/1.0" {
					t.Errorf("expected User-Agent osde2e/1.0, got %s", ua)
				}
				w.WriteHeader(http.StatusOK)
			},
			expectError: false,
		},
		{
			name: "successful webhook with struct payload",
			payload: struct {
				Text string `json:"text"`
			}{
				Text: "Test message",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectError: false,
		},
		{
			name:    "server returns error status",
			payload: map[string]string{"text": "Test"},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			expectError:   true,
			errorContains: "status 400",
		},
		{
			name:    "server returns 500",
			payload: map[string]string{"text": "Test"},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectError:   true,
			errorContains: "status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := NewClient()
			err := client.SendWebhook(context.Background(), server.URL, tt.payload)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.errorContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			}
		})
	}
}

func TestClient_SendMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		expectError bool
	}{
		{
			name:        "sends simple text message",
			message:     "Hello, Slack!",
			expectError: false,
		},
		{
			name:        "sends empty message",
			message:     "",
			expectError: false,
		},
		{
			name:        "sends message with special characters",
			message:     "Test with emoji 🚀 and symbols @#$%",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedBody map[string]interface{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&receivedBody)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient()
			err := client.SendMessage(context.Background(), server.URL, tt.message)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSendWebhook_PackageLevel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	payload := map[string]string{"text": "Test message"}
	err := SendWebhook(context.Background(), server.URL, payload)
	if err != nil {
		t.Errorf("unexpected error from package-level SendWebhook: %v", err)
	}
}

func TestClient_SendWebhook_InvalidPayload(t *testing.T) {
	client := NewClient()

	// Create an invalid payload that cannot be marshaled to JSON
	invalidPayload := make(chan int) // channels cannot be marshaled to JSON

	err := client.SendWebhook(context.Background(), "http://example.com", invalidPayload)

	if err == nil {
		t.Error("expected error for invalid payload, got none")
	}
	if !strings.Contains(err.Error(), "marshal") {
		t.Errorf("expected error to contain 'marshal', got: %v", err)
	}
}

func TestClient_SendWebhook_ContextCancellation(t *testing.T) {
	// Create a server that delays the response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient()
	err := client.SendWebhook(ctx, server.URL, map[string]string{"text": "test"})

	if err == nil {
		t.Error("expected error for cancelled context, got none")
	}
}

func TestClient_SendWebhook_InvalidURL(t *testing.T) {
	client := NewClient()

	// Use an invalid URL that will cause request creation to fail
	err := client.SendWebhook(context.Background(), "://invalid-url", map[string]string{"text": "test"})

	if err == nil {
		t.Error("expected error for invalid URL, got none")
	}
}
