package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

type GeminiProvider struct {
	BaseURL   string
	APIKey    string
	modelName string
}

func (p *GeminiProvider) Name() string {
	return "Google Gemini"
}

func (p *GeminiProvider) Model() string {
	return p.modelName
}

func (p *GeminiProvider) Configure(cfg *config.Config) error {
	p.modelName = "gemini-1.5-flash-001"
	if cfg.AIModel != "" {
		if cfg.AIModel == "gemini" {
			p.modelName = "gemini-1.5-flash-001"
		} else {
			p.modelName = cfg.AIModel
		}
	}
	p.APIKey = cfg.AIAPIKey
	if cfg.GeminiAPIKey != "" {
		p.APIKey = cfg.GeminiAPIKey
	}
	if cfg.AIBaseURL != "" && (strings.HasPrefix(cfg.AIBaseURL, "http://") || strings.HasPrefix(cfg.AIBaseURL, "https://")) {
		p.BaseURL = cfg.AIBaseURL
	} else {
		p.BaseURL = "https://generativelanguage.googleapis.com/v1beta/models"
	}
	return nil
}

func (p *GeminiProvider) IsLocal() bool {
	return false
}

type geminiPart struct {
	Text string `json:"text"`
}
type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (p *GeminiProvider) Send(messages []ai.Message) (string, error) {
	var geminiMsgs []geminiContent

	// Gemini roles: "user", "model"
	for _, m := range messages {
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		geminiMsgs = append(geminiMsgs, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	reqBody := geminiRequest{Contents: geminiMsgs}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", p.BaseURL, p.modelName, p.APIKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini API connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		safeURL := fmt.Sprintf("%s/%s:generateContent", p.BaseURL, p.modelName)
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return "", fmt.Errorf("Gemini: Invalid API Key or access denied. Please check your configuration.")
		case http.StatusNotFound:
			return "", fmt.Errorf("Gemini: Model '%s' not found or API endpoint is incorrect.", p.modelName)
		case http.StatusTooManyRequests:
			return "", fmt.Errorf("Gemini: Rate limit exceeded or insufficient quota.")
		case http.StatusInternalServerError:
			return "", fmt.Errorf("Gemini: Server error. Please try again later.")
		default:
			return "", fmt.Errorf("Gemini API error (%d) at %s: %s", resp.StatusCode, safeURL, string(body))
		}
	}
	var parsedResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if len(parsedResp.Candidates) == 0 || len(parsedResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini API")
	}
	return parsedResp.Candidates[0].Content.Parts[0].Text, nil
}
