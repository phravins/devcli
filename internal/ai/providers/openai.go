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

type OpenAIProvider struct {
	BaseURL    string
	APIKey     string
	modelName  string
	IsLMStudio bool
	httpClient *http.Client
}

func (p *OpenAIProvider) Name() string {
	if p.IsLMStudio {
		return "LM Studio"
	}
	return "OpenAI"
}

func (p *OpenAIProvider) Model() string {
	return p.modelName
}

func (p *OpenAIProvider) Configure(cfg *config.Config) error {
	// Defaults for OpenAI (only if not already set by Factory)
	if p.BaseURL == "" {
		p.BaseURL = "https://api.openai.com/v1"
	}
	p.modelName = "gpt-3.5-turbo"
	p.APIKey = cfg.AIAPIKey

	if cfg.AIBackend == "lmstudio" {
		p.IsLMStudio = true
		p.BaseURL = "http://localhost:1234/v1" // LM Studio default
		p.modelName = "local-model"
	}

	// Overrides
	if cfg.AIBaseURL != "" {
		p.BaseURL = cfg.AIBaseURL
	}
	if cfg.AIModel != "" {
		p.modelName = cfg.AIModel
	}

	p.httpClient = &http.Client{
		Timeout: 90 * time.Second, // Global timeout
	}

	return nil
}

func (p *OpenAIProvider) IsLocal() bool {
	return p.IsLMStudio
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (p *OpenAIProvider) Send(messages []ai.Message) (string, error) {
	// Convert internal messages to OpenAI struct
	var apiMessages []openAIMessage
	for _, m := range messages {
		apiMessages = append(apiMessages, openAIMessage{Role: m.Role, Content: m.Content})
	}

	reqBody := openAIRequest{
		Model:    p.modelName,
		Messages: apiMessages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", p.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if !p.IsLMStudio {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	if p.httpClient == nil {
		p.httpClient = &http.Client{}
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp openAIErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			switch resp.StatusCode {
			case http.StatusUnauthorized:
				return "", fmt.Errorf("OpenAI: Invalid API Key. Please check your configuration.")
			case http.StatusNotFound:
				return "", fmt.Errorf("OpenAI: Model '%s' not found or you don't have access to it.", p.modelName)
			case http.StatusTooManyRequests:
				return "", fmt.Errorf("OpenAI: Rate limit exceeded or insufficient quota.")
			case http.StatusInternalServerError:
				return "", fmt.Errorf("OpenAI: Server error. Please try again later.")
			default:
				return "", fmt.Errorf("OpenAI error (%d): %s", resp.StatusCode, errResp.Error.Message)
			}
		}
		// Fallback for non-JSON or unexpected error format
		return "", fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(body))
	}

	var parsedResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(parsedResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return parsedResp.Choices[0].Message.Content, nil
}
