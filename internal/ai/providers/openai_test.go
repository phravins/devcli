package providers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/phravins/devcli/internal/ai"
	"github.com/phravins/devcli/internal/config"
)

func TestOpenAIProvider_Send_Success(t *testing.T) {
	mockResponse := openAIResponse{
		Choices: []struct {
			Message openAIMessage `json:"message"`
		}{
			{
				Message: openAIMessage{
					Role:    "assistant",
					Content: "Hello from OpenAI!",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		var req openAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected model 'gpt-3.5-turbo', got '%s'", req.Model)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	p := &OpenAIProvider{}
	cfg := &config.Config{
		AIBaseURL: server.URL,
		AIAPIKey:  "test-key",
	}
	if err := p.Configure(cfg); err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	messages := []ai.Message{{Role: "user", Content: "Hello"}}
	resp, err := p.Send(messages)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if resp != "Hello from OpenAI!" {
		t.Errorf("Expected 'Hello from OpenAI!', got '%s'", resp)
	}
}

func TestOpenAIProvider_Send_Errors(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  string
	}{
		{
			name:       "Unauthorized 401",
			statusCode: http.StatusUnauthorized,
			responseBody: `{"error": {"message": "Invalid API key provided."}}`,
			expectedError: "openAI: invalid API key",
		},
		{
			name:       "Not Found 404",
			statusCode: http.StatusNotFound,
			responseBody: `{"error": {"message": "The model 'gpt-3.5-turbo' does not exist"}}`,
			expectedError: "openAI: model 'gpt-3.5-turbo' not found or no access",
		},
		{
			name:       "Too Many Requests 429",
			statusCode: http.StatusTooManyRequests,
			responseBody: `{"error": {"message": "Rate limit reached for requests"}}`,
			expectedError: "openAI: rate limit exceeded or insufficient quota",
		},
		{
			name:       "Internal Server Error 500",
			statusCode: http.StatusInternalServerError,
			responseBody: `{"error": {"message": "The server had an error while processing your request"}}`,
			expectedError: "openAI: server error",
		},
		{
			name:       "Other Error 400 with JSON",
			statusCode: http.StatusBadRequest,
			responseBody: `{"error": {"message": "Bad request format"}}`,
			expectedError: "OpenAI error (400): Bad request format",
		},
		{
			name:       "Fallback Non-JSON Error 502",
			statusCode: http.StatusBadGateway,
			responseBody: "Bad Gateway",
			expectedError: "OpenAI API error (502): Bad Gateway",
		},
		{
			name:       "Empty Error Message JSON",
			statusCode: http.StatusForbidden,
			responseBody: `{"error": {"message": ""}}`,
			expectedError: "OpenAI API error (403): {\"error\": {\"message\": \"\"}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			p := &OpenAIProvider{}
			cfg := &config.Config{
				AIBaseURL: server.URL,
				AIAPIKey:  "test-key",
			}
			if err := p.Configure(cfg); err != nil {
				t.Fatalf("Failed to configure provider: %v", err)
			}

			messages := []ai.Message{{Role: "user", Content: "Hello"}}
			_, err := p.Send(messages)

			if err == nil {
				t.Fatalf("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error to contain '%s', got '%v'", tt.expectedError, err)
			}
		})
	}
}
