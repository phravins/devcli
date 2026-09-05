package auth

import (
	"os"
	"testing"
)

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		valid       bool
		expectedErr string
	}{
		{
			name:        "Too short",
			password:    "short",
			valid:       false,
			expectedErr: "password must be at least 8 characters long",
		},
		{
			name:        "Exactly 8 characters valid",
			password:    "Aa345678",
			valid:       true,
		},
		{
			name:        "Missing uppercase",
			password:    "alllowercase123",
			valid:       false,
			expectedErr: "password must contain at least one uppercase letter (A-Z)",
		},
		{
			name:        "Missing lowercase",
			password:    "ALLUPPERCASE123",
			valid:       false,
			expectedErr: "password must contain at least one lowercase letter (a-z)",
		},
		{
			name:        "Missing number",
			password:    "NoNumbersHere!",
			valid:       false,
			expectedErr: "password must contain at least one digit (0-9)",
		},
		{
			name:        "Valid secure password",
			password:    "SecureP@ssw0rd2026",
			valid:       true,
		},
		{
			name:        "Another valid password",
			password:    "DevCLISecure!99",
			valid:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.valid {
				if err != nil {
					t.Errorf("expected password to be valid, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected password to be invalid, but validation passed")
				} else if err.Error() != tt.expectedErr {
					t.Errorf("expected error '%s', got '%v'", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestAuthWorkflow(t *testing.T) {
	// Create temporary home directory for testing auth storage
	tmpDir, err := os.MkdirTemp("", "devcli_auth_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	if IsSetup() {
		t.Errorf("expected IsSetup() to be false initially")
	}

	// Setup user
	err = SetupUser("testuser", "SecureP@ssw0rd2026")
	if err != nil {
		t.Fatalf("SetupUser failed: %v", err)
	}

	if !IsSetup() {
		t.Errorf("expected IsSetup() to be true after setup")
	}

	// Verify permissions
	path, err := GetAuthFilePath()
	if err != nil {
		t.Fatalf("GetAuthFilePath failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat auth file: %v", err)
	}

	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected auth file permissions to be 0600, got: %04o", perm)
	}

	// Verify password
	valid, err := VerifyPassword("SecureP@ssw0rd2026")
	if err != nil || !valid {
		t.Errorf("expected correct password verification to succeed")
	}

	valid, _ = VerifyPassword("WrongPassword123!")
	if valid {
		t.Errorf("expected incorrect password verification to fail")
	}

	// Change password
	err = ChangePassword("SecureP@ssw0rd2026", "NewSecur3P@ss!")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	valid, _ = VerifyPassword("NewSecur3P@ss!")
	if !valid {
		t.Errorf("expected verification with new password to succeed")
	}
}
