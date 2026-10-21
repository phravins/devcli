package devtools

import (
	"testing"
)

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name        string
		port        string
		expectError bool
		expectedVal int
	}{
		{"valid port 8080", "8080", false, 8080},
		{"valid min port 1", "1", false, 1},
		{"valid max port 65535", "65535", false, 65535},
		{"invalid port 0", "0", true, 0},
		{"invalid port negative", "-80", true, 0},
		{"invalid port out of range", "70000", true, 0},
		{"invalid port non-numeric string", "abc", true, 0},
		{"flag injection attempt", "--bind=0.0.0.0", true, 0},
		{"command injection attempt", "8080; id", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := validatePort(tt.port)
			if tt.expectError {
				if err == nil {
					t.Errorf("validatePort(%q) expected error, got nil", tt.port)
				}
			} else {
				if err != nil {
					t.Errorf("validatePort(%q) unexpected error: %v", tt.port, err)
				}
				if val != tt.expectedVal {
					t.Errorf("validatePort(%q) = %d, want %d", tt.port, val, tt.expectedVal)
				}
			}
		})
	}
}
