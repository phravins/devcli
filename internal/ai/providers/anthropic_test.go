package providers

import (
	"testing"

	"github.com/phravins/devcli/internal/config"
)

func TestAnthropicPrefersDedicatedKey(t *testing.T) {
	p := &AnthropicProvider{}
	if err := p.Configure(&config.Config{AIAPIKey: "generic", AnthropicAPIKey: "dedicated"}); err != nil {
		t.Fatal(err)
	}
	if p.APIKey != "dedicated" {
		t.Fatalf("got %q, want anthropic_api_key to take precedence", p.APIKey)
	}

	p = &AnthropicProvider{}
	if err := p.Configure(&config.Config{AIAPIKey: "generic"}); err != nil {
		t.Fatal(err)
	}
	if p.APIKey != "generic" {
		t.Fatalf("got %q, want fallback to ai_api_key", p.APIKey)
	}
}
