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

func TestAnthropicProvider_Metadata(t *testing.T) {
	p := &AnthropicProvider{}

	if p.Name() != "Anthropic (Claude)" {
		t.Errorf("Expected Name 'Anthropic (Claude)', got '%s'", p.Name())
	}

	if p.IsLocal() != false {
		t.Errorf("Expected IsLocal to be false")
	}
}

func TestAnthropicProvider_Configure(t *testing.T) {
	p := &AnthropicProvider{}
	cfg := &config.Config{
		AIModel:  "claude-3-haiku-20240307",
		AIAPIKey: "test-anthropic-key",
	}

	if err := p.Configure(cfg); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if p.Model() != "claude-3-haiku-20240307" {
		t.Errorf("Expected model 'claude-3-haiku-20240307', got '%s'", p.Model())
	}
	if p.APIKey != "test-anthropic-key" {
		t.Errorf("Expected APIKey 'test-anthropic-key', got '%s'", p.APIKey)
	}
}

func TestAnthropicProvider_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("Expected x-api-key header 'test-key', got '%s'", r.Header.Get("x-api-key"))
		}

		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
				Type string `json:"type"`
			}{
				{Text: "Hello from Claude!", Type: "text"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &AnthropicProvider{}
	p.Configure(&config.Config{
		AIBaseURL: server.URL,
		AIAPIKey:  "test-key",
	})

	messages := []ai.Message{
		{Role: "user", Content: "Hello"},
	}

	resp, err := p.Send(messages)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if resp != "Hello from Claude!" {
		t.Errorf("Expected 'Hello from Claude!', got '%s'", resp)
	}
}

func TestAnthropicProvider_Send_ErrorStatuses(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError string
	}{
		{"Unauthorized", http.StatusUnauthorized, "claude: invalid API Key or access denied"},
		{"Not Found", http.StatusNotFound, "claude: model 'claude-3-opus-20240229' not found or no access"},
		{"Rate Limit", http.StatusTooManyRequests, "claude: rate limit exceeded or insufficient quota"},
		{"Server Error", http.StatusInternalServerError, "claude: server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			p := &AnthropicProvider{}
			p.Configure(&config.Config{AIBaseURL: server.URL})

			_, err := p.Send([]ai.Message{{Role: "user", Content: "Hi"}})
			if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
			}
		})
	}
}
