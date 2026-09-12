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

func TestHFProvider_Configure(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.Config
		expectedModel   string
		expectedBaseURL string
		expectedAPIKey  string
	}{
		{
			name:            "default configuration",
			cfg:             &config.Config{},
			expectedModel:   "HuggingFaceH4/zephyr-7b-beta",
			expectedBaseURL: "https://router.huggingface.co/models/HuggingFaceH4/zephyr-7b-beta",
			expectedAPIKey:  "",
		},
		{
			name: "custom model and AI API key",
			cfg: &config.Config{
				AIModel:  "meta-llama/Llama-2-7b-chat-hf",
				AIAPIKey: "ai-key-123",
			},
			expectedModel:   "meta-llama/Llama-2-7b-chat-hf",
			expectedBaseURL: "https://router.huggingface.co/models/meta-llama/Llama-2-7b-chat-hf",
			expectedAPIKey:  "ai-key-123",
		},
		{
			name: "custom base URL and HF token override",
			cfg: &config.Config{
				AIModel:       "custom/model",
				AIBaseURL:     "https://custom-proxy.example.com",
				AIAPIKey:      "ai-key-123",
				HFAccessToken: "hf-token-456",
			},
			expectedModel:   "custom/model",
			expectedBaseURL: "https://custom-proxy.example.com",
			expectedAPIKey:  "hf-token-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &HFProvider{}
			err := p.Configure(tt.cfg)
			if err != nil {
				t.Fatalf("unexpected error during Configure: %v", err)
			}

			if p.Model() != tt.expectedModel {
				t.Errorf("expected Model() = %q, got %q", tt.expectedModel, p.Model())
			}
			if p.BaseURL != tt.expectedBaseURL {
				t.Errorf("expected BaseURL = %q, got %q", tt.expectedBaseURL, p.BaseURL)
			}
			if p.APIKey != tt.expectedAPIKey {
				t.Errorf("expected APIKey = %q, got %q", tt.expectedAPIKey, p.APIKey)
			}
		})
	}
}

func TestHFProvider_Metadata(t *testing.T) {
	p := &HFProvider{}
	if p.Name() != "Hugging Face API" {
		t.Errorf("expected Name() = 'Hugging Face API', got %q", p.Name())
	}
	if p.IsLocal() != false {
		t.Errorf("expected IsLocal() = false, got %v", p.IsLocal())
	}
}

func TestHFProvider_Send_Success(t *testing.T) {
	mockResp := []hfResponseItem{
		{GeneratedText: "Hello from Hugging Face!"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if auth := r.Header.Get("Authorization"); auth != "Bearer hf_test_token" {
			t.Errorf("expected Authorization header 'Bearer hf_test_token', got %q", auth)
		}

		var req hfRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		expectedPrompt := "[INST] Hello [/INST] Hi there! [INST] How are you? [/INST] "
		if req.Inputs != expectedPrompt {
			t.Errorf("expected inputs %q, got %q", expectedPrompt, req.Inputs)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	p := &HFProvider{}
	cfg := &config.Config{
		AIBaseURL:     server.URL,
		HFAccessToken: "hf_test_token",
	}
	if err := p.Configure(cfg); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	messages := []ai.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	resp, err := p.Send(messages)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if resp != "Hello from Hugging Face!" {
		t.Errorf("expected response 'Hello from Hugging Face!', got %q", resp)
	}
}

func TestHFProvider_Send_Errors(t *testing.T) {
	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Model rate limit exceeded"))
		}))
		defer server.Close()

		p := &HFProvider{}
		cfg := &config.Config{
			AIBaseURL: server.URL,
		}
		_ = p.Configure(cfg)

		_, err := p.Send([]ai.Message{{Role: "user", Content: "hi"}})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "HF API error (400)") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("invalid json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		p := &HFProvider{}
		cfg := &config.Config{
			AIBaseURL: server.URL,
		}
		_ = p.Configure(cfg)

		_, err := p.Send([]ai.Message{{Role: "user", Content: "hi"}})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to decode response") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("empty response array", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
		}))
		defer server.Close()

		p := &HFProvider{}
		cfg := &config.Config{
			AIBaseURL: server.URL,
		}
		_ = p.Configure(cfg)

		_, err := p.Send([]ai.Message{{Role: "user", Content: "hi"}})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "empty response from HF API") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
