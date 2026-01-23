package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

type OllamaProvider struct {
	BaseURL    string
	modelName  string
	httpClient *http.Client
}

func (p *OllamaProvider) Name() string {
	return "Ollama"
}

func (p *OllamaProvider) Model() string {
	return p.modelName
}

func (p *OllamaProvider) Configure(cfg *config.Config) error {
	p.BaseURL = "http://localhost:11434"
	if cfg.AIBaseURL != "" {
		p.BaseURL = cfg.AIBaseURL
	}
	p.modelName = "mistral" // Default
	if cfg.AIModel != "" {
		p.modelName = cfg.AIModel
	}

	// Reuse client with reasonable timeout
	p.httpClient = &http.Client{
		Timeout: 90 * time.Second,
	}

	return nil
}

func (p *OllamaProvider) IsLocal() bool {
	return true
}

type ollamaRequest struct {
	Model    string                 `json:"model"`
	Messages []ai.Message           `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ollamaResponse struct {
	Message ai.Message `json:"message"`
	Done    bool       `json:"done"`
}

func (p *OllamaProvider) Send(messages []ai.Message) (string, error) {
	reqBody := ollamaRequest{
		Model:    p.modelName,
		Messages: messages,
		Stream:   false,
		Options: map[string]interface{}{
			"num_predict": 512,  // Limit response length for faster generation
			"temperature": 0.7,  // Balanced creativity/speed
			"top_p":       0.9,  // Nucleus sampling for better quality
			"num_ctx":     2048, // Context window
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	if p.httpClient == nil {
		p.httpClient = &http.Client{Timeout: 90 * time.Second}
	}

	resp, err := p.httpClient.Post(p.BaseURL+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Ollama: Connection failed. Is Ollama running at %s?", p.BaseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("Ollama: Model '%s' not found. Have you run 'ollama pull %s'?", p.modelName, p.modelName)
		}
		return "", fmt.Errorf("Ollama API error (%d): %s", resp.StatusCode, string(body))
	}

	var parsedResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return parsedResp.Message.Content, nil
}
