package providers

import (
	"fmt"
	"strings"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

// GetProvider returns an AI provider based on the configuration
func GetProvider(cfg *config.Config) (ai.Provider, error) {
	backend := strings.TrimSpace(strings.ToLower(cfg.AIBackend))
	if backend == "" {
		backend = "ollama"
	}

	var p ai.Provider

	switch backend {
	case "ollama":
		p = &OllamaProvider{}
	case "huggingface":
		p = &HFProvider{}
	case "local":
		p = &LocalHFProvider{}
	case "claude", "anthropic":
		p = &AnthropicProvider{}
	case "gemini", "google":
		p = &GeminiProvider{}

	// Pre-configured OpenAI Compatible Shortcuts
	case "mistral":
		p = &OpenAIProvider{BaseURL: "https://api.mistral.ai/v1"}
	case "kimi", "moonshot":
		p = &OpenAIProvider{BaseURL: "https://api.moonshot.cn/v1"}
	case "groq":
		p = &OpenAIProvider{BaseURL: "https://api.groq.com/openai/v1"}
	case "deepseek":
		p = &OpenAIProvider{BaseURL: "https://api.deepseek.com/v1"}
	case "lmstudio":
		p = &OpenAIProvider{}

	default:
		// Catch-all: Assume "Generic OpenAI Compatible" for any other name
		p = &OpenAIProvider{}
	}

	if err := p.Configure(cfg); err != nil {
		return nil, fmt.Errorf("failed to configure provider %s: %w", backend, err)
	}

	return p, nil
}
