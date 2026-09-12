package web

import (
	"bytes"
	"encoding/json"
	"net"
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

func getFreePort(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on free port: %v", err)
	}
	defer l.Close()
	_, portStr, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		t.Fatalf("failed to split host port: %v", err)
	}
	return portStr
}

func TestStartServerAlreadyRunningSamePort(t *testing.T) {
	serverStarted = true
	serverPort = "8080"
	defer func() {
		serverStarted = false
		serverPort = ""
		logChan = nil
	}()

	err := StartServer("8080", nil)
	if err != nil {
		t.Errorf("expected nil error when calling StartServer with same port, got: %v", err)
	}
}

func TestStartServerAlreadyRunningDifferentPort(t *testing.T) {
	serverStarted = true
	serverPort = "8080"
	defer func() {
		serverStarted = false
		serverPort = ""
		logChan = nil
	}()

	err := StartServer("8081", nil)
	if err == nil {
		t.Errorf("expected error when calling StartServer with different port, got nil")
	}
}

func TestStartServerInvalidPort(t *testing.T) {
	serverStarted = false
	serverPort = ""
	defer func() {
		serverStarted = false
		serverPort = ""
		logChan = nil
	}()

	err := StartServer("invalid_port_999999", nil)
	if err == nil {
		t.Errorf("expected error for invalid port, got nil")
	}
	if serverStarted {
		t.Errorf("expected serverStarted to be false after ListenAndServe error")
	}
}

func TestStartServerSuccess(t *testing.T) {
	serverStarted = false
	serverPort = ""
	defer func() {
		serverStarted = false
		serverPort = ""
		logChan = nil
	}()

	port := getFreePort(t)
	logs := make(chan string, 10)

	go func() {
		_ = StartServer(port, logs)
	}()

	targetURL := "http://127.0.0.1:" + port + "/"
	var resp *http.Response
	var err error

	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		resp, err = http.Get(targetURL)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("failed to connect to server on port %s: %v", port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", resp.StatusCode)
	}

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range expectedHeaders {
		if got := resp.Header.Get(header); got != expectedValue {
			t.Errorf("header %s = %q, want %q", header, got, expectedValue)
		}
	}

	logMsg := "test log message"
	logReq, err := http.NewRequest("POST", "http://127.0.0.1:"+port+"/logs", bytes.NewBufferString(logMsg))
	if err != nil {
		t.Fatalf("failed to create log request: %v", err)
	}

	logResp, err := http.DefaultClient.Do(logReq)
	if err != nil {
		t.Fatalf("failed to send log request: %v", err)
	}
	logResp.Body.Close()

	select {
	case receivedLog := <-logs:
		if !strings.Contains(receivedLog, logMsg) {
			t.Errorf("expected log to contain %q, got %q", logMsg, receivedLog)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("timed out waiting for log on channel")
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

func TestHandleSavePathTraversal(t *testing.T) {
	// Isolate working directory
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
