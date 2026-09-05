package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
