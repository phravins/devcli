package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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

func TestHandleRunAndTerminalAuth(t *testing.T) {
	// Test handleRun unauthenticated
	reqRun, _ := http.NewRequest("POST", "/run", bytes.NewBufferString(`print("hello")`))
	rrRun := httptest.NewRecorder()
	handleRun(rrRun, reqRun)
	if rrRun.Code != http.StatusUnauthorized {
		t.Errorf("handleRun unauthenticated got status %d, want %d", rrRun.Code, http.StatusUnauthorized)
	}

	// Test handleTerminal unauthenticated
	reqTerm, _ := http.NewRequest("POST", "/terminal", bytes.NewBufferString("ls"))
	rrTerm := httptest.NewRecorder()
	handleTerminal(rrTerm, reqTerm)
	if rrTerm.Code != http.StatusUnauthorized {
		t.Errorf("handleTerminal unauthenticated got status %d, want %d", rrTerm.Code, http.StatusUnauthorized)
	}
}

func TestHandleSavePathTraversal(t *testing.T) {
	// Isolate working directory
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	// Set up valid session
	authMu.Lock()
	sessions["valid-test-session"] = Session{
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	authMu.Unlock()
	defer func() {
		authMu.Lock()
		delete(sessions, "valid-test-session")
		authMu.Unlock()
	}()

	tests := []struct {
		name           string
		filename       string
		content        string
		unauthenticated bool
		expectedStatus int
	}{
		{
			name:           "Unauthenticated request rejected",
			filename:       "valid_file.txt",
			content:        "hello world",
			unauthenticated: true,
			expectedStatus: http.StatusUnauthorized,
		},
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
