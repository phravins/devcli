package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

type AnthropicProvider struct {
	BaseURL   string
	APIKey    string
	modelName string
}

func (p *AnthropicProvider) Name() string {
	return "Anthropic (Claude)"
}

func (p *AnthropicProvider) Model() string {
	return p.modelName
}

func (p *AnthropicProvider) Configure(cfg *config.Config) error {
	p.BaseURL = "https://api.anthropic.com/v1"
	p.modelName = "claude-3-opus-20240229"
	p.APIKey = cfg.AIAPIKey

	if cfg.AIModel != "" {
		p.modelName = cfg.AIModel
	}
	if cfg.AIBaseURL != "" {
		p.BaseURL = cfg.AIBaseURL
	}
	return nil
}

func (p *AnthropicProvider) IsLocal() bool {
	return false
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *AnthropicProvider) Send(messages []ai.Message) (string, error) {
	var apiMessages []anthropicMessage
	for _, m := range messages {
		if m.Role == "system" {
			continue
		}
		apiMessages = append(apiMessages, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	reqBody := anthropicRequest{
		Model:     p.modelName,
		Messages:  apiMessages,
		MaxTokens: 1024,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", p.BaseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude API connection failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return "", fmt.Errorf("Claude: Invalid API Key or access denied. Please check your configuration.")
		case http.StatusNotFound:
			return "", fmt.Errorf("Claude: Model '%s' not found or you don't have access to it.", p.modelName)
		case http.StatusTooManyRequests:
			return "", fmt.Errorf("Claude: Rate limit exceeded or insufficient quota.")
		case http.StatusInternalServerError:
			return "", fmt.Errorf("Claude: Server error. Please try again later.")
		default:
			return "", fmt.Errorf("Claude API error (%d): %s", resp.StatusCode, string(body))
		}
	}

	var parsedResp anthropicResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&parsedResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if parsedResp.Error != nil {
		return "", fmt.Errorf("claude API Error: %s", parsedResp.Error.Message)
	}

	if len(parsedResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude API")
	}

	return parsedResp.Content[0].Text, nil
}
