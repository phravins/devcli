package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
func TestHandleRegister(t *testing.T) {
	// Isolate the test environment by resetting users and setting a temp HOME dir
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Reset the users map for a clean state
	users = make(map[string]User)

	tests := []struct {
		name           string
		body           map[string]interface{}
		setup          func()
		expectedStatus int
	}{
		{
			name: "Valid registration",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			setup:          func() {},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Duplicate registration",
			body: map[string]interface{}{
				"email":    "duplicate@example.com",
				"password": "password123",
			},
			setup: func() {
				// Pre-populate a user to simulate conflict
				users["duplicate@example.com"] = User{Email: "duplicate@example.com", Password: "hashedpassword"}
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Invalid JSON",
			// Invalid body type that causes json encode error or we can just send raw string below
			// We'll handle this specially in the test loop
			setup:          func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty email",
			body: map[string]interface{}{
				"email":    "   ",
				"password": "password123",
			},
			setup:          func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty password",
			body: map[string]interface{}{
				"email":    "user@example.com",
				"password": "   ",
			},
			setup:          func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Password exceeds length limit",
			body: map[string]interface{}{
				"email":    "longpass@example.com",
				"password": string(make([]byte, 80)),
			},
			setup:          func() {},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset users map for each test
			users = make(map[string]User)
			tt.setup()

			var reqBody []byte
			if tt.name == "Invalid JSON" {
				reqBody = []byte(`{"email": "test@example.com", "password":`)
			} else {
				reqBody, _ = json.Marshal(tt.body)
			}

			req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handleRegister(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}

func TestHandleDriveSave(t *testing.T) {
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
			filename:       "document.txt",
			content:        "Hello world",
			unauthenticated: true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid filename",
			filename:       "document.txt",
			content:        "Hello world",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid relative path",
			filename:       "folder/document.txt",
			content:        "Hello subfolder",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty filename",
			filename:       "",
			content:        "Empty file",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Path traversal parent directory",
			filename:       "../secret.txt",
			content:        "Secret content",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Absolute path traversal",
			filename:       "/etc/passwd",
			content:        "Malicious",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"filename": tt.filename,
				"content":  tt.content,
			})

			req, err := http.NewRequest("POST", "/drive/save", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			if !tt.unauthenticated {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-test-session"})
			}

			rr := httptest.NewRecorder()
			handleDriveSave(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handleDriveSave(%q) status = %v, want %v", tt.filename, status, tt.expectedStatus)
			}
		})
	}
}
