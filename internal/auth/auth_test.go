package auth

import (
	"os"
	"testing"
)

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"short", false},
		{"alllowercase123", false},
		{"ALLUPPERCASE123", false},
		{"NoNumbersHere!", false},
		{"SecureP@ssw0rd2026", true},
		{"DevCLISecure!99", true},
	}

	for _, tt := range tests {
		err := ValidatePasswordStrength(tt.password)
		if tt.valid && err != nil {
			t.Errorf("expected password '%s' to be valid, got: %v", tt.password, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("expected password '%s' to be invalid, but validation passed", tt.password)
		}
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
