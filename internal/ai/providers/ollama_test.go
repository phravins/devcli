package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

func TestOllamaProvider_Send(t *testing.T) {
	// Mock Ollama Server
	mockResponse := ollamaResponse{
		Message: ai.Message{
			Role:    "assistant",
			Content: "Hello from Ollama!",
		},
		Done: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request
		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected path /api/chat, got %s", r.URL.Path)
		}

		var req ollamaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "test-model" {
			t.Errorf("Expected model 'test-model', got '%s'", req.Model)
		}

		// Send Response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Configure Provider
	p := &OllamaProvider{}
	cfg := &config.Config{
		AIBaseURL: server.URL,
		AIModel:   "test-model",
	}
	if err := p.Configure(cfg); err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	// Test Send
	messages := []ai.Message{{Role: "user", Content: "Hello"}}
	resp, err := p.Send(messages)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if resp != "Hello from Ollama!" {
		t.Errorf("Expected 'Hello from Ollama!', got '%s'", resp)
	}
}
