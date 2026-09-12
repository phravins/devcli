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

func TestGeminiProvider_Metadata(t *testing.T) {
	p := &GeminiProvider{}

	if p.Name() != "Google Gemini" {
		t.Errorf("Expected Name 'Google Gemini', got '%s'", p.Name())
	}

	if p.IsLocal() != false {
		t.Errorf("Expected IsLocal to be false")
	}
}

func TestGeminiProvider_Configure(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.Config
		expectedModel   string
		expectedKey     string
		expectedBaseURL string
	}{
		{
			name:            "Default values",
			cfg:             &config.Config{},
			expectedModel:   "gemini-1.5-flash-001",
			expectedKey:     "",
			expectedBaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		},
		{
			name: "Explicit model 'gemini'",
			cfg: &config.Config{
				AIModel: "gemini",
			},
			expectedModel:   "gemini-1.5-flash-001",
			expectedKey:     "",
			expectedBaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		},
		{
			name: "Custom model and generic AIAPIKey",
			cfg: &config.Config{
				AIModel:  "gemini-1.5-pro",
				AIAPIKey: "generic-key",
			},
			expectedModel:   "gemini-1.5-pro",
			expectedKey:     "generic-key",
			expectedBaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		},
		{
			name: "GeminiAPIKey overrides AIAPIKey",
			cfg: &config.Config{
				AIAPIKey:     "generic-key",
				GeminiAPIKey: "gemini-specific-key",
			},
			expectedModel:   "gemini-1.5-flash-001",
			expectedKey:     "gemini-specific-key",
			expectedBaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		},
		{
			name: "Valid HTTP BaseURL",
			cfg: &config.Config{
				AIBaseURL: "http://localhost:8080/v1",
			},
			expectedModel:   "gemini-1.5-flash-001",
			expectedKey:     "",
			expectedBaseURL: "http://localhost:8080/v1",
		},
		{
			name: "Valid HTTPS BaseURL",
			cfg: &config.Config{
				AIBaseURL: "https://custom.api.com/v1",
			},
			expectedModel:   "gemini-1.5-flash-001",
			expectedKey:     "",
			expectedBaseURL: "https://custom.api.com/v1",
		},
		{
			name: "Invalid BaseURL prefix uses default BaseURL",
			cfg: &config.Config{
				AIBaseURL: "invalid-url-scheme",
			},
			expectedModel:   "gemini-1.5-flash-001",
			expectedKey:     "",
			expectedBaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &GeminiProvider{}
			err := p.Configure(tt.cfg)
			if err != nil {
				t.Fatalf("Configure failed: %v", err)
			}

			if p.Model() != tt.expectedModel {
				t.Errorf("Expected model '%s', got '%s'", tt.expectedModel, p.Model())
			}
			if p.APIKey != tt.expectedKey {
				t.Errorf("Expected APIKey '%s', got '%s'", tt.expectedKey, p.APIKey)
			}
			if p.BaseURL != tt.expectedBaseURL {
				t.Errorf("Expected BaseURL '%s', got '%s'", tt.expectedBaseURL, p.BaseURL)
			}
		})
	}
}

func TestGeminiProvider_Send_Success(t *testing.T) {
	mockResponse := geminiResponse{
		Candidates: []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		}{
			{
				Content: struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				}{
					Parts: []struct {
						Text string `json:"text"`
					}{
						{Text: "Hello from Gemini!"},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gemini-1.5-flash-001:generateContent" {
			t.Errorf("Expected path /gemini-1.5-flash-001:generateContent, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("key") != "test-key" {
			t.Errorf("Expected key 'test-key', got '%s'", r.URL.Query().Get("key"))
		}

		var req geminiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if len(req.Contents) != 2 {
			t.Fatalf("Expected 2 contents, got %d", len(req.Contents))
		}
		if req.Contents[0].Role != "user" || req.Contents[0].Parts[0].Text != "Hi" {
			t.Errorf("Unexpected content[0]: %+v", req.Contents[0])
		}
		if req.Contents[1].Role != "model" || req.Contents[1].Parts[0].Text != "Hello human" {
			t.Errorf("Unexpected content[1]: %+v", req.Contents[1])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	p := &GeminiProvider{}
	cfg := &config.Config{
		AIBaseURL:    server.URL,
		GeminiAPIKey: "test-key",
	}
	if err := p.Configure(cfg); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	messages := []ai.Message{
		{Role: "user", Content: "Hi"},
		{Role: "assistant", Content: "Hello human"},
	}

	resp, err := p.Send(messages)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if resp != "Hello from Gemini!" {
		t.Errorf("Expected 'Hello from Gemini!', got '%s'", resp)
	}
}

func TestGeminiProvider_Send_Errors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "Unauthorized 401",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"error": "unauthorized"}`,
			expectedError: "gemini: invalid API Key or access denied",
		},
		{
			name:          "Forbidden 403",
			statusCode:    http.StatusForbidden,
			responseBody:  `{"error": "forbidden"}`,
			expectedError: "gemini: invalid API Key or access denied",
		},
		{
			name:          "Not Found 404",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"error": "not found"}`,
			expectedError: "gemini: model 'gemini-1.5-flash-001' not found or API endpoint incorrect",
		},
		{
			name:          "Too Many Requests 429",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `{"error": "rate limit"}`,
			expectedError: "gemini: rate limit exceeded or insufficient quota",
		},
		{
			name:          "Internal Server Error 500",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error": "internal error"}`,
			expectedError: "gemini: server error",
		},
		{
			name:          "Bad Request 400",
			statusCode:    http.StatusBadRequest,
			responseBody:  "invalid prompt",
			expectedError: "gemini API error (400) at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			p := &GeminiProvider{}
			cfg := &config.Config{
				AIBaseURL:    server.URL,
				GeminiAPIKey: "test-key",
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

func TestGeminiProvider_Send_DecodeAndEmptyErrors(t *testing.T) {
	t.Run("Invalid JSON Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		p := &GeminiProvider{}
		p.Configure(&config.Config{AIBaseURL: server.URL})

		_, err := p.Send([]ai.Message{{Role: "user", Content: "Hello"}})
		if err == nil || !strings.Contains(err.Error(), "failed to decode response") {
			t.Errorf("Expected decode error, got '%v'", err)
		}
	})

	t.Run("Empty Candidates Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(geminiResponse{Candidates: nil})
		}))
		defer server.Close()

		p := &GeminiProvider{}
		p.Configure(&config.Config{AIBaseURL: server.URL})

		_, err := p.Send([]ai.Message{{Role: "user", Content: "Hello"}})
		if err == nil || !strings.Contains(err.Error(), "empty response from Gemini API") {
			t.Errorf("Expected empty response error, got '%v'", err)
		}
	})
}
