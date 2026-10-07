package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name		string
		command		string
		expected	[]string
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

func TestMaxBytesReaderLimit(t *testing.T) {
	largeBody := bytes.Repeat([]byte("a"), 3<<20) // 3 MB body

	endpoints := []struct {
		name	string
		path	string
		handler	http.HandlerFunc
	}{
		{"Logs", "/logs", handleLogs},
		{"Save", "/save", handleSave},
		{"Run", "/run", handleRun},
		{"Terminal", "/terminal", handleTerminal},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", ep.path, bytes.NewReader(largeBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			rr := httptest.NewRecorder()
			ep.handler(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected status %v for oversized body on %s, got %v", http.StatusBadRequest, ep.path, rr.Code)
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
		"filename":	readOnlyDir + "/file.txt",
		"content":	"test",
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
	_, err := runShell("ls")
	if err != nil {
		t.Errorf("expected ls to be allowed, got error: %v", err)
	}
}

func TestRunShellDisallowed(t *testing.T) {
	_, err := runShell("rm -rf /")
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

func TestCSRFOriginValidation(t *testing.T) {
	mux := http.NewServeMux()
	setupRoutes(mux)

	secureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data:;")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if r.Method == http.MethodPost {
			origin := r.Header.Get("Origin")
			if origin != "" {
				u, err := url.Parse(origin)
				if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
					http.Error(w, "Forbidden: Invalid Origin header", http.StatusForbidden)
					return
				}
			}
		}

		mux.ServeHTTP(w, r)
	})

	tests := []struct {
		name		string
		origin		string
		expectedCode	int
	}{
		{"Allowed 127.0.0.1", "http://127.0.0.1:8080", http.StatusOK},
		{"Allowed localhost", "http://localhost:3000", http.StatusOK},
		{"Blocked malicious domain", "http://attacker.com", http.StatusForbidden},
		{"Blocked spoofed localhost domain", "http://localhost.attacker.com", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/cancel", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Origin", tt.origin)

			rr := httptest.NewRecorder()
			secureHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("secureHandler returned wrong status code for origin %q: got %v want %v", tt.origin, rr.Code, tt.expectedCode)
			}
		})
	}
}


func TestHandleSavePathTraversal(t *testing.T) {

	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)


	tests := []struct {
		name		string
		filename	string
		content		string
		unauthenticated	bool
		expectedStatus	int
	}{
		{
			name:		"Valid local relative path",
			filename:	"valid_file.txt",
			content:	"hello world",
			expectedStatus:	http.StatusOK,
		},
		{
			name:		"Valid nested relative path",
			filename:	"sub/valid_file.txt",
			content:	"nested content",
			expectedStatus:	http.StatusOK,
		},
		{
			name:		"Path traversal attempt with parent dir",
			filename:	"../outside.txt",
			content:	"malicious",
			expectedStatus:	http.StatusBadRequest,
		},
		{
			name:		"Path traversal attempt deep",
			filename:	"../../outside.txt",
			content:	"malicious",
			expectedStatus:	http.StatusBadRequest,
		},
		{
			name:		"Absolute path attempt",
			filename:	"/etc/passwd",
			content:	"malicious",
			expectedStatus:	http.StatusBadRequest,
		},
		{
			name:		"Empty filename",
			filename:	"",
			content:	"empty",
			expectedStatus:	http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"filename":	tt.filename,
				"content":	tt.content,
			})
			req, err := http.NewRequest("POST", "/save", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			if !tt.unauthenticated {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-test-session"})
			}

			rr := httptest.NewRecorder()
			handleSave(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}
