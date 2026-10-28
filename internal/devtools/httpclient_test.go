package devtools

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteAPIRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/large" {
			// Write 11MB of data
			largeData := bytes.Repeat([]byte("A"), 11<<20)
			w.WriteHeader(http.StatusOK)
			w.Write(largeData)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer ts.Close()

	t.Run("Normal Response", func(t *testing.T) {
		req := APIRequest{
			Method: "GET",
			URL:    ts.URL + "/normal",
		}
		resp := ExecuteAPIRequest(req)
		if resp.Err != nil {
			t.Fatalf("unexpected error: %v", resp.Err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		if resp.Body != `{"status": "ok"}` {
			t.Errorf("expected body %q, got %q", `{"status": "ok"}`, resp.Body)
		}
	})

	t.Run("Oversized Response Truncation", func(t *testing.T) {
		req := APIRequest{
			Method: "GET",
			URL:    ts.URL + "/large",
		}
		resp := ExecuteAPIRequest(req)
		if resp.Err != nil {
			t.Fatalf("unexpected error: %v", resp.Err)
		}
		if len(resp.Body) != 10<<20 {
			t.Errorf("expected body size limited to %d bytes, got %d", 10<<20, len(resp.Body))
		}
	})
}
