package devtools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type APIRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type APIResponse struct {
	StatusCode int                 `json:"status_code"`
	Status     string              `json:"status"`
	LatencyMs  int64               `json:"latency_ms"`
	Headers    http.Header         `json:"headers"`
	Body       string              `json:"body"`
	Formatted  string              `json:"formatted"`
	Err        error               `json:"err"`
}

func ExecuteAPIRequest(req APIRequest) APIResponse {
	start := time.Now()

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "GET"
	}

	urlStr := strings.TrimSpace(req.URL)
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return APIResponse{Err: fmt.Errorf("invalid request: %w", err), LatencyMs: time.Since(start).Milliseconds()}
	}

	// Add headers
	for k, v := range req.Headers {
		if strings.TrimSpace(k) != "" {
			httpReq.Header.Set(k, v)
		}
	}
	if req.Body != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return APIResponse{Err: fmt.Errorf("request failed: %w", err), LatencyMs: latency}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return APIResponse{StatusCode: resp.StatusCode, Status: resp.Status, LatencyMs: latency, Headers: resp.Header, Err: err}
	}

	bodyStr := string(respBody)
	formattedStr := bodyStr

	// Try formatting JSON if response is JSON
	var jsonObj interface{}
	if err := json.Unmarshal(respBody, &jsonObj); err == nil {
		if pretty, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
			formattedStr = string(pretty)
		}
	}

	return APIResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		LatencyMs:  latency,
		Headers:    resp.Header,
		Body:       bodyStr,
		Formatted:  formattedStr,
	}
}
