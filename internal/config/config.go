package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const Version = "v1.0.0"

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

	return &config, nil
}

func SaveConfig(key string, value interface{}) error {
	viper.Set(key, value)
	return Write()
}

func Write() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".devcli.yaml")
	return viper.WriteConfigAs(configPath)
}

func Set(key string, value interface{}) {
	viper.Set(key, value)
}

func GetString(key string) string {
	return viper.GetString(key)
}
