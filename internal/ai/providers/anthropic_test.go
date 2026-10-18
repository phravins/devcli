package providers

import (
	"testing"
	"time"

	"github.com/phravins/devcli/internal/config"
)

func TestAnthropicProvider_Configure(t *testing.T) {
	p := &AnthropicProvider{}
	cfg := &config.Config{
		AIAPIKey: "test-anthropic-key",
	}

	if err := p.Configure(cfg); err != nil {
		t.Fatalf("unexpected error configuring AnthropicProvider: %v", err)
	}

	if p.Name() != "Anthropic (Claude)" {
		t.Errorf("Expected Name 'Anthropic (Claude)', got '%s'", p.Name())
	}

	if p.IsLocal() != false {
		t.Errorf("Expected IsLocal to be false")
	}

	if p.httpClient == nil {
		t.Fatalf("Expected httpClient to be initialized, got nil")
	}

	if p.httpClient.Timeout != 90*time.Second {
		t.Errorf("Expected httpClient timeout 90s, got %v", p.httpClient.Timeout)
	}
}
