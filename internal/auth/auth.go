package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	BcryptCost	= 12
	SecurityVersion	= "1.0"
)

type AuthData struct {
	Username	string		`json:"username"`
	PasswordHash	string		`json:"password_hash"`
	CreatedAt	time.Time	`json:"created_at"`
	LastLogin	time.Time	`json:"last_login"`
	SecurityVersion	string		`json:"security_version"`
	RequireAuth	bool		`json:"require_auth"`
}

var (
	sessionUnlocked	bool
	sessionUser	string
	sessionMu	sync.RWMutex
)

func GetAuthDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	dir := filepath.Join(home, ".devcli")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create secure config directory: %w", err)
	}

	_ = os.Chmod(dir, 0700)
	return dir, nil
}

func GetAuthFilePath() (string, error) {
	dir, err := GetAuthDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "auth.json"), nil
}

func IsSetup() bool {
	path, err := GetAuthFilePath()
	if err != nil {
		return false
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	data, err := GetAuthData()
	if err != nil || data == nil {
		return false
	}

	return data.Username != "" && data.PasswordHash != ""
}

func GetAuthData() (*AuthData, error) {
	path, err := GetAuthFilePath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var data AuthData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse auth configuration: %w", err)
	}

	return &data, nil
}

func SaveAuthData(data *AuthData) error {
	path, err := GetAuthFilePath()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode auth data: %w", err)
	}

	if err := os.WriteFile(path, bytes, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	_ = os.Chmod(path, 0600)
	return nil
}

func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasUpper	bool
		hasLower	bool
		hasNumber	bool
	)

	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasNumber = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter (A-Z)")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter (a-z)")
	}
	if !hasNumber {
		return errors.New("password must contain at least one digit (0-9)")
	}

	return nil
}

func SetupUser(username, password string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	if err := ValidatePasswordStrength(password); err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	data := &AuthData{
		Username:		username,
		PasswordHash:		string(hashed),
		CreatedAt:		time.Now(),
		LastLogin:		time.Now(),
		SecurityVersion:	SecurityVersion,
		RequireAuth:		true,
	}

	if err := SaveAuthData(data); err != nil {
		return err
	}

	UnlockSession(username)
	return nil
}

func VerifyPassword(password string) (bool, error) {
	data, err := GetAuthData()
	if err != nil || data == nil {
		return false, errors.New("no account setup found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(data.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}

	data.LastLogin = time.Now()
	_ = SaveAuthData(data)

	UnlockSession(data.Username)
	return true, nil
}

func ChangePassword(currentPassword, newPassword string) error {
	valid, err := VerifyPassword(currentPassword)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("current password is incorrect")
	}

	if err := ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	data, err := GetAuthData()
	if err != nil || data == nil {
		return errors.New("failed to load user auth data")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	data.PasswordHash = string(hashed)
	return SaveAuthData(data)
}

func UnlockSession(username string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	sessionUnlocked = true
	sessionUser = username
}

func LockSession() {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	sessionUnlocked = false
	sessionUser = ""
}

func IsSessionUnlocked() bool {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return sessionUnlocked
}

func GetSessionUser() string {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return sessionUser
}
