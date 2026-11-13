package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{"simple", "ls -la", []string{"ls", "-la"}},
		{"single quotes", "echo 'hello world'", []string{"echo", "hello world"}},
		{"double quotes", "echo \"hello world\"", []string{"echo", "hello world"}},
		{"escaped spaces", `cat file\ with\ space.txt`, []string{"cat", "file with space.txt"}},
		{"multiple spaces", "npm   run   dev", []string{"npm", "run", "dev"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommand(tt.command)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseCommand(%q) = %v, want %v", tt.command, got, tt.expected)
			}
		})
	}
}

func TestServerTimeouts(t *testing.T) {
	server := newServer("0", "tok")

	if server.Addr != "127.0.0.1:0" {
		t.Errorf("server must bind to loopback only, got %q", server.Addr)
	}
	if server.ReadHeaderTimeout == 0 {
		t.Error("expected non-zero ReadHeaderTimeout to prevent Slowloris attacks")
	}
	if server.ReadTimeout == 0 {
		t.Error("expected non-zero ReadTimeout")
	}
	if server.WriteTimeout <= execTimeout {
		t.Error("expected WriteTimeout to exceed execTimeout so results can be delivered")
	}
	if server.IdleTimeout == 0 {
		t.Error("expected non-zero IdleTimeout")
	}
}

func TestMaxBytesReaderLimit(t *testing.T) {
	oversizedBody := bytes.Repeat([]byte("a"), (1<<20)+100) // Slightly over 1MB

	endpoints := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"logs endpoint", handleLogs, "/logs"},
		{"save endpoint", handleSave, "/save"},
		{"run endpoint", handleRun, "/run"},
		{"terminal endpoint", handleTerminal, "/terminal"},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", ep.path, bytes.NewReader(oversizedBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			ep.handler(rr, req)

			if rr.Code == http.StatusOK {
				t.Errorf("expected error status for oversized payload on %s, got 200 OK", ep.path)
			}
		})
	}
}

func TestHandleSaveErrorSanitization(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	readOnlyDir := "readonly_dir"
	if err := os.Mkdir(readOnlyDir, 0444); err != nil {
		t.Fatalf("Failed to create read-only dir: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755)

	body, _ := json.Marshal(map[string]string{
		"filename": readOnlyDir + "/file.txt",
		"content":  "test",
	})
	req, err := http.NewRequest("POST", "/save", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handleSave(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %v, got %v", http.StatusInternalServerError, rr.Code)
	}

	resBody := rr.Body.String()
	if bytes.Contains([]byte(resBody), []byte("permission denied")) || bytes.Contains([]byte(resBody), []byte(tempDir)) {
		t.Errorf("handleSave response leaked sensitive OS error details: %q", resBody)
	}
	if !bytes.Contains([]byte(resBody), []byte("Failed to save file")) && !bytes.Contains([]byte(resBody), []byte("Failed to create directory")) {
		t.Errorf("expected sanitized generic error message, got: %q", resBody)
	}
}

func TestHandleLogsSanitization(t *testing.T) {
	ch := make(chan string, 1)
	logChan = ch
	defer func() { logChan = nil }()

	rawLog := "First line\r\nSecond line [ADMIN] Fake Log entry"
	req, err := http.NewRequest("POST", "/logs", bytes.NewBufferString(rawLog))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handleLogs(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handleLogs returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	select {
	case msg := <-ch:
		if bytes.Contains([]byte(msg), []byte("\n")) || bytes.Contains([]byte(msg), []byte("\r")) {
			t.Errorf("handleLogs output contains unescaped newline or carriage return: %q", msg)
		}
	default:
		t.Error("expected log message in channel, got none")
	}
}

func TestRunShellAllowed(t *testing.T) {
	_, err := runShell(context.Background(), "ls")
	if err != nil {
		t.Errorf("expected ls to be allowed, got error: %v", err)
	}
}

func TestRunShellDisallowed(t *testing.T) {
	_, err := runShell(context.Background(), "rm -rf /")
	if err == nil {
		t.Errorf("expected rm to be disallowed, but it succeeded")
	}
}

func TestHandleCancelMethodValidation(t *testing.T) {
	req, err := http.NewRequest("GET", "/cancel", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handleCancel(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handleCancel returned wrong status code for GET: got %v want %v", status, http.StatusMethodNotAllowed)
	}

	postReq, err := http.NewRequest("POST", "/cancel", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	postRr := httptest.NewRecorder()
	handleCancel(postRr, postReq)

	if status := postRr.Code; status != http.StatusOK {
		t.Errorf("handleCancel returned wrong status code for POST: got %v want %v", status, http.StatusOK)
	}
}

const testPort = "8080"
const testToken = "test-session-token"

func newTestHandler() http.Handler {
	mux := http.NewServeMux()
	setupRoutes(mux)
	return newSecureHandler(testPort, testToken, mux)
}

func authedRequest(method, target string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.Host = "127.0.0.1:" + testPort
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: testToken})
	return req
}

func TestCSRFOriginValidation(t *testing.T) {
	h := newTestHandler()

	tests := []struct {
		name         string
		origin       string
		expectedCode int
	}{
		{"Allowed 127.0.0.1 same port", "http://127.0.0.1:8080", http.StatusOK},
		{"Allowed localhost same port", "http://localhost:8080", http.StatusOK},
		{"Blocked other localhost port", "http://localhost:3000", http.StatusForbidden},
		{"Blocked malicious domain", "http://attacker.com", http.StatusForbidden},
		{"Blocked spoofed localhost domain", "http://localhost.attacker.com", http.StatusForbidden},
		{"Blocked null origin", "null", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := authedRequest("POST", "/cancel")
			req.Header.Set("Origin", tt.origin)

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("origin %q: got %v want %v", tt.origin, rr.Code, tt.expectedCode)
			}
		})
	}
}

func TestHostHeaderValidation(t *testing.T) {
	h := newTestHandler()
	for _, host := range []string{"attacker.com:8080", "127.0.0.1:9999", "evil.localhost:8080"} {
		req := authedRequest("POST", "/cancel")
		req.Host = host
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Errorf("host %q: got %v want %v (DNS rebinding must be rejected)", host, rr.Code, http.StatusForbidden)
		}
	}
}

func TestSessionRequired(t *testing.T) {
	h := newTestHandler()
	for _, path := range []string{"/run", "/terminal", "/save", "/cancel", "/logs"} {
		req := httptest.NewRequest("POST", path, strings.NewReader("ls"))
		req.Host = "127.0.0.1:" + testPort
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("%s without session: got %v want 401", path, rr.Code)
		}

		req = httptest.NewRequest("POST", path, strings.NewReader("ls"))
		req.Host = "127.0.0.1:" + testPort
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "wrong"})
		rr = httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("%s with bad session: got %v want 401", path, rr.Code)
		}
	}
}

