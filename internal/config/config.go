package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const Version = "v1.1.0"

type Config struct {
	AIBackend     string            `mapstructure:"ai_backend"`
	AIModel       string            `mapstructure:"ai_model"`
	AIAPIKey      string            `mapstructure:"ai_api_key"`
	AIBaseURL     string            `mapstructure:"ai_base_url"`
	EditorTheme   string            `mapstructure:"editor_theme"`
	UserName      string            `mapstructure:"user_name"`
	HFAccessToken string            `mapstructure:"hf_access_token"`
	GeminiAPIKey  string            `mapstructure:"gemini_api_key"`
	Compilers     map[string]string `mapstructure:"compilers"` // Persisted detected paths
}

func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	viper.AddConfigPath(home)
	viper.SetConfigName(".devcli")
	viper.SetConfigType("yaml")

	viper.SetDefault("ai_backend", "")
	viper.SetDefault("editor_theme", "default")
	viper.SetDefault("user_name", "Developer")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// or create a default one
		} else {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	config.AIBackend = CleanString(config.AIBackend)
	config.AIModel = CleanString(config.AIModel)
	config.AIAPIKey = CleanString(config.AIAPIKey)
	config.AIBaseURL = CleanString(config.AIBaseURL)
	config.EditorTheme = CleanString(config.EditorTheme)
	config.UserName = CleanString(config.UserName)
	config.HFAccessToken = CleanString(config.HFAccessToken)
	config.GeminiAPIKey = CleanString(config.GeminiAPIKey)

	return &config, nil
}

func CleanString(s string) string {
	if s == "" {
		return ""
	}
	if idx := strings.Index(s, "]11;"); idx != -1 {
		s = s[:idx]
	}
	if idx := strings.Index(s, "]10;"); idx != -1 {
		s = s[:idx]
	}
	if idx := strings.Index(s, "rgb:"); idx != -1 {
		s = s[:idx]
	}
	var builder strings.Builder
	for _, r := range s {
		if r >= 32 && r != 127 && r != '\x1b' {
			builder.WriteRune(r)
		}
	}
	return strings.TrimSpace(builder.String())
}

func SaveConfig(key string, value interface{}) error {
	Set(key, value)
	return Write()
}

func Write() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	devcliDir := filepath.Join(home, ".devcli")
	_ = os.MkdirAll(devcliDir, 0700)
	_ = os.Chmod(devcliDir, 0700)

	configPath := filepath.Join(home, ".devcli.yaml")
	viper.SetConfigFile(configPath)
	var writeErr error
	if _, err := os.Stat(configPath); err == nil {
		writeErr = viper.WriteConfig()
	} else {
		writeErr = viper.WriteConfigAs(configPath)
	}

	if writeErr == nil {
		_ = os.Chmod(configPath, 0600)
	}
	return writeErr
}

func Set(key string, value interface{}) {
	if strVal, ok := value.(string); ok {
		viper.Set(key, CleanString(strVal))
	} else {
		viper.Set(key, value)
	}
}

func GetString(key string) string {
	return CleanString(viper.GetString(key))
}
