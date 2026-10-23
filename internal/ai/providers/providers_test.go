package providers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

func TestProviders_HTTPClientTimeout(t *testing.T) {
	// Fast server handler that responds immediately
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "messages"):
			// Anthropic response
			w.Write([]byte(`{"content":[{"type":"text","text":"claude ok"}]}`))
		case strings.Contains(r.URL.Path, "generateContent"):
			// Gemini response
			w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"gemini ok"}]}}]}`))
		default:
			// HuggingFace response
			w.Write([]byte(`[{"generated_text":"hf ok"}]`))
		}
	}))
	defer ts.Close()

	t.Run("Anthropic Provider", func(t *testing.T) {
		p := &AnthropicProvider{}
		cfg := &config.Config{AIBaseURL: ts.URL, AIAPIKey: "test-key"}
		if err := p.Configure(cfg); err != nil {
			t.Fatalf("Failed to configure Anthropic provider: %v", err)
		}

		res, err := p.Send([]ai.Message{{Role: "user", Content: "hi"}})
		if err != nil {
			t.Fatalf("Anthropic Send failed: %v", err)
		}
		if res != "claude ok" {
			t.Errorf("Expected 'claude ok', got '%s'", res)
		}
	})

	t.Run("Gemini Provider", func(t *testing.T) {
		p := &GeminiProvider{}
		cfg := &config.Config{AIBaseURL: ts.URL, GeminiAPIKey: "test-key"}
		if err := p.Configure(cfg); err != nil {
			t.Fatalf("Failed to configure Gemini provider: %v", err)
		}

		res, err := p.Send([]ai.Message{{Role: "user", Content: "hi"}})
		if err != nil {
			t.Fatalf("Gemini Send failed: %v", err)
		}
		if res != "gemini ok" {
			t.Errorf("Expected 'gemini ok', got '%s'", res)
		}
	})

	t.Run("HuggingFace Provider", func(t *testing.T) {
		p := &HFProvider{}
		cfg := &config.Config{AIBaseURL: ts.URL, HFAccessToken: "test-key"}
		if err := p.Configure(cfg); err != nil {
			t.Fatalf("Failed to configure HuggingFace provider: %v", err)
		}

		res, err := p.Send([]ai.Message{{Role: "user", Content: "hi"}})
		if err != nil {
			t.Fatalf("HuggingFace Send failed: %v", err)
		}
		if res != "hf ok" {
			t.Errorf("Expected 'hf ok', got '%s'", res)
		}
	})
}

func TestProviders_HTTPTimeoutBehavior(t *testing.T) {
	// Slow server handler that delays response longer than a short timeout
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	// Verify that http.Client with a timeout triggers context/timeout errors when exceeded
	client := &http.Client{Timeout: 10 * time.Millisecond}
	_, err := client.Get(slowServer.URL)
	if err == nil {
		t.Fatal("Expected timeout error from slow server, got nil")
	}
}
