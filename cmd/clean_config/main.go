package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func main() {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".devcli.yaml")

	viper.SetConfigFile(configPath)
	viper.ReadInConfig()
	viper.Set("ai_base_url", "")

	if err := viper.WriteConfig(); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
	} else {
		fmt.Println("Successfully cleared ai_base_url in .devcli.yaml")
	}
}
