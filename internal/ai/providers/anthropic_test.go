package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

func TestAnthropicProvider_MetadataAndConfigure(t *testing.T) {
	p := &AnthropicProvider{}
	if p.Name() != "Anthropic (Claude)" {
		t.Errorf("Expected Name 'Anthropic (Claude)', got '%s'", p.Name())
	}
	if p.IsLocal() != false {
		t.Errorf("Expected IsLocal to be false")
	}

	cfg := &config.Config{
		AIModel:  "claude-3-5-sonnet-20241022",
		AIAPIKey: "test-anthropic-key",
		AIBaseURL: "https://custom.anthropic.com/v1",
	}

	if err := p.Configure(cfg); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if p.Model() != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got '%s'", p.Model())
	}
	if p.APIKey != "test-anthropic-key" {
		t.Errorf("Expected APIKey 'test-anthropic-key', got '%s'", p.APIKey)
	}
	if p.BaseURL != "https://custom.anthropic.com/v1" {
		t.Errorf("Expected BaseURL 'https://custom.anthropic.com/v1', got '%s'", p.BaseURL)
	}
}

func TestAnthropicProvider_Send_Success(t *testing.T) {
	mockResponse := anthropicResponse{
		Content: []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		}{
			{Text: "Hello from Claude!", Type: "text"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages" {
			t.Errorf("Expected path /messages, got %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-anthropic-key" {
			t.Errorf("Expected x-api-key 'test-anthropic-key', got '%s'", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Expected anthropic-version '2023-06-01', got '%s'", r.Header.Get("anthropic-version"))
		}

		var req anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if len(req.Messages) != 1 || req.Messages[0].Content != "Hi" {
			t.Errorf("Unexpected messages: %+v", req.Messages)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	p := &AnthropicProvider{}
	cfg := &config.Config{
		AIBaseURL: server.URL,
		AIAPIKey:  "test-anthropic-key",
	}
	if err := p.Configure(cfg); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	resp, err := p.Send([]ai.Message{
		{Role: "system", Content: "System prompt ignored"},
		{Role: "user", Content: "Hi"},
	})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if resp != "Hello from Claude!" {
		t.Errorf("Expected 'Hello from Claude!', got '%s'", resp)
	}
}

func TestAnthropicProvider_Send_Errors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "Unauthorized 401",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"error": {"type": "authentication_error", "message": "invalid api key"}}`,
			expectedError: "claude: invalid API Key or access denied",
		},
		{
			name:          "Not Found 404",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"error": {"type": "not_found_error", "message": "model not found"}}`,
			expectedError: "claude: model 'claude-3-opus-20240229' not found or no access",
		},
		{
			name:          "Too Many Requests 429",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `{"error": {"type": "rate_limit_error", "message": "rate limit exceeded"}}`,
			expectedError: "claude: rate limit exceeded or insufficient quota",
		},
		{
			name:          "Internal Server Error 500",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error": {"type": "api_error", "message": "internal error"}}`,
			expectedError: "claude: server error",
		},
		{
			name:          "Bad Request 400",
			statusCode:    http.StatusBadRequest,
			responseBody:  "invalid prompt structure",
			expectedError: "claude API error (400): invalid prompt structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			p := &AnthropicProvider{}
			cfg := &config.Config{
				AIBaseURL: server.URL,
				AIAPIKey:  "test-key",
			}
			if err := p.Configure(cfg); err != nil {
				t.Fatalf("Configure failed: %v", err)
			}

			_, err := p.Send([]ai.Message{{Role: "user", Content: "Hello"}})
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
			}
		})
	}
}
