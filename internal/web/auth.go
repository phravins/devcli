package web

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"` // Hashed
}

type Session struct {
	Email     string
	ExpiresAt time.Time
}

var (
	users    = make(map[string]User)
	sessions = make(map[string]Session)
	authMu   sync.RWMutex
)

func init() {
	loadUsers()
}

func getUsersFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "web_users.json"
	}
	dir := filepath.Join(home, ".devcli")
	_ = os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "web_users.json")
}

func loadUsers() {
	path := getUsersFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &users)
}

func saveUsers() {
	path := getUsersFilePath()
	data, _ := json.MarshalIndent(users, "", "  ")
	if err := os.WriteFile(path, data, 0600); err == nil {
		_ = os.Chmod(path, 0600)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	authMu.Lock()
	defer authMu.Unlock()

	if _, exists := users[req.Email]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	users[req.Email] = User{Email: req.Email, Password: string(hashed)}
	saveUsers()

	msg := "New user registered: " + req.Email
	if logChan != nil {
		logChan <- msg
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User created")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	authMu.RLock()
	user, exists := users[req.Email]
	authMu.RUnlock()

	if !exists || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create session
	sessionID := generateSessionID()
	authMu.Lock()
	sessions[sessionID] = Session{
		Email:     req.Email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	authMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	msg := "User logged in: " + req.Email
	if logChan != nil {
		logChan <- msg
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"email": req.Email})
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func GetSessionUser(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", false
	}

	authMu.RLock()
	defer authMu.RUnlock()
	session, exists := sessions[cookie.Value]
	if !exists || session.ExpiresAt.Before(time.Now()) {
		return "", false
	}
	return session.Email, true
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		authMu.Lock()
		delete(sessions, cookie.Value)
		authMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

func handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	authMu.RLock()
	_, exists := users[req.Email]
	authMu.RUnlock()

	if !exists {
		// Don't leak exists status, just return OK
		w.WriteHeader(http.StatusOK)
		return
	}

	token := generateSessionID()
	fmt.Printf("\n[MOCK EMAIL] Password reset for %s: http://127.0.0.1:8080/reset-password?token=%s\n", req.Email, token)

	w.WriteHeader(http.StatusOK)
}

func handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	authMu.Lock()
	defer authMu.Unlock()

	// Find user by token (mock)
	fmt.Fprintf(w, "Email verified successfully! You can now login.")
}

func handleDriveSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Mock Drive Integration
	msg := fmt.Sprintf("Uploading file '%s' (%d bytes) to Google Drive...", req.Filename, len(req.Content))
	fmt.Printf("\n[MOCK DRIVE] %s\n", msg)
	
	if logChan != nil {
		logChan <- msg
		logChan <- "Drive Status: SUCCESS (File ID: mock_id_123)"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Saved to Google Drive!",
		"fileId":  "mock_id_123",
	})
}