func TestTokenExchangeSetsCookie(t *testing.T) {
	h := newTestHandler()

	bad := httptest.NewRequest("GET", "/?token=nope", nil)
	bad.Host = "127.0.0.1:" + testPort
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, bad)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("bad token: got %v want 401", rr.Code)
	}

	good := httptest.NewRequest("GET", "/?token="+testToken, nil)
	good.Host = "127.0.0.1:" + testPort
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, good)
	if rr.Code != http.StatusSeeOther || rr.Header().Get("Location") != "/" {
		t.Fatalf("good token: got %v location %q, want 303 to /", rr.Code, rr.Header().Get("Location"))
	}
	var c *http.Cookie
	for _, ck := range rr.Result().Cookies() {
		if ck.Name == sessionCookieName {
			c = ck
		}
	}
	if c == nil || c.Value != testToken || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected HttpOnly SameSite=Strict session cookie, got %+v", c)
	}
}

func TestEmbeddedIndexServed(t *testing.T) {
	h := newTestHandler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, authedRequest("GET", "/"))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "<html") {
		t.Fatalf("index not served from embed: %v %q", rr.Code, rr.Body.String()[:min(80, rr.Body.Len())])
	}
	for _, p := range []string{"/app.js", "/style.css"} {
		rr = httptest.NewRecorder()
		h.ServeHTTP(rr, authedRequest("GET", p))
		if rr.Code != http.StatusOK {
			t.Errorf("%s: got %v want 200", p, rr.Code)
		}
	}
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, authedRequest("GET", "/style.scss"))
	if rr.Code != http.StatusNotFound {
		t.Errorf("unlisted asset should 404, got %v", rr.Code)
	}
}

func TestHandleLogsDoesNotBlock(t *testing.T) {
	ch := make(chan string) // unbuffered and never read
	logChan = ch
	defer func() { logChan = nil }()

	done := make(chan struct{})
	go func() {
		rr := httptest.NewRecorder()
		handleLogs(rr, httptest.NewRequest("POST", "/logs", strings.NewReader("x")))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleLogs blocked on an unread log channel")
	}
}

func TestHandleSavePathTraversal(t *testing.T) {

	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	tests := []struct {
		name           string
		filename       string
		content        string
		expectedStatus int
	}{
		{
			name:           "Valid local relative path",
			filename:       "valid_file.txt",
			content:        "hello world",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid nested relative path",
			filename:       "sub/valid_file.txt",
			content:        "nested content",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Path traversal attempt with parent dir",
			filename:       "../outside.txt",
			content:        "malicious",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Path traversal attempt deep",
			filename:       "../../outside.txt",
			content:        "malicious",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Absolute path attempt",
			filename:       "/etc/passwd",
			content:        "malicious",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty filename",
			filename:       "",
			content:        "empty",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"filename": tt.filename,
				"content":  tt.content,
			})
			req, err := http.NewRequest("POST", "/save", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handleSave(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}
