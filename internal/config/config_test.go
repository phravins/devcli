package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig_NoHomeDir(t *testing.T) {
	// Save original env vars

	// Unset env vars to simulate missing home directory
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	// Reset viper to clear any cached states
	viper.Reset()

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error when home directory is missing, got nil")
	}
}

func TestLoadConfig_Success(t *testing.T) {
	// Create a temporary home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Create a dummy config file
	configContent := []byte(`
ai_backend: "openai"
ai_model: "gpt-4"
editor_theme: "dark"
user_name: "TestUser"
`)
	err := os.WriteFile(filepath.Join(tmpHome, ".devcli.yaml"), configContent, 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	viper.Reset()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.AIBackend != "openai" {
		t.Errorf("expected AIBackend 'openai', got '%s'", cfg.AIBackend)
	}
	if cfg.EditorTheme != "dark" {
		t.Errorf("expected EditorTheme 'dark', got '%s'", cfg.EditorTheme)
	}
	if cfg.UserName != "TestUser" {
		t.Errorf("expected UserName 'TestUser', got '%s'", cfg.UserName)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Create an invalid yaml file
	configContent := []byte(`
ai_backend: "openai"
	invalid_yaml: { [ :
`)
	err := os.WriteFile(filepath.Join(tmpHome, ".devcli.yaml"), configContent, 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	viper.Reset()

	_, err = LoadConfig()
	if err == nil {
		t.Fatal("expected error with invalid YAML config, got nil")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	viper.Reset()

	// Should not error if file is not found, should use defaults
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.EditorTheme != "default" {
		t.Errorf("expected EditorTheme 'default', got '%s'", cfg.EditorTheme)
	}
	if cfg.UserName != "Developer" {
		t.Errorf("expected UserName 'Developer', got '%s'", cfg.UserName)
	}
}

func TestLoadConfig_UnmarshalError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Create a yaml file with type mismatch to trigger Unmarshal error
	configContent := []byte(`
ai_backend:
  - "openai"
`)
	err := os.WriteFile(filepath.Join(tmpHome, ".devcli.yaml"), configContent, 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	viper.Reset()

	_, err = LoadConfig()
	if err == nil {
		t.Fatal("expected error with type mismatch YAML config, got nil")
	}
}

func TestCleanString(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"normal string", "normal string"},
		{"string with ]11; background color", "string with"},
		{"string with ]10; foreground color", "string with"},
		{"string with rgb: color", "string with"},
		{"string with \x1b escape sequence", "string with  escape sequence"},
	}

	for _, c := range cases {
		actual := CleanString(c.input)
		if actual != c.expected {
			t.Errorf("CleanString(%q) == %q, expected %q", c.input, actual, c.expected)
		}
	}
}

func TestSaveConfigAndWrite(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	viper.Reset()

	err := SaveConfig("test_key", "test_value")
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	val := GetString("test_key")
	if val != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", val)
	}

	// Test Set with non-string
	Set("test_int_key", 42)
	intVal := viper.GetInt("test_int_key")
	if intVal != 42 {
		t.Errorf("expected 42, got %d", intVal)
	}
}

func TestWrite_NoHomeDir(t *testing.T) {

	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	viper.Reset()

	err := Write()
	if err == nil {
		t.Fatal("expected error when home directory is missing, got nil")
	}
}

func TestWrite_ExistingFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	viper.Reset()

	// Write first time
	err := Write()
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Write second time (existing file path)
	err = Write()
	if err != nil {
		t.Fatalf("Write failed on existing file: %v", err)
	}
}
