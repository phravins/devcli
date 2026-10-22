package devtools_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phravins/devcli/internal/devtools"
)

func TestExecuteAPIRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	req := devtools.APIRequest{
		Method: "GET",
		URL:    server.URL,
	}

	resp := devtools.ExecuteAPIRequest(req)
	if resp.Err != nil {
		t.Fatalf("expected no error, got: %v", resp.Err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if resp.Body != `{"status": "ok"}` {
		t.Errorf("expected body '{\"status\": \"ok\"}', got '%s'", resp.Body)
	}
}

func TestExecuteAPIRequest_ResponseBodyLimit(t *testing.T) {
	// Generate data larger than 10MB (e.g. 11MB)
	largeDataSize := 11 * 1024 * 1024
	largeData := bytes.Repeat([]byte("A"), largeDataSize)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(largeData)
	}))
	defer server.Close()

	req := devtools.APIRequest{
		Method: "GET",
		URL:    server.URL,
	}

	resp := devtools.ExecuteAPIRequest(req)
	if resp.Err != nil {
		t.Fatalf("expected no error, got: %v", resp.Err)
	}

	maxAllowedSize := 10 * 1024 * 1024
	if len(resp.Body) != maxAllowedSize {
		t.Errorf("expected body truncated to %d bytes, got %d bytes", maxAllowedSize, len(resp.Body))
	}
}
