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
		config          *config.Config
		expectedModel   string
		expectedBaseURL string
		expectedAPIKey  string
	}{
		{
			name:            "Defaults when config fields are empty",
			config:          &config.Config{},
			expectedModel:   "HuggingFaceH4/zephyr-7b-beta",
			expectedBaseURL: "https://router.huggingface.co/models/HuggingFaceH4/zephyr-7b-beta",
			expectedAPIKey:  "",
		},
		{
			name: "Custom AIModel and AIAPIKey set",
			config: &config.Config{
				AIModel:  "meta-llama/Llama-2-7b-chat-hf",
				AIAPIKey: "api-key-123",
			},
			expectedModel:   "meta-llama/Llama-2-7b-chat-hf",
			expectedBaseURL: "https://router.huggingface.co/models/meta-llama/Llama-2-7b-chat-hf",
			expectedAPIKey:  "api-key-123",
		},
		{
			name: "Custom AIBaseURL overrides formatted BaseURL",
			config: &config.Config{
				AIModel:   "custom-model",
				AIBaseURL: "https://custom.huggingface.router/v1",
			},
			expectedModel:   "custom-model",
			expectedBaseURL: "https://custom.huggingface.router/v1",
			expectedAPIKey:  "",
		},
		{
			name: "HFAccessToken overrides AIAPIKey",
			config: &config.Config{
				AIAPIKey:      "generic-key",
				HFAccessToken: "hf_specific_token",
			},
			expectedModel:   "HuggingFaceH4/zephyr-7b-beta",
			expectedBaseURL: "https://router.huggingface.co/models/HuggingFaceH4/zephyr-7b-beta",
			expectedAPIKey:  "hf_specific_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &HFProvider{}
			err := p.Configure(tt.config)
			if err != nil {
				t.Fatalf("unexpected error configuring HFProvider: %v", err)
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

func TestHFProvider_NameAndIsLocal(t *testing.T) {
	p := &HFProvider{}
	if p.Name() != "Hugging Face API" {
		t.Errorf("expected Name() = 'Hugging Face API', got %q", p.Name())
	}

	if p.IsLocal() != false {
		t.Errorf("expected IsLocal() = false, got true")
	}
}

func TestHFProvider_Send_Success(t *testing.T) {
	mockResponse := []hfResponseItem{
		{GeneratedText: "Hello from Hugging Face!"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		if auth := r.Header.Get("Authorization"); auth != "Bearer hf_test_token" {
			t.Errorf("expected Authorization header 'Bearer hf_test_token', got %q", auth)
		}

		var req hfRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		expectedInput := "[INST] Hello [/INST] Hi there! [INST] How are you? [/INST] "
		if req.Inputs != expectedInput {
			t.Errorf("expected prompt input %q, got %q", expectedInput, req.Inputs)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	p := &HFProvider{}
	cfg := &config.Config{
		AIBaseURL:     server.URL,
		HFAccessToken: "hf_test_token",
	}
	if err := p.Configure(cfg); err != nil {
		t.Fatalf("failed to configure provider: %v", err)
	}

	messages := []ai.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	resp, err := p.Send(messages)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp != "Hello from Hugging Face!" {
		t.Errorf("expected response 'Hello from Hugging Face!', got %q", resp)
	}
}

func TestHFProvider_Send_Errors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "Non-200 Status Code Error",
			statusCode:    http.StatusBadRequest,
			responseBody:  "Model too busy",
			expectedError: "HF API error (400)",
		},
		{
			name:          "Invalid JSON Response",
			statusCode:    http.StatusOK,
			responseBody:  `{invalid json}`,
			expectedError: "failed to decode response",
		},
		{
			name:          "Empty Response Array",
			statusCode:    http.StatusOK,
			responseBody:  `[]`,
			expectedError: "empty response from HF API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			p := &HFProvider{}
			cfg := &config.Config{
				AIBaseURL: server.URL,
			}
			if err := p.Configure(cfg); err != nil {
				t.Fatalf("failed to configure provider: %v", err)
			}

			messages := []ai.Message{{Role: "user", Content: "Test"}}
			_, err := p.Send(messages)

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.expectedError)
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}
