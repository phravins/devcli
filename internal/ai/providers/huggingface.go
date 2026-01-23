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

type HFProvider struct {
	BaseURL   string
	APIKey    string
	modelName string
}

func (p *HFProvider) Name() string {
	return "Hugging Face API"
}

func (p *HFProvider) Model() string {
	return p.modelName
}

func (p *HFProvider) Configure(cfg *config.Config) error {
	p.modelName = "HuggingFaceH4/zephyr-7b-beta"
	if cfg.AIModel != "" {
		p.modelName = cfg.AIModel
	}
	// URL: https://router.huggingface.co/models/%s
	p.BaseURL = fmt.Sprintf("https://router.huggingface.co/models/%s", p.modelName)
	if cfg.AIBaseURL != "" {
		p.BaseURL = cfg.AIBaseURL
	}

	p.APIKey = cfg.AIAPIKey
	if cfg.HFAccessToken != "" {
		p.APIKey = cfg.HFAccessToken
	}
	return nil
}

func (p *HFProvider) IsLocal() bool {
	return false
}

type hfRequest struct {
	Inputs     string `json:"inputs"`
	Parameters struct {
		MaxNewTokens   int  `json:"max_new_tokens"`
		ReturnFullText bool `json:"return_full_text"`
	} `json:"parameters"`
}

type hfResponseItem struct {
	GeneratedText string `json:"generated_text"`
}

func (p *HFProvider) Send(messages []ai.Message) (string, error) {
	var prompt bytes.Buffer
	for _, m := range messages {
		switch m.Role {
		case "user":
			prompt.WriteString(fmt.Sprintf("[INST] %s [/INST] ", m.Content))
		case "assistant":
			prompt.WriteString(fmt.Sprintf("%s ", m.Content))
		}
	}

	reqBody := hfRequest{
		Inputs: prompt.String(),
	}
	reqBody.Parameters.MaxNewTokens = 500
	reqBody.Parameters.ReturnFullText = false

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", p.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HF API connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HF API error (%d) for model '%s': %s", resp.StatusCode, p.modelName, string(body))
	}

	var parsedResp []hfResponseItem
	if err := json.NewDecoder(resp.Body).Decode(&parsedResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(parsedResp) == 0 {
		return "", fmt.Errorf("empty response from HF API")
	}

	return parsedResp[0].GeneratedText, nil
}
